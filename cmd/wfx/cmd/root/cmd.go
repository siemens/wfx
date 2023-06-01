package root

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
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
	"strings"
	"syscall"
	"time"

	"github.com/Southclaws/fault"
	"github.com/coreos/go-systemd/v22/activation"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/posflag"
	"github.com/knadh/koanf/v2"
	"github.com/rs/zerolog/log"
	"github.com/siemens/wfx/cmd/wfx/metadata"
	"github.com/siemens/wfx/internal/config"
	"github.com/siemens/wfx/persistence"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var k = config.New()

var signalChannel = make(chan os.Signal, 1)

var Command = &cobra.Command{
	Use:   "wfx",
	Short: "wfx server",
	Long: `This API allows to create and execute reusable workflows for clients.
Each workflow is modeled as a state machine running in the storage, with tasks to be executed by clients.

Examples of tasks are installation of firmware or other types of commands issued to clients.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		f := cmd.Flags()
		knownOptions := make(map[string]bool, 64)
		f.VisitAll(func(flag *pflag.Flag) {
			knownOptions[flag.Name] = true
		})

		mergeFn := koanf.WithMergeFunc(func(src, dest map[string]any) error {
			// merge src into dest
			for k, v := range src {
				if _, exists := knownOptions[k]; !exists {
					fmt.Fprintf(os.Stderr, "ERROR: Unknown config option '%s'", k)
				}
				dest[k] = v
			}
			return nil
		})

		// Load the config files provided in the commandline.
		cFiles, _ := f.GetStringSlice(configFlag)
		var fileProvider *file.File
		for _, fname := range cFiles {
			if _, err := os.Stat(fname); err == nil {
				fileProvider = file.Provider(fname)
				k.Write(func(k *koanf.Koanf) {
					if err := k.Load(fileProvider, yaml.Parser(), mergeFn); err != nil {
						fmt.Fprintf(os.Stderr, "ERROR: Failed to load config file '%s'", fname)
					}
				})

			}
		}

		envProvider := env.Provider("WFX_", ".", func(s string) string {
			// WFX_LOG_LEVEL becomes log-level
			return strings.ReplaceAll(strings.ToLower(strings.TrimPrefix(s, "WFX_")), "_", "-")
		})
		k.Write(func(k *koanf.Koanf) {
			if err := k.Load(envProvider, nil, mergeFn); err != nil {
				fmt.Fprintln(os.Stderr, "ERROR: Could not load env variables")
			}
			if err := k.Load(posflag.Provider(f, ".", k), nil); err != nil {
				fmt.Fprintln(os.Stderr, "ERROR: Could not load CLI flags")
			}
		})

		// now that we have merged all config sources, set up logger
		var logLevel, logFormat string
		k.Read(func(k *koanf.Koanf) {
			logLevel = k.String(logLevelFlag)
			logFormat = k.String(logFormatFlag)
		})
		setupLogging(os.Stdout, logFormat, logLevel)

		// start watching config
		if fileProvider != nil {
			err := fileProvider.Watch(func(event interface{}, err error) {
				if err == nil {
					k.Write(func(k *koanf.Koanf) {
						if err := k.Load(fileProvider, yaml.Parser(), mergeFn); err == nil {
							if err := reloadConfig(k); err != nil {
								log.Error().Err(err).Msg("Failed to reload config")
							}
						}
					})
				}
			})
			if err != nil {
				log.Error().Err(err).Msg("Failed to set up config file watcher")
			}
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
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

		var name, options string
		k.Read(func(k *koanf.Koanf) {
			name = k.String(storageFlag)
			options = k.String(storageOptFlag)
		})
		log.Debug().Str("name", name).Str("options", options).Msg("Setting up persistence storage")
		if name != preferedStorage && options == defaultStorageOpts {
			options = ""
		}

		// note: storage is shared between north- and southbound API
		storage := persistence.GetStorage(name)
		if storage == nil {
			return fmt.Errorf("unknown storage %s", name)
		}
		var err error
		for i := 0; i < 300; i++ {
			log.Debug().Str("name", name).Msg("Initializing storage")
			err = storage.Initialize(context.Background(), options)
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
			return fault.Wrap(err)
		}
		defer storage.Shutdown()

		signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)

		allServers := make([]myServer, 0, 2)
		{
			servers, err := createNorthboundServers(storage)
			if err != nil {
				return fault.Wrap(err)
			}
			allServers = append(allServers, servers...)
		}
		{
			servers, err := createSouthboundServers(storage)
			if err != nil {
				return fault.Wrap(err)
			}
			allServers = append(allServers, servers...)
		}

		errChan := make(chan error)
		for i := range allServers {
			// make copy of loop variable
			idx := i
			go func() {
				maxAttempts := 30
				for attempt := 1; attempt <= maxAttempts; attempt++ {
					srv := allServers[idx]
					log.Info().Str("addr", srv.Srv.Addr).Str("kind", srv.Kind.String()).Msg("Starting server")
					var err error
					switch srv.Kind {
					case kindHTTP:
						err = srv.Srv.ListenAndServe()
					case kindHTTPS:
						err = srv.Srv.ListenAndServeTLS("", "")
					case kindUnix:
						var l net.Listener
						l, err = net.Listen("unix", srv.Srv.Addr)
						if err != nil {
							log.Err(err).Int("idx", idx).Msg("Failed to launch unix-domain socket server")
							errChan <- err
							return
						}
						err = srv.Srv.Serve(l)
					}
					if err == nil || errors.Is(err, http.ErrServerClosed) {
						break
					}
					if err != nil {
						log.Err(err).
							Int("idx", idx).
							Int("attempt", attempt).
							Int("kind", int(srv.Kind)).
							Str("addr", srv.Srv.Addr).
							Msg("Failed to start server")
						if attempt >= maxAttempts {
							errChan <- err
							return
						}
						time.Sleep(time.Second)
					}
				}
			}()
		}

		// check for systemd socket-based activation
		listeners, _ := activation.Listeners()
		err = adoptSystemdSockets(listeners, storage, errChan)
		if err != nil {
			log.Error().Err(err).Msg("Failed to adopt systemd sockets")
		}

		// wait for signal or an error
		running := true
		for running {
			select {
			case <-signalChannel:
				running = false
			case <-errChan:
				running = false
			}
		}

		// Create a context with a timeout to allow outstanding requests to complete
		var timeout time.Duration
		k.Read(func(k *koanf.Koanf) {
			timeout = k.Duration(gracefulTimeoutFlag)
		})
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		for i, s := range allServers {
			log.Info().Int("i", i).Msg("Shutting down server...")
			if err := s.Srv.Shutdown(ctx); err != nil {
				log.Err(err).Msg("Shutdown error")
			}
			log.Debug().Int("i", i).Msg("Shut down successful")
		}

		return nil
	},
}
