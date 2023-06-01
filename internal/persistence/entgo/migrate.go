package entgo

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"errors"

	"github.com/Southclaws/fault"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/source"
	"github.com/rs/zerolog/log"
)

func runMigrations(src source.Driver, db string, driver database.Driver) error {
	mig, err := migrate.NewWithInstance("migrations", src, db, driver)
	if err != nil {
		return fault.Wrap(err)
	}

	err = mig.Up()
	if err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Debug().Msg("No schema changes necessary")
			return nil
		}
		log.Error().Err(err).Msg("Failed to apply migrations")
		return fault.Wrap(err)
	}
	log.Info().Msg("Applied schema migrations")
	return nil
}
