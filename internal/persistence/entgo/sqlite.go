//go:build sqlite

package entgo

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"context"
	"embed"

	"github.com/Southclaws/fault"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/rs/zerolog/log"
	"github.com/siemens/wfx/generated/ent"
	"github.com/siemens/wfx/persistence"

	// this is the sqlite driver code
	_ "github.com/mattn/go-sqlite3"
)

//go:embed migrations/sqlite/*.sql
var sqliteMigrations embed.FS

type SQLite struct {
	Database
}

func init() {
	persistence.RegisterStorage("sqlite", &SQLite{})
}

func (wrapper *SQLite) Initialize(_ context.Context, options string) error {
	log.Debug().
		Str("dsn", options).
		Msg("Initializing SQLite storage")

	src, err := iofs.New(sqliteMigrations, "migrations/sqlite")
	if err != nil {
		return fault.Wrap(err)
	}

	var sqlite sqlite3.Sqlite
	driver, err := sqlite.Open(options)
	if err != nil {
		return fault.Wrap(err)
	}

	if err := runMigrations(src, "wfx", driver); err != nil {
		return fault.Wrap(err)
	}

	log.Debug().Msg("Connecting to SQLite")
	client, err := ent.Open("sqlite3", options)
	if err != nil {
		log.Error().Err(err).Msg("Failed opening connection to sqlite")
		return fault.Wrap(err)
	}
	log.Debug().Msg("Connected to SQLite")

	wrapper.Database = Database{client: client}
	return nil
}
