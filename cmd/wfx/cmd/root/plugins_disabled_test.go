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

func TestLoadPlugins(t *testing.T) {
	plugins, err := loadPlugins("/plugins")
	assert.Nil(t, plugins)
	assert.ErrorContains(t, err, "this binary was built without plugin support")
}

func TestLoadNorthboundPlugins_None(t *testing.T) {
	plugins, err := loadPlugins("")
	assert.NoError(t, err)
	assert.Empty(t, plugins)
}
