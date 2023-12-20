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
	"github.com/siemens/wfx/internal/handler/job/events"
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

		var schemes []string
		k.Read(func(k *koanf.Koanf) {
			schemes = k.Strings(schemeFlag)
		})
		serverCollections := make([]*serverCollection, 0, 3)
		chQuit := make(chan error)
		{
			collection, err := createNorthboundCollection(schemes, storage, chQuit)
			if err != nil {
				return fault.Wrap(err)
			}
			serverCollections = append(serverCollections, collection)
		}
		{
			collection, err := createSouthboundCollection(schemes, storage, chQuit)
			if err != nil {
				return fault.Wrap(err)
			}
			serverCollections = append(serverCollections, collection)
		}

		// check for socket-based activation (systemd)
		listeners, _ := activation.Listeners()
		serverCollections = append(serverCollections, adoptListeners(listeners, storage, chQuit)...)

		for _, collection := range serverCollections {
			for i := range collection.servers {
				// capture loop variables
				// see https://go.dev/blog/loopvar-preview
				// TODO: remove this once our go.mod targets Go 1.22
				name := collection.name
				srv := collection.servers[i]
				go func() {
					if err := launchServer(name, srv); err != nil {
						chQuit <- err
					}
				}()
			}
		}

		// wait for signal or an error
		running := true
		for running {
			select {
			case <-signalChannel:
				running = false
			case <-chQuit:
				running = false
			}
		}

		// shut down (disconnect) subscribers otherwise we cannot stop the web server due to open connections
		events.ShutdownSubscribers()

		// create a context with a timeout to allow outstanding requests to complete
		var timeout time.Duration
		k.Read(func(k *koanf.Koanf) {
			timeout = k.Duration(gracefulTimeoutFlag)
		})
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		for _, collection := range serverCollections {
			collection.Shutdown(ctx)
		}
		return nil
	},
}

func adoptListeners(listeners []net.Listener, storage persistence.Storage, chQuit chan error) []*serverCollection {
	if len(listeners) == 2 {
		log.Debug().Msg("Adopting sockets provided by systemd")
		south, err := createSouthboundCollection([]string{kindHTTP.String()}, storage, chQuit)
		if err != nil {
			log.Err(err).Msg("Failed to create southbound collection")
			return nil
		}
		south.servers[0].Listener = func() (net.Listener, error) {
			// use listener created by systemd
			return listeners[0], nil
		}

		north, err := createNorthboundCollection([]string{kindHTTP.String()}, storage, chQuit)
		if err != nil {
			log.Err(err).Msg("Failed to create northbound collection")
			return nil
		}
		north.servers[0].Listener = func() (net.Listener, error) {
			// use listener created by systemd
			return listeners[1], nil
		}
		return []*serverCollection{south, north}
	}
	log.Debug().Msg("No sockets provided by systemd")
	return nil
}

func launchServer(name string, srv myServer) error {
	ln, err := srv.Listener()
	if err != nil {
		return fault.Wrap(err)
	}
	log.Info().Str("name", name).Str("addr", ln.Addr().String()).Str("kind", srv.Kind.String()).Msg("Starting server")
	if srv.Kind == kindHTTPS {
		err = srv.Srv.ServeTLS(ln, "", "")
	} else {
		err = srv.Srv.Serve(ln)
	}
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fault.Wrap(err)
	}
	return nil
}
