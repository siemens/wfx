package main

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"github.com/rs/zerolog/log"
	_ "go.uber.org/automaxprocs"

	"github.com/siemens/wfx/cmd/wfx/cmd/root"
	"github.com/siemens/wfx/cmd/wfx/metadata"
)

func main() {
	root.Command.Version = metadata.Version
	if err := root.Command.Execute(); err != nil {
		log.Fatal().Err(err).Msg("wfx encountered an error")
	}
}
