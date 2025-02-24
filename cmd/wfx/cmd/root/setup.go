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

	"github.com/coreos/go-systemd/v22/journal"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/journald"
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
			if ok, _ := journal.StderrIsJournalStream(); ok {
				format = "journald"
			} else {
				format = "json"
			}
		}
	}
	switch format {
	case "json":
		logger = zerolog.New(out)
	case "pretty":
		logger = zerolog.New(zerolog.ConsoleWriter{
			Out:        out,
			TimeFormat: time.Stamp,
		})
	case "journald":
		logger = zerolog.New(journald.NewJournalDWriter())
	default:
		fmt.Fprintf(os.Stderr, "Invalid log format specified: %s\n", format)
	}
	log.Logger = logger.With().Timestamp().Caller().Logger()

	log.Debug().Str("format", format).Msg("Logging configured successfully")
}
