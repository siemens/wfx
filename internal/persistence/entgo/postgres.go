//go:build postgres

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
	"time"

	"github.com/Southclaws/fault"
	migrate_pgx "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	// using the same version as in https://github.com/golang-migrate/migrate/blob/v4.16.2/database/pgx/v5/pgx.go
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/rs/zerolog/log"

	"github.com/siemens/wfx/generated/ent"
	"github.com/siemens/wfx/persistence"
)

//go:embed migrations/postgres/*.sql
var postgresMigrations embed.FS

type PostgreSQL struct {
	Database
}

func init() {
	persistence.RegisterStorage("postgres", &PostgreSQL{})
}

// example dsn: "user=<username> password=<password> host=localhost port=5432 database=wfx sslmode=disable"
// see https://github.com/jackc/pgx/blob/master/stdlib/sql.go
func (wrapper *PostgreSQL) Initialize(ctx context.Context, options string) error {
	connConfig, err := pgx.ParseConfig(options)
	if err != nil {
		return fault.Wrap(err)
	}
	// connConfig.Logger = zerolog.Logger
	connStr := stdlib.RegisterConnConfig(connConfig)

	log.Debug().Msg("Initializing PostgreSQL storage")
	db, err := sql.Open("pgx", connStr)
	// consider exposing db.SetMaxIdleConns() in the future
	if err != nil {
		return fault.Wrap(err)
	}
	if err := db.PingContext(ctx); err != nil {
		log.Error().Err(err).Msg("Failed to ping PostgreSQL database")
		return fault.Wrap(err)
	}

	driver, err := migrate_pgx.WithInstance(db, &migrate_pgx.Config{
		MigrationsTable:       migrate_pgx.DefaultMigrationsTable,
		StatementTimeout:      time.Hour,
		MultiStatementEnabled: false,
		MultiStatementMaxSize: migrate_pgx.DefaultMultiStatementMaxSize,
	})
	if err != nil {
		return fault.Wrap(err)
	}

	src, err := iofs.New(postgresMigrations, "migrations/postgres")
	if err != nil {
		return fault.Wrap(err)
	}

	if err := runMigrations(src, connConfig.Database, driver); err != nil {
		return fault.Wrap(err)
	}

	// Create an ent.Driver from `db`.
	drv := entsql.OpenDB(dialect.Postgres, db)
	client := ent.NewClient(ent.Driver(drv))
	wrapper.Database = Database{client: client}
	return nil
}
