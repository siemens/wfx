//go:build ignore

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"ariga.io/atlas/sql/sqltool"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql/schema"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/siemens/wfx/generated/ent/migrate"
)

func main() {
	if len(os.Args) != 3 {
		log.Fatalln("migration name is required. Use: 'go run -mod=mod generated/ent/migrate/main.go <driver> <name>'")
	}
	driver := os.Args[1]
	name := os.Args[2]

	switch driver {
	case "sqlite":
		d := "internal/persistence/entgo/migrations/sqlite"
		_ = os.MkdirAll(d, 0o755)
		dir, err := sqltool.NewGolangMigrateDir(d)
		if err != nil {
			log.Fatalf("failed creating atlas migration directory: %v", err)
		}
		// Migrate diff options.
		opts := []schema.MigrateOption{
			schema.WithDir(dir),
			schema.WithMigrationMode(schema.ModeReplay),
			schema.WithDialect(dialect.SQLite),
		}

		// Generate migrations using Atlas support (note the Ent dialect option passed above).
		err = migrate.NamedDiff(context.Background(), "sqlite://wfx?mode=memory&cache=shared&_fk=1", name, opts...)
		if err != nil {
			log.Fatalf("failed generating migration file: %v", err)
		}
	case "postgres":
		d := "internal/persistence/entgo/migrations/postgres"
		_ = os.MkdirAll(d, 0o755)
		dir, err := sqltool.NewGolangMigrateDir(d)
		if err != nil {
			log.Fatalf("failed creating atlas migration directory: %v", err)
		}
		// Migrate diff options.
		opts := []schema.MigrateOption{
			schema.WithDir(dir),
			schema.WithMigrationMode(schema.ModeReplay),
			schema.WithDialect(dialect.Postgres),
		}

		// Generate migrations using Atlas support (note the Ent dialect option passed above).
		url := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", os.Getenv("PGUSER"), os.Getenv("PGPASSWORD"),
			os.Getenv("PGHOST"), os.Getenv("PGDATABASE"))
		err = migrate.NamedDiff(context.Background(), url, name, opts...)
		if err != nil {
			log.Fatalf("failed generating migration file: %v", err)
		}
	case "mysql":
		d := "internal/persistence/entgo/migrations/mysql"
		_ = os.MkdirAll(d, 0o755)
		dir, err := sqltool.NewGolangMigrateDir(d)
		if err != nil {
			log.Fatalf("failed creating atlas migration directory: %v", err)
		}
		// Migrate diff options.
		opts := []schema.MigrateOption{
			schema.WithDir(dir),
			schema.WithMigrationMode(schema.ModeReplay),
			schema.WithDialect(dialect.MySQL),
		}

		// Generate migrations using Atlas support (note the Ent dialect option passed above).
		url := fmt.Sprintf("mysql://%s:%s@%s/%s", os.Getenv("MYSQL_USER"), os.Getenv("MYSQL_PASSWORD"),
			os.Getenv("MYSQL_HOST"), os.Getenv("MYSQL_DATABASE"))
		err = migrate.NamedDiff(context.Background(), url, name, opts...)
		if err != nil {
			log.Fatalf("failed generating migration file: %v", err)
		}
	default:
		log.Fatalf("unsupported driver: %s", driver)
	}
}
