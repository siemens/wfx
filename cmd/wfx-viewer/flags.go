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

	mcobra "github.com/muesli/mango-cobra"
	"github.com/muesli/roff"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

const (
	outputFlag       = "output"
	outputFormatFlag = "output-format"
	krokiURLFlag     = "kroki-url"
)

func init() {
	rootCmd.AddCommand(&cobra.Command{
		Use:   "man",
		Short: "Generate man page and exit",
		Run: func(cmd *cobra.Command, args []string) {
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
	f.String(outputFlag, "", "output file (default: stdout)")
	f.String(outputFormatFlag, "plantuml", "output format. possible values: plantuml, svg")
	f.String(krokiURLFlag, "https://kroki.io/plantuml/svg", "url to kroki (used for svg)")
}
