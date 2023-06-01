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
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/siemens/wfx/cmd/wfx/metadata"
	"github.com/spf13/cobra"
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
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"yaml", "yml", "json"}, cobra.ShellCompDirectiveFilterFileExt
	},
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		log.Logger = zerolog.New(zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: time.Stamp,
		}).With().Timestamp().Logger()
	},
	Run: func(cmd *cobra.Command, args []string) {
		f := cmd.PersistentFlags()

		input := args[0]
		output, err := f.GetString(outputFlag)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to get outputFlag")
		}
		outputFormat, err := f.GetString(outputFormatFlag)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to get outputFormat")
		}
		krokiURL, err := f.GetString(krokiURLFlag)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to get krokiURL")
		}

		var inFile, outFile *os.File
		defer func() {
			if inFile != nil {
				inFile.Close()
			}
			if outFile != nil {
				outFile.Close()
			}
		}()
		inFile, err = os.OpenFile(input, os.O_RDONLY, 0o644)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to open input file")
		}
		cmd.SetIn(bufio.NewReader(inFile))
		if output != "" {
			var err error
			outFile, err = os.OpenFile(output, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o644)
			if err != nil {
				log.Fatal().Err(err).Msg("Failed to open output file")
			}
			cmd.SetOut(outFile)
		}

		workflow := readInput(cmd.InOrStdin())
		outWriter := bufio.NewWriter(cmd.OutOrStdout())

		outputFormat = strings.ToLower(outputFormat)
		switch outputFormat {
		case "plantuml":
			generatePlantUML(outWriter, workflow)
		case "svg":
			if err := generateSvg(outWriter, krokiURL, workflow); err != nil {
				log.Fatal().Err(err).Msg("Failed to generate svg")
			}
		default:
			log.Fatal().Str("outputFormat", outputFormat).Msg("Unsupported format")
		}

		outWriter.Flush()
	},
}

func main() {
	rootCmd.Version = metadata.Version
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
