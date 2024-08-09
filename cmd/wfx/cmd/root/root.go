package root

/*
 * SPDX-FileCopyrightText: 2024 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"os/user"
	"sync"
	"syscall"
	"time"

	"github.com/Southclaws/fault"
	"github.com/coreos/go-systemd/v22/activation"
	"github.com/rs/zerolog/log"
	"github.com/siemens/wfx/cmd/wfx/cmd/config"
	"github.com/siemens/wfx/cmd/wfx/metadata"
	"github.com/siemens/wfx/internal/handler/job/events"
	"github.com/siemens/wfx/internal/man"
	"github.com/siemens/wfx/persistence"
	"github.com/spf13/cobra"
	"go.uber.org/automaxprocs/maxprocs"
	"golang.org/x/sync/errgroup"
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "wfx",
		Short: "wfx server",
		Long: `This API allows to create and execute reusable workflows for clients.
Each workflow is modeled as a state machine running in the storage, with tasks to be executed by clients.

Examples of tasks are installation of firmware or other types of commands issued to clients.`,
		Version: metadata.Version,
		RunE: func(cmd *cobra.Command, _ []string) error {
			rootCtx, rootCancel := context.WithCancel(cmd.Context())
			defer rootCancel()

			cfg, err := config.NewAppConfig(rootCtx, cmd.Flags())
			if err != nil {
				return fault.Wrap(err)
			}
			setupLogging(os.Stdout, cfg.LogFormat(), cfg.LogLevel())

			if _, err := maxprocs.Set(maxprocs.Logger(log.Printf)); err != nil {
				log.Warn().Err(err).Msg("Failed to set GOMAXPROCS")
			}

			var username string
			if u, err := user.Current(); err == nil {
				username = u.Username
			}

			log.Info().
				Str("version", metadata.Version).
				Str("date", metadata.Date).
				Str("commit", metadata.Commit).
				Str("user", username).
				Msg("Starting wfx")

			storage, err := initStorage(rootCtx, cfg)
			if err != nil {
				return fault.Wrap(err)
			}
			defer storage.Shutdown()

			// channel to catch signals
			chSignal := make(chan os.Signal, 1)
			signal.Notify(chSignal, os.Interrupt, syscall.SIGTERM)

			schemes := cfg.Schemes()

			// check for socket-based activation; order of sockets: south, north
			systemdListeners, _ := activation.Listeners()
			if len(systemdListeners) > 0 {
				// perform sanity checks
				if len(systemdListeners) != 2 {
					return errors.New("systemd socket-based activation requires two sockets")
				}
				if len(schemes) != 1 || schemes[0] != config.SchemeUnix {
					return errors.New("systemd socket-based activation only supports unix scheme")
				}
			}

			// for each scheme we create a north- and southbound server
			servers := make([]ServerCollection, 0, 2*len(schemes))
			var serverGroup errgroup.Group
			chPluginErrors := make(chan error, 1)
			for _, scheme := range schemes {
				collection, err := NewServerCollection(rootCtx, cfg, storage, chPluginErrors)
				if err != nil {
					return fault.Wrap(err)
				}
				servers = append(servers, *collection)

				var northListener, southListener net.Listener
				if len(systemdListeners) > 0 {
					log.Debug().Msg("Using sockets provided by systemd")
					southListener, northListener = systemdListeners[0], systemdListeners[1]
				} else {
					northSettings := ListenerSettings{
						Host:    cfg.MgmtHost(),
						Port:    cfg.MgmtPort(),
						TLSHost: cfg.MgmtTLSHost(),
						TLSPort: cfg.MgmtTLSPort(),
						UDSPath: cfg.MgmtUnixSocket(),
					}
					northListener, err = createListener(scheme, northSettings)
					if err != nil {
						return fault.Wrap(err)
					}

					southSettings := ListenerSettings{
						Host:    cfg.ClientHost(),
						Port:    cfg.ClientPort(),
						TLSHost: cfg.ClientTLSHost(),
						TLSPort: cfg.ClientTLSPort(),
						UDSPath: cfg.ClientUnixSocket(),
					}
					southListener, err = createListener(scheme, southSettings)
					if err != nil {
						return fault.Wrap(err)
					}
				}

				isTLS := scheme == config.SchemeHTTPS
				log.Info().
					Bool("tls", isTLS).
					Str("scheme", scheme.String()).
					Str("addr", northListener.Addr().String()).
					Msg("Starting northbound server")
				northServer := collection.North
				serverGroup.Go(func() error {
					var err error
					if isTLS {
						err = northServer.ServeTLS(northListener, cfg.TLSCertificate(), cfg.TLSKey())
					} else {
						err = northServer.Serve(northListener)
					}
					if err != nil && !errors.Is(err, http.ErrServerClosed) {
						log.Err(err).Msg("Northbound server encountered an error")
						return fault.Wrap(err)
					}
					return nil
				})

				log.Info().
					Bool("tls", isTLS).
					Str("scheme", scheme.String()).
					Str("addr", southListener.Addr().String()).
					Msg("Starting southbound server")
				southServer := collection.South
				serverGroup.Go(func() error {
					var err error
					if isTLS {
						err = southServer.ServeTLS(southListener, cfg.TLSCertificate(), cfg.TLSKey())
					} else {
						err = southServer.Serve(southListener)
					}
					if err != nil && !errors.Is(err, http.ErrServerClosed) {
						log.Err(err).Msg("Southbound server encountered an error")
						return fault.Wrap(err)
					}
					return nil
				})
			}

			var pluginErr error
			// wait for signal or an error
			running := true
			for running {
				select {
				case sig := <-chSignal:
					log.Debug().Str("signal", sig.String()).Msg("Caught signal")
					running = false
				case <-rootCtx.Done():
					log.Debug().Msg("Root Context done")
					running = false
				case pluginErr = <-chPluginErrors:
					log.Err(pluginErr).Msg("Received an error from a plugin")
					running = false
				}
			}

			// shut down (disconnect) subscribers otherwise we cannot stop the web server due to open connections
			events.ShutdownSubscribers()

			// create a context with a gracefulTimeout to allow outstanding requests to complete
			timeoutCtx, timeoutCancel := context.WithTimeout(rootCtx, cfg.GracefulTimeout())
			defer timeoutCancel()

			var shutdownGroup sync.WaitGroup
			for _, srv := range servers {
				shutdownGroup.Add(1)
				go func() {
					defer shutdownGroup.Done()
					_ = srv.North.Shutdown(timeoutCtx)
				}()
				shutdownGroup.Add(1)
				go func() {
					defer shutdownGroup.Done()
					_ = srv.South.Shutdown(timeoutCtx)
				}()
			}
			shutdownGroup.Wait()

			if err := serverGroup.Wait(); err != nil {
				return fault.Wrap(err)
			}
			if pluginErr != nil {
				return fault.Wrap(pluginErr)
			}
			return nil
		},
	}
	cmd.AddCommand(man.NewCommand())
	cmd.PersistentFlags().AddFlagSet(config.NewFlagset())
	_ = cmd.MarkPersistentFlagDirname(config.SimpleFileServerFlag)
	_ = cmd.MarkPersistentFlagFilename(config.ConfigFlag, "yml", "yaml")
	_ = cmd.MarkPersistentFlagDirname(config.ClientPluginsDirFlag)
	_ = cmd.MarkPersistentFlagDirname(config.MgmtPluginsDirFlag)
	return cmd
}

func initStorage(ctx context.Context, cfg *config.AppConfig) (persistence.Storage, error) {
	name, options := cfg.Storage(), cfg.StorageOptions()
	log.Debug().Str("name", name).Str("options", options).Msg("Setting up persistence storage")

	// note: storage is shared between north- and southbound API
	storage := persistence.GetStorage(name)
	if storage == nil {
		return nil, fmt.Errorf("unknown storage %s", name)
	}
	var err error
	for i := 0; i < 300; i++ {
		log.Debug().Str("name", name).Msg("Initializing storage")
		err = storage.Initialize(ctx, options)
		if err == nil {
			log.Info().Str("name", name).Msg("Initialized storage")
			break
		}
		dur := time.Second
		log.Warn().
			Err(err).
			Str("storage", name).
			Msg("Failed to initialize persistent storage. Trying again in one second...")
		time.Sleep(dur)
	}
	if err != nil {
		return nil, fault.Wrap(err)
	}
	return storage, nil
}
