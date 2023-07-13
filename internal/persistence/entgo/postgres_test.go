//go:build integration && postgres

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
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/siemens/wfx/internal/persistence/tests"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestPostgreSQL_Initialize(t *testing.T) {
	defer goleak.VerifyNone(t)
	db := setupPostgreSQL(t)
	db.Shutdown()
}

func TestPostgreSQL(t *testing.T) {
	db := setupPostgreSQL(t)
	t.Cleanup(db.Shutdown)
	for _, testFn := range tests.AllTests {
		name := runtime.FuncForPC(reflect.ValueOf(testFn).Pointer()).Name()
		name = strings.TrimPrefix(filepath.Ext(name), ".")
		t.Run(name, func(t *testing.T) {
			defer resetDB(t, db.Database)
			testFn(t, &db)
		})
	}
}

func setupPostgreSQL(t *testing.T) PostgreSQL {
	var postgres PostgreSQL
	err := postgres.Initialize(context.Background(), "sslmode=disable")
	require.NoError(t, err)
	return postgres
}
