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
	"github.com/siemens/wfx/cmd/wfxctl/cmd"
	"github.com/siemens/wfx/cmd/wfxctl/metadata"
)

func main() {
	cmd.RootCmd.Version = metadata.Version
	if err := cmd.RootCmd.Execute(); err != nil {
		log.Fatal().Err(err).Msg("wfxctl encountered an error")
	}
}
