package main

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/Southclaws/fault"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/siemens/wfx/cmd/wfx-viewer/output"
	"github.com/siemens/wfx/cmd/wfx/metadata"
	"github.com/siemens/wfx/generated/model"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var rootCmd = &cobra.Command{
	Use:   "wfx-viewer",
	Short: "Visualize workflows.",
	Long: `Visualize workflows.

Note: svg generation sends your workflow to a remote Kroki server.
Do not use this for confidential information.
`,
	Example: "wfx-viewer --output-format svg --output wfx.workflow.dau.direct.svg wfx.workflow.dau.direct.yml",
	Args:    cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	ValidArgsFunction: func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
		return []string{"yaml", "yml", "json"}, cobra.ShellCompDirectiveFilterFileExt
	},
	PersistentPreRun: func(cmd *cobra.Command, _ []string) {
		log.Logger = zerolog.New(zerolog.ConsoleWriter{
			Out:        cmd.ErrOrStderr(),
			TimeFormat: time.Stamp,
		}).With().Timestamp().Logger()
		logLevel, _ := cmd.PersistentFlags().GetString("log-level")
		if lvl, err := zerolog.ParseLevel(logLevel); err == nil {
			zerolog.SetGlobalLevel(lvl)
		}
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		f := cmd.PersistentFlags()

		src := args[0]
		dest, err := f.GetString(outputFlag)
		if err != nil {
			return fault.Wrap(err)
		}
		format, err := f.GetString(outputFormatFlag)
		if err != nil {
			return fault.Wrap(err)
		}
		log.Debug().Str("src", src).Str("dest", dest).Str("format", format).Msg("Starting conversion")

		var inFile, outFile *os.File
		defer func() {
			if inFile != nil {
				inFile.Close()
			}
			if outFile != nil {
				outFile.Close()
			}
		}()
		inFile, err = os.OpenFile(src, os.O_RDONLY, 0o644)
		if err != nil {
			return fault.Wrap(err)
		}
		cmd.SetIn(bufio.NewReader(inFile))
		if dest != "" {
			var err error
			outFile, err = os.OpenFile(dest, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o644)
			if err != nil {
				return fault.Wrap(err)
			}
			cmd.SetOut(outFile)
		}

		var workflow model.Workflow
		{ // parse workflow
			b, err := io.ReadAll(cmd.InOrStdin())
			if err != nil {
				return fault.Wrap(err)
			}
			if err = yaml.Unmarshal(b, &workflow); err != nil {
				return fault.Wrap(err)
			}
		}
		log.Debug().Str("name", workflow.Name).Msg("Workflow parsed")

		outWriter := bufio.NewWriter(cmd.OutOrStdout())

		format = strings.ToLower(format)
		gen, ok := output.Generators[format]
		if !ok {
			return fmt.Errorf("unsupported output format: %s", format)
		}
		log.Debug().Msg("Generating output")
		if err := gen.Generate(outWriter, &workflow); err != nil {
			return fault.Wrap(err)
		}
		_ = outWriter.Flush()
		log.Debug().Msg("Successfully generated output")
		return nil
	},
}

func main() {
	rootCmd.Version = metadata.Version
	if err := rootCmd.Execute(); err != nil {
		log.Fatal().Err(err).Msg("wfx-viewer encountered an error")
	}
}
