package colors

/*
 * SPDX-FileCopyrightText: 2024 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"testing"

	"github.com/siemens/wfx/workflow/dau"
	"github.com/stretchr/testify/assert"
)

func TestGroupColor(t *testing.T) {
	cp := NewColorPalette(dau.DirectWorkflow())
	c := cp.GroupColor("OPEN")
	assert.NotNil(t, c)
	c = cp.GroupColor("FOO")
	assert.Nil(t, c)
}

func TestStateColor(t *testing.T) {
	cp := NewColorPalette(dau.DirectWorkflow())
	fg, bg := cp.StateColor("INSTALL")
	assert.NotEmpty(t, fg)
	assert.NotEmpty(t, bg)
}

func TestStateColor_Fallback(t *testing.T) {
	cp := NewColorPalette(dau.DirectWorkflow())
	fg, bg := cp.StateColor("FOO")
	assert.Equal(t, DefaultFgColor, fg)
	assert.Equal(t, DefaultBgColor, bg)
}
