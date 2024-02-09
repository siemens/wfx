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
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/posflag"
	"github.com/knadh/koanf/v2"
	mcobra "github.com/muesli/mango-cobra"
	"github.com/muesli/roff"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/siemens/wfx/cmd/wfx-loadtest/loadtest"
	"github.com/spf13/cobra"
)

const (
	logLevelFlag = "log-level"
)

func init() {
	rootCmd.AddCommand(&cobra.Command{
		Use:   "man",
		Short: "Generate man page and exit",
		Run: func(*cobra.Command, []string) {
			manPage, err := mcobra.NewManPage(1, rootCmd)
			if err != nil {
				log.Fatal().Err(err).Msg("Failed to generate man page")
			}
			manPage = manPage.WithSection("Copyright", "(C) 2023 Siemens AG.\n"+
				"Licensed under the Apache License, Version 2.0")
			fmt.Println(manPage.Build(roff.NewDocument()))
		},
	})

	f := rootCmd.PersistentFlags()

	f.String(loadtest.HostFlag, "localhost", "host")
	f.Int(loadtest.PortFlag, 8080, "port")
	f.String(loadtest.MgmtHostFlag, "localhost", "management host")
	f.Int(loadtest.MgmtPortFlag, 8081, "management port")

	f.String(logLevelFlag, "info", fmt.Sprintf("set log level. one of: %s,%s,%s,%s,%s,%s,%s",
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
}

var k = koanf.New(".")

var rootCmd = &cobra.Command{
	Use:     "wfx-loadtest",
	Short:   "Run a loadtest against wfx",
	Example: "wfx-loadtest --duration 10s",
	PreRun: func(cmd *cobra.Command, _ []string) {
		if err := k.Load(env.Provider("WFX_", ".", func(s string) string {
			result := strings.ReplaceAll(
				strings.ToLower(strings.TrimPrefix(s, "WFX_")), "_", "-")
			return result
		}), nil); err != nil {
			log.Err(err).Msg("Failed to env variables")
		}

		// --log-level becomes log.level
		if err := k.Load(posflag.Provider(cmd.Flags(), ".", k), nil); err != nil {
			log.Fatal().Err(err).Msg("Failed to load pflags")
		}

		log.Logger = zerolog.New(zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: time.Stamp,
		}).With().Timestamp().Logger()
		if lvl, err := zerolog.ParseLevel(k.String(logLevelFlag)); err == nil {
			zerolog.SetGlobalLevel(lvl)
		}
	},
	RunE: func(*cobra.Command, []string) error {
		return fault.Wrap(loadtest.Run(k))
	},
}
