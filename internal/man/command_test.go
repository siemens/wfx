package man

/*
 * SPDX-FileCopyrightText: 2024 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewManCmd(t *testing.T) {
	buf := new(bytes.Buffer)
	cmd := NewCommand()
	cmd.SetOutput(buf)
	err := cmd.Execute()
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "Siemens AG")
}
