//go:build !no_sqlite

package entgo

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/siemens/wfx/internal/persistence/tests"
	"github.com/siemens/wfx/persistence"
	"github.com/stretchr/testify/require"
)

func TestSQLite(t *testing.T) {
	db := setupSQLite(t)
	t.Cleanup(db.Shutdown)
	var storage persistence.Storage = &db
	for _, testFn := range tests.AllTests {
		name := runtime.FuncForPC(reflect.ValueOf(testFn).Pointer()).Name()
		name = strings.TrimPrefix(filepath.Ext(name), ".")
		t.Run(name, func(t *testing.T) {
			defer resetDB(t, db.Database)
			testFn(t, storage)
		})
	}
}

func setupSQLite(t *testing.T) SQLite {
	var sqlite SQLite

	dir, err := os.MkdirTemp("", "wfx.db.*")
	require.NoError(t, err)
	f, err := os.Create(path.Join(dir, "wfx.db"))
	require.NoError(t, err)
	t.Logf("Database is backed by file %s", f.Name())
	t.Cleanup(func() {
		_ = os.RemoveAll(dir)
	})

	opts := fmt.Sprintf("file:%s?_fk=1&_journal=WAL", f.Name())
	err = sqlite.Initialize(opts)
	require.NoError(t, err)
	require.NotNil(t, sqlite.Database)
	return sqlite
}
