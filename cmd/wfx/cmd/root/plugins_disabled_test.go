//go:build !plugin

package root

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadNorthboundPlugins(t *testing.T) {
	mw, err := LoadNorthboundPlugins(nil)
	assert.Nil(t, mw)
	assert.ErrorContains(t, err, "this binary was built without plugin support")
}

func TestLoadSouthboundPlugins(t *testing.T) {
	mw, err := LoadSouthboundPlugins(nil)
	assert.Nil(t, mw)
	assert.ErrorContains(t, err, "this binary was built without plugin support")
}
