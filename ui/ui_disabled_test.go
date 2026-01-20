//go:build !ui

package ui

/*
 * SPDX-FileCopyrightText: 2025 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMux(t *testing.T) {
	assert.NotNil(t, Mux("", ""))
}

func TestFaviconHandler(t *testing.T) {
	assert.Nil(t, FaviconHandler())
}
