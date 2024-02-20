package output

/*
 * SPDX-FileCopyrightText: 2024 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSVG(t *testing.T) {
	_, ok := Generators["svg"]
	assert.True(t, ok)
}

func TestPlantUML(t *testing.T) {
	_, ok := Generators["plantuml"]
	assert.True(t, ok)
}

func TestSMCat(t *testing.T) {
	_, ok := Generators["smcat"]
	assert.True(t, ok)
}

func TestMermaid(t *testing.T) {
	_, ok := Generators["mermaid"]
	assert.True(t, ok)
}
