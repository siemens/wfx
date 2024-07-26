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

	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
)

func TestLoadNorthboundPlugins(t *testing.T) {
	k.Write(func(k *koanf.Koanf) {
		_ = k.Set(mgmtPluginsDirFlag, "/plugins")
	})
	mw, err := LoadNorthboundPlugins(nil)
	assert.Nil(t, mw)
	assert.ErrorContains(t, err, "this binary was built without plugin support")
}

func TestLoadNorthboundPlugins_None(t *testing.T) {
	mw, err := LoadNorthboundPlugins(nil)
	assert.NoError(t, err)
	assert.Empty(t, mw)
}

func TestLoadSouthboundPlugins(t *testing.T) {
	k.Write(func(k *koanf.Koanf) {
		_ = k.Set(clientPluginsDirFlag, "/plugins")
	})
	mw, err := LoadSouthboundPlugins(nil)
	assert.Nil(t, mw)
	assert.ErrorContains(t, err, "this binary was built without plugin support")
}

func TestLoadSouthboundPlugins_None(t *testing.T) {
	mw, err := LoadSouthboundPlugins(nil)
	assert.NoError(t, err)
	assert.Empty(t, mw)
}
