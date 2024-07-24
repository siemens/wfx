package root

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
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/term"
)

func setupLogging(out *os.File, format string, lvl zerolog.Level) {
	zerolog.SetGlobalLevel(lvl)

	var logger zerolog.Logger
	if format == "auto" {
		if term.IsTerminal(int(out.Fd())) {
			format = "pretty"
		} else {
			format = "json"
		}
	}
	switch format {
	case "json":
		logger = zerolog.New(out).With().Caller().Logger()
	case "pretty":
		logger = zerolog.New(zerolog.ConsoleWriter{
			Out:        out,
			TimeFormat: time.Stamp,
		})
	default:
		fmt.Fprintf(os.Stderr, "Invalid log format: %s\n", format)
	}
	log.Logger = logger.With().Timestamp().Caller().Logger()
}
