package main

/*
 * SPDX-FileCopyrightText: 2024 Siemens AG
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
	"sort"
	"strings"
	"time"

	"github.com/Southclaws/fault"
	"github.com/goccy/go-yaml"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/siemens/wfx/cmd/wfx-viewer/output"
	"github.com/siemens/wfx/cmd/wfx/metadata"
	"github.com/siemens/wfx/generated/api"
	"github.com/siemens/wfx/internal/cmd/man"
	"github.com/spf13/cobra"
)

const (
	outputFlag       = "output"
	outputFormatFlag = "output-format"
)

func init() {
	rootCmd.Version = metadata.Version
	rootCmd.AddCommand(man.NewCommand())
	f := rootCmd.PersistentFlags()
	f.String("log-level", "info", fmt.Sprintf("set log level. one of: %s,%s,%s,%s,%s,%s,%s",
		zerolog.TraceLevel.String(),
		zerolog.DebugLevel.String(),
		zerolog.InfoLevel.String(),
		zerolog.WarnLevel.String(),
		zerolog.ErrorLevel.String(),
		zerolog.FatalLevel.String(),
		zerolog.PanicLevel.String()))

	f.String(outputFlag, "", "output file (default: stdout)")

	allFormats := make([]string, 0, len(output.Generators))
	for format, gen := range output.Generators {
		allFormats = append(allFormats, format)
		gen.RegisterFlags(f)
	}
	sort.Strings(allFormats)
	f.String(outputFormatFlag, allFormats[0], fmt.Sprintf("output format. possible values: %s", strings.Join(allFormats, ",")))
}

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
		return []string{"yaml", "yml"}, cobra.ShellCompDirectiveFilterFileExt
	},
	PersistentPreRun: func(cmd *cobra.Command, _ []string) {
		writer := zerolog.ConsoleWriter{Out: cmd.ErrOrStderr(), TimeFormat: time.Stamp}
		log.Logger = zerolog.New(writer).With().Timestamp().Logger()
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
				_ = inFile.Close()
			}
			if outFile != nil {
				_ = outFile.Close()
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

		var workflow api.Workflow
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
