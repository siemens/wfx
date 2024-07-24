package root

/*
 * SPDX-FileCopyrightText: 2024 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"fmt"
	"os"
	"os/signal"
	"os/user"
	"sync"
	"syscall"
	"time"

	"github.com/Southclaws/fault"
	"github.com/rs/zerolog/log"
	"github.com/siemens/wfx/cmd/wfx/cmd/config"
	"github.com/siemens/wfx/cmd/wfx/metadata"
	"github.com/siemens/wfx/internal/cmd/man"
	"github.com/siemens/wfx/persistence"
	"github.com/spf13/cobra"
	"go.uber.org/automaxprocs/maxprocs"
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "wfx",
		Short: "wfx server",
		Long: `wfx is a server that models workflows as finite-state machines and exposes a REST API to update the state machine.
It drives tasks in coordination with clients through jobs, with each job instantiation including metadata to guide the client.

Examples of tasks are installation of firmware or other types of commands issued to clients.`,
		Version:      metadata.Version,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := config.NewAppConfig(cmd.Flags())
			if err != nil {
				return fault.Wrap(err)
			}
			defer cfg.Stop()

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

			storage, err := initStorage(cfg)
			if err != nil {
				return fault.Wrap(err)
			}
			defer storage.Shutdown()

			// channel to catch signals
			chSignal := make(chan os.Signal, 1)
			signal.Notify(chSignal, os.Interrupt, syscall.SIGTERM)

			chErr := make(chan error, 1)
			collection, err := NewServerCollection(cfg, storage)
			if err != nil {
				return fault.Wrap(err)
			}

			var g sync.WaitGroup
			g.Add(1)
			go func() {
				defer g.Done()
				err := collection.Start()
				log.Debug().Msg("Server collection done")
				if err != nil {
					log.Debug().Msg("Sending error")
					chErr <- err
				}
				close(chErr)
			}()

			// reset error variable
			err = nil
			// wait for any shutdown event
			select {
			case sig := <-chSignal:
				log.Info().Str("signal", sig.String()).Msg("Caught signal")
			case err = <-chErr:
				log.Err(err).Msg("Error in server collection")
			case <-cmd.Context().Done():
				log.Info().Msg("Context done")
			}

			collection.Stop()
			log.Trace().Msg("Waiting for server collection goroutine")
			g.Wait()

			log.Info().Msg("Exiting")
			return fault.Wrap(err)
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

func initStorage(cfg *config.AppConfig) (persistence.Storage, error) {
	name, options := cfg.Storage(), cfg.StorageOptions()
	log.Debug().Str("name", name).Str("options", options).Msg("Setting up persistent storage")

	// note: storage is shared between north- and southbound API
	storage := persistence.GetStorage(name)
	if storage == nil {
		return nil, fmt.Errorf("unknown storage %s", name)
	}
	var err error
	for i := 0; i < 300; i++ {
		log.Debug().Str("name", name).Msg("Initializing storage")
		err = storage.Initialize(options)
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
