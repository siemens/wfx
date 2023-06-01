package entgo

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"github.com/rs/zerolog/log"
	"github.com/siemens/wfx/generated/ent"
)

// Database manages access to a SQLite database.
// It implements the wfx persistence interface.
// It holds a pointer to the connection and is safe to copy by value.
type Database struct {
	client *ent.Client
}

func (db Database) Shutdown() {
	if err := db.client.Close(); err != nil {
		log.Error().Err(err).Msg("Error closing database connection")
	}
	log.Info().Msg("Closed database connection")
}
