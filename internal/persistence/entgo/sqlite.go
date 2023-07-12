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
	"net/url"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
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

func (instance *SQLite) Initialize(_ context.Context, dsn string) error {
	log.Debug().Str("dsn", dsn).Msg("Connecting to SQLite")
	drv, err := sql.Open(dialect.SQLite, dsn)
	if err != nil {
		log.Error().Err(err).Msg("Failed opening connection to SQLite")
		return fault.Wrap(err)
	}
	client := ent.NewClient(ent.Driver(drv))
	log.Debug().Msg("Connected to SQLite")
	instance.Database = Database{client: client}

	{
		// run schema migrations
		src, err := iofs.New(sqliteMigrations, "migrations/sqlite")
		if err != nil {
			return fault.Wrap(err)
		}

		purl, err := url.Parse(dsn)
		if err != nil {
			return fault.Wrap(err)
		}

		driver, err := sqlite3.WithInstance(drv.DB(), &sqlite3.Config{
			MigrationsTable: sqlite3.DefaultMigrationsTable,
			DatabaseName:    purl.Path,
			NoTxWrap:        false,
		})
		if err != nil {
			return fault.Wrap(err)
		}

		if err := runMigrations(src, "wfx", driver); err != nil {
			return fault.Wrap(err)
		}
	}
	return nil
}
