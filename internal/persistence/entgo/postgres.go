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
	"database/sql/driver"
	"embed"
	"fmt"
	"net"
	"os"
	"strings"
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

// iamDB implements driver.Connector to support AWS RDS IAM authentication
type iamDB struct {
	config     *pgx.ConnConfig
	awsConfig  aws.Config
	useIAMAuth bool
}

func init() {
	persistence.RegisterStorage("postgres", &PostgreSQL{})
}

// Connect implements driver.Connector interface
// This method is called each time a new connection is needed, allowing for fresh IAM token generation
func (id *iamDB) Connect(ctx context.Context) (driver.Conn, error) {
	connConfig := id.config.Copy()

	if id.useIAMAuth {
		// Resolve the actual RDS endpoint (handles CNAME lookups)
		actualHost, region, err := resolveRDSEndpoint(connConfig.Host)
		if err != nil {
			log.Error().Err(err).Str("host", connConfig.Host).Msg("Failed to resolve RDS endpoint")
			return nil, fault.Wrap(err)
		}

		// Generate new IAM authentication token
		authToken, err := auth.BuildAuthToken(
			ctx,
			fmt.Sprintf("%s:%d", actualHost, connConfig.Port),
			region,
			connConfig.User,
			id.awsConfig.Credentials,
		)
		if err != nil {
			log.Error().Err(err).Msg("Failed to build IAM auth token")
			return nil, fault.Wrap(err)
		}

		// Update connection config with new token and actual host
		connConfig.Host = actualHost
		connConfig.Password = authToken

		log.Debug().
			Str("host", actualHost).
			Str("region", region).
			Str("user", connConfig.User).
			Msg("Generated IAM auth token for RDS connection")
	}

	// Use stdlib to get a proper database/sql compatible connector
	stdlibConn := stdlib.GetConnector(*connConfig)
	return stdlibConn.Connect(ctx)
}

// Driver implements driver.Connector interface
func (id *iamDB) Driver() driver.Driver {
	return id
}

// Open implements driver.Driver interface (not supported for IAM auth)
func (id *iamDB) Open(name string) (driver.Conn, error) {
	return nil, fmt.Errorf("driver open method not supported for IAM authentication")
}

// resolveRDSEndpoint resolves CNAME to actual RDS endpoint and extracts region
// Returns: actualHost, region, error
func resolveRDSEndpoint(host string) (string, string, error) {
	// Try to resolve CNAME
	cname, err := net.LookupCNAME(host)
	if err != nil {
		// If lookup fails, assume it's already the actual endpoint
		log.Debug().Str("host", host).Msg("CNAME lookup failed, using host as-is")
		cname = host
	} else {
		// Trim trailing dot from CNAME
		cname = strings.TrimRight(cname, ".")
		log.Debug().Str("original", host).Str("resolved", cname).Msg("Resolved CNAME")
	}

	// Parse region from RDS endpoint format: <instance>.<id>.<region>.rds.amazonaws.com
	parts := strings.Split(cname, ".")
	if len(parts) >= 6 && strings.Contains(cname, ".rds.amazonaws.com") {
		region := parts[len(parts)-4] // Region is 4th from the end
		return cname, region, nil
	}

	// If not standard RDS format, try to get region from environment or AWS config
	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = os.Getenv("AWS_DEFAULT_REGION")
	}
	if region == "" {
		// Default to us-east-1 if no region found
		region = "us-east-1"
		log.Warn().Str("host", cname).Msg("Could not determine region from endpoint, using default: us-east-1")
	}

	return cname, region, nil
}

// checkIAMAuthEnabled checks if IAM authentication should be enabled based on environment variable
func checkIAMAuthEnabled() bool {
	iamAuth := os.Getenv("WFX_POSTGRES_IAM_AUTH")
	return iamAuth == "true" || iamAuth == "1" || iamAuth == "yes"
}

// Initialize sets up the PostgreSQL database connection using the provided DSN (options), runs migrations, and initializes the ent client.
//
// Parameters:
//   - options: PostgreSQL connection string (DSN) in the format:
//     "user=<username> password=<password> host=localhost port=5432 database=wfx sslmode=disable"
//     See https://github.com/jackc/pgx/blob/master/stdlib/sql.go for DSN syntax details.
//
// IAM Authentication:
//   - Set environment variable WFX_POSTGRES_IAM_AUTH=true to enable AWS RDS IAM authentication
//   - When IAM auth is enabled, the password in DSN is ignored and IAM tokens are generated automatically
//   - Requires AWS credentials to be configured (via IRSA, instance profile, or environment variables)
//   - AWS region is auto-detected from RDS endpoint or AWS_REGION environment variable
func (wrapper *PostgreSQL) Initialize(options string) error {
	connConfig, err := pgx.ParseConfig(options)
	if err != nil {
		return fault.Wrap(err)
	}

	useIAMAuth := checkIAMAuthEnabled()
	var db *sql.DB

	if useIAMAuth {
		log.Info().Msg("IAM authentication enabled for PostgreSQL")

		// Load AWS configuration (supports IRSA, instance profiles, env vars, etc.)
		ctx := context.Background()
		awsConfig, err := config.LoadDefaultConfig(ctx)
		if err != nil {
			log.Error().Err(err).Msg("Failed to load AWS configuration for IAM auth")
			return fault.Wrap(err)
		}

		// Create custom connector with IAM auth support
		connector := &iamDB{
			config:     connConfig,
			awsConfig:  awsConfig,
			useIAMAuth: true,
		}

		// Open database connection using the custom connector
		db = sql.OpenDB(connector)
		log.Info().
			Str("host", connConfig.Host).
			Str("user", connConfig.User).
			Str("database", connConfig.Database).
			Msg("Initialized PostgreSQL with IAM authentication")
	} else {
		// Standard authentication with password
		// connConfig.Logger = zerolog.Logger
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

	// Test the connection
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
