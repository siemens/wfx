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

	"github.com/Southclaws/fault"
)

func (db Database) CheckHealth(_ context.Context) error {
	_, err := db.client.ExecContext(context.Background(), "SELECT 1")
	return fault.Wrap(err)
}
