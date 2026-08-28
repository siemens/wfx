package main

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"os"

	"github.com/Southclaws/fault"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/siemens/wfx/cmd/wfx/cmd/root"
)

func configureErrorStackMarshaler() {
	zerolog.ErrorStackMarshaler = func(err error) any { //nolint:reassign // zerolog's documented global configuration hook
		return fault.Flatten(err)
	}
}

func main() {
	configureErrorStackMarshaler()
	if err := root.NewCommand().Execute(); err != nil {
		log.Error().Stack().Err(err).Msg("wfx encountered an error")
		os.Exit(1)
	}
}
