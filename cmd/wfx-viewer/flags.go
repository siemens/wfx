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
	"sort"
	"strings"

	mcobra "github.com/muesli/mango-cobra"
	"github.com/muesli/roff"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/siemens/wfx/cmd/wfx-viewer/output"
	"github.com/spf13/cobra"
)

const (
	outputFlag       = "output"
	outputFormatFlag = "output-format"
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
