//go:build integration && !no_mysql

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
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/siemens/wfx/internal/persistence/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMySQL_Initialize(t *testing.T) {
	db := setupMySQL(t)
	db.Shutdown()
}

func TestMain_InitializeFail(t *testing.T) {
	dsn := "foo:bar@tcp(localhost)/wfx"
	var mysql MySQL
	err := mysql.Initialize(dsn)
	assert.NotNil(t, err)
}

func TestMySQL(t *testing.T) {
	db := setupMySQL(t)
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

func setupMySQL(t *testing.T) MySQL {
	user := os.Getenv("MYSQL_USER")
	pass := os.Getenv("MYSQL_PASSWORD")
	host := os.Getenv("MYSQL_HOST")
	db := os.Getenv("MYSQL_DATABASE")
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s", user, pass, host, db)
	var mysql MySQL
	err := mysql.Initialize(dsn)
	require.NoError(t, err)
	return mysql
}
