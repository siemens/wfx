//go:build !no_postgres

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
	"fmt"
	"os"
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

	// AWS SDK for IAM authentication
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/rds/auth"
)

//go:embed migrations/postgres/*.sql
var postgresMigrations embed.FS

const databaseConnectionTimeoutMilliseconds = 5000

type PostgreSQL struct {
	Database
}

func init() {
	persistence.RegisterStorage("postgres", &PostgreSQL{})
}

// iamAuthHook is a BeforeConnect hook that generates IAM authentication tokens for each connection.
func iamAuthHook(awsConfig aws.Config, region string) func(context.Context, *pgx.ConnConfig) error {
	return func(ctx context.Context, connConfig *pgx.ConnConfig) error {
		// Generate new IAM authentication token
		authToken, err := auth.BuildAuthToken(
			ctx,
			fmt.Sprintf("%s:%d", connConfig.Host, connConfig.Port),
			region,
			connConfig.User,
			awsConfig.Credentials,
		)
		if err != nil {
			log.Error().Err(err).Msg("Failed to build IAM auth token")
			return fault.Wrap(err)
		}

		// Update connection config with new token
		connConfig.Password = authToken

		log.Debug().
			Str("host", connConfig.Host).
			Str("region", region).
			Str("user", connConfig.User).
			Msg("Generated IAM auth token for RDS connection")

		return nil
	}
}

// checkIAMAuthEnabled checks if IAM authentication is enabled via environment variable.
func checkIAMAuthEnabled() bool {
	iamAuth := os.Getenv("WFX_POSTGRES_IAM_AUTH")
	return iamAuth == "true" || iamAuth == "1" || iamAuth == "yes"
}

// Initialize sets up the PostgreSQL database connection using the provided DSN (options), runs migrations, and initializes the ent client.
//
// Parameters:
//   - options: PostgreSQL connection string (DSN) in the format:
//     "user=<username> password=<password> host=localhost port=5432 database=wfx sslmode=disable"
//     Or use environment variables: PGHOST, PGPORT, PGUSER, PGDATABASE, PGSSLMODE
//     See https://github.com/jackc/pgx/blob/master/stdlib/sql.go for DSN syntax details.
//
// IAM Authentication:
//   - Set environment variable WFX_POSTGRES_IAM_AUTH=true to enable AWS RDS IAM authentication
//   - When IAM auth is enabled, the password in DSN is ignored and IAM tokens are generated automatically
//   - Requires AWS credentials to be configured (via IRSA, instance profile, or environment variables)
//   - AWS region must be specified via WFX_POSTGRES_REGION or AWS_REGION environment variables
func (wrapper *PostgreSQL) Initialize(options string) error {
	connConfig, err := pgx.ParseConfig(options)
	if err != nil {
		return fault.Wrap(err)
	}

	useIAMAuth := checkIAMAuthEnabled()
	var db *sql.DB

	if useIAMAuth {
		log.Info().Msg("IAM authentication enabled for PostgreSQL")

		// Get AWS region from environment variables
		region := os.Getenv("WFX_POSTGRES_REGION")
		if region == "" {
			region = os.Getenv("AWS_REGION")
		}
		if region == "" {
			return fmt.Errorf("AWS region not configured: set WFX_POSTGRES_REGION or AWS_REGION environment variable")
		}

		// Load AWS configuration (supports IRSA, instance profiles, env vars, etc.)
		ctx := context.Background()
		awsConfig, err := config.LoadDefaultConfig(ctx)
		if err != nil {
			log.Error().Err(err).Msg("Failed to load AWS configuration for IAM auth")
			return fault.Wrap(err)
		}

		connector := stdlib.GetConnector(*connConfig, stdlib.OptionBeforeConnect(iamAuthHook(awsConfig, region)))
		db = sql.OpenDB(connector)

		log.Info().
			Str("host", connConfig.Host).
			Str("user", connConfig.User).
			Str("database", connConfig.Database).
			Str("region", region).
			Msg("Initialized PostgreSQL with IAM authentication")
	} else {
		// Standard authentication with password
		connStr := stdlib.RegisterConnConfig(connConfig)

		log.Debug().Msg("Initializing PostgreSQL storage")
		db, err = sql.Open("pgx", connStr)
		// consider exposing db.SetMaxIdleConns() in the future
		if err != nil {
			return fault.Wrap(err)
		}
		log.Info().
			Str("host", connConfig.Host).
			Str("user", connConfig.User).
			Str("database", connConfig.Database).
			Msg("Initialized PostgreSQL with password authentication")
	}

	ctx, cancel := context.WithTimeout(context.Background(), databaseConnectionTimeoutMilliseconds*time.Millisecond)
	defer cancel()

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
