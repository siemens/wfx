package cmd

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

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/posflag"
	mcobra "github.com/muesli/mango-cobra"
	"github.com/muesli/roff"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/siemens/wfx/cmd/wfx/metadata/config"
	"github.com/siemens/wfx/cmd/wfxctl/cmd/health"
	"github.com/siemens/wfx/cmd/wfxctl/cmd/job"
	"github.com/siemens/wfx/cmd/wfxctl/cmd/version"
	"github.com/siemens/wfx/cmd/wfxctl/cmd/workflow"
	"github.com/siemens/wfx/cmd/wfxctl/flags"
	"github.com/siemens/wfx/cmd/wfxctl/metadata"
	"github.com/spf13/cobra"
)

const (
	logLevelFlag = "log-level"
)

var RootCmd = &cobra.Command{
	Use:   "wfxctl",
	Short: "wfxctl is a command-line tool to interact with the wfx REST API",
	Long: `wfxctl can be used for management or diagnostic purposes.

To see raw HTTP requests and responses, export the environment variable DEBUG=true.

Tip: Shell completion is available for Bash, Fish and Zsh. See wfxctl completion --help for more information.
`,
	SilenceUsage:     true,
	TraverseChildren: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		f := cmd.Flags()
		if level, err := f.GetString(logLevelFlag); err == nil {
			if lvl, err := zerolog.ParseLevel(level); err == nil {
				zerolog.SetGlobalLevel(lvl)
			}
		}

		// Load the config files provided in the commandline.
		configFiles, _ := f.GetStringSlice(flags.ConfigFlag)
		log.Debug().Strs("configFiles", configFiles).Msg("Checking config files")
		for _, fname := range configFiles {
			if _, err := os.Stat(fname); err == nil {
				log.Debug().Str("fname", fname).Msg("Loading config file")
				prov := file.Provider(fname)
				if err := flags.Koanf.Load(prov, yaml.Parser()); err != nil {
					log.Fatal().Err(err).Msg("Failed to parse config file")
				}
			}
		}

		if err := flags.Koanf.Load(env.Provider("WFX_", ".", func(s string) string {
			result := strings.ReplaceAll(
				strings.ToLower(strings.TrimPrefix(s, "WFX_")), "_", "-")
			return result
		}), nil); err != nil {
			log.Err(err).Msg("Failed to env variables")
		}

		// --log-level becomes log.level
		if err := flags.Koanf.Load(posflag.Provider(cmd.Flags(), ".", flags.Koanf), nil); err != nil {
			log.Fatal().Err(err).Msg("Failed to load pflags")
		}

		log.Logger = zerolog.New(zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: time.Stamp,
		}).With().Timestamp().Logger()
		if lvl, err := zerolog.ParseLevel(flags.Koanf.String(logLevelFlag)); err == nil {
			zerolog.SetGlobalLevel(lvl)
		}

		log.Debug().
			Str("version", metadata.Version).
			Str("date", metadata.Date).
			Str("commit", metadata.Commit).
			Msg("wfxctl")

		return nil
	},
}

func init() {
	RootCmd.AddCommand(&cobra.Command{
		Use:   "man",
		Short: "Generate man page and exit",
		Run: func(cmd *cobra.Command, args []string) {
			manPage, err := mcobra.NewManPage(1, RootCmd)
			if err != nil {
				log.Fatal().Err(err).Msg("Failed to generate man page")
			}
			manPage = manPage.WithSection("Copyright", "(C) 2023 Siemens AG.\n"+
				"Licensed under the Apache License, Version 2.0")
			fmt.Fprintln(cmd.OutOrStdout(), manPage.Build(roff.NewDocument()))
		},
	})

	RootCmd.AddCommand(job.Command)
	RootCmd.AddCommand(workflow.Command)
	RootCmd.AddCommand(version.Command)
	RootCmd.AddCommand(health.Command)

	f := RootCmd.PersistentFlags()
	f.StringSlice(flags.ConfigFlag, config.DefaultConfigFiles(), "path to one or more .yaml config files; if this option is not set, then the default paths are tried")
	_ = RootCmd.MarkPersistentFlagFilename(flags.ConfigFlag, "yml", "yaml")

	f.String(flags.ClientHostFlag, "localhost", "host")
	f.Int(flags.ClientPortFlag, 8080, "port")
	f.String(flags.ClientTLSHostFlag, "localhost", "TLS host")
	f.Int(flags.ClientTLSPortFlag, 8443, "TLS port")
	f.String(flags.ClientUnixSocketFlag, "", "connect via the given unix-domain socket (if set, this overrides http/tls)")

	f.String(flags.MgmtHostFlag, "localhost", "management host")
	f.Int(flags.MgmtPortFlag, 8081, "management port")
	f.String(flags.MgmtTLSHostFlag, "localhost", "management TLS host")
	f.Int(flags.MgmtTLSPortFlag, 8444, "management TLS port")
	f.String(flags.MgmtUnixSocketFlag, "", "connect via the given unix-domain socket (if set, this overrides http/tls)")

	f.String(flags.TLSCaFlag, "/etc/ssl/cert.pem", "ca bundle (PEM)")
	f.Bool(flags.EnableTLSFlag, false, "whether to enable TLS (https)")

	f.String(flags.FilterFlag, "", "output filter (jq-expression). example: '.id'")
	f.Bool(flags.RawFlag, false, "output raw strings, not JSON texts; use --filter to select a single entity")

	f.String(logLevelFlag, "info", fmt.Sprintf("set log level. one of: %s,%s,%s,%s,%s,%s,%s",
		zerolog.TraceLevel.String(),
		zerolog.DebugLevel.String(),
		zerolog.InfoLevel.String(),
		zerolog.WarnLevel.String(),
		zerolog.ErrorLevel.String(),
		zerolog.FatalLevel.String(),
		zerolog.PanicLevel.String()))
}
