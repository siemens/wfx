package main

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Southclaws/fault"
	"github.com/knadh/koanf/providers/env/v2"
	"github.com/knadh/koanf/providers/posflag"
	"github.com/knadh/koanf/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/siemens/wfx/cmd/wfx-loadtest/loadtest"
	"github.com/siemens/wfx/cmd/wfxctl/flags"
	"github.com/siemens/wfx/internal/cmd/man"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	k := koanf.New(".")
	cmd := cobra.Command{
		Use:     "wfx-loadtest",
		Short:   "Run a loadtest against wfx",
		Example: "wfx-loadtest --duration 10s",
		PreRun: func(cmd *cobra.Command, _ []string) {
			envProvider := env.Provider(".", env.Opt{
				Prefix: "WFX_",
				TransformFunc: func(k string, v string) (string, any) {
					// WFX_LOG_LEVEL becomes log-level
					return strings.ReplaceAll(strings.ToLower(strings.TrimPrefix(k, "WFX_")), "_", "-"), v
				},
			})
			if err := k.Load(envProvider, nil); err != nil {
				fmt.Fprintln(os.Stderr, "ERROR: Could not load env variables")
			}

			// --log-level becomes log.level
			if err := k.Load(posflag.Provider(cmd.Flags(), ".", k), nil); err != nil {
				log.Fatal().Err(err).Msg("Failed to load pflags")
			}

			log.Logger = zerolog.New(zerolog.ConsoleWriter{
				Out:        os.Stderr,
				TimeFormat: time.Stamp,
			}).With().Timestamp().Logger()
			if lvl, err := zerolog.ParseLevel(k.String(flags.LogLevelFlag)); err == nil {
				zerolog.SetGlobalLevel(lvl)
			}
		},
		RunE: func(*cobra.Command, []string) error {
			return fault.Wrap(loadtest.Run(k))
		},
	}
	cmd.AddCommand(man.NewCommand())
	f := cmd.PersistentFlags()

	f.String(loadtest.HostFlag, "localhost", "host")
	f.Int(loadtest.PortFlag, 8080, "port")
	f.String(loadtest.MgmtHostFlag, "localhost", "management host")
	f.Int(loadtest.MgmtPortFlag, 8081, "management port")

	f.String(flags.LogLevelFlag, "info", fmt.Sprintf("set log level. one of: %s,%s,%s,%s,%s,%s,%s",
		zerolog.TraceLevel.String(),
		zerolog.DebugLevel.String(),
		zerolog.InfoLevel.String(),
		zerolog.WarnLevel.String(),
		zerolog.ErrorLevel.String(),
		zerolog.FatalLevel.String(),
		zerolog.PanicLevel.String()))

	f.Int(loadtest.ReadFreqFlag, 75, "number of read (GET) requests per second")
	f.Int(loadtest.WriteFreqFlag, 25, "number of write (POST) requests per second")
	f.Duration(loadtest.DurationFlag, time.Minute, "how long the benchmark shall run")
	return &cmd
}
