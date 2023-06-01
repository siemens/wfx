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
	"testing"

	"github.com/stretchr/testify/require"
)

func resetDB(t *testing.T, db Database) {
	queries := []string{
		"delete from tag",
		"delete from job",
		"delete from history",
		"delete from workflow",
	}
	for _, query := range queries {
		_, err := db.client.ExecContext(context.Background(), query)
		require.NoError(t, err)
	}
}
