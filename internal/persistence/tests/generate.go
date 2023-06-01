//go:build testing

//go:generate go run ./generator.go

package tests

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"testing"

	"github.com/siemens/wfx/persistence"
)

type PersistenceTest func(t *testing.T, db persistence.Storage)
