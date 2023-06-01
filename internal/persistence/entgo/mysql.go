//go:build mysql

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
	"database/sql"
	"embed"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/Southclaws/fault"
	driver "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/rs/zerolog/log"
	"github.com/siemens/wfx/generated/ent"
	"github.com/siemens/wfx/persistence"
)

//go:embed migrations/mysql/*.sql
var mysqlMigrations embed.FS

type MySQL struct {
	Database
}

func init() {
	persistence.RegisterStorage("mysql", &MySQL{})
}

func (wrapper *MySQL) Initialize(ctx context.Context, options string) error {
	// parse user-supplied dsn and enrich it
	cfg, err := driver.ParseDSN(options)
	if err != nil {
		return fault.Wrap(err)
	}
	cfg.ParseTime = true // needed to store timestamp with microsecond precision
	log.Debug().
		Str("user", cfg.User).
		Str("addr", cfg.Addr).
		Msg("Initializing MySQL storage")

	// needed for golang-migrate
	cfg.MultiStatements = true
	connector, err := driver.NewConnector(cfg)
	if err != nil {
		return fault.Wrap(err)
	}

	db := sql.OpenDB(connector)
	if err := db.PingContext(ctx); err != nil {
		log.Error().Err(err).Msg("Failed to ping PostgreSQL database")
		return fault.Wrap(err)
	}

	log.Info().Msg("Applying migrations")
	src, err := iofs.New(mysqlMigrations, "migrations/mysql")
	if err != nil {
		return fault.Wrap(err)
	}

	{ // run migrations
		drv, err := mysql.WithInstance(db, &mysql.Config{})
		if err != nil {
			return fault.Wrap(err)
		}
		if err := runMigrations(src, cfg.DBName, drv); err != nil {
			return fault.Wrap(err)
		}
	}

	// Create an ent.Driver from `db`.
	drv := entsql.OpenDB(dialect.MySQL, db)
	client := ent.NewClient(ent.Driver(drv))

	wrapper.Database = Database{client: client}
	return nil
}
