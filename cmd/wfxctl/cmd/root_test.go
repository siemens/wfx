package cmd

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManPageSubcommand(t *testing.T) {
	buf := new(bytes.Buffer)
	cmd := NewCommand()
	cmd.SetOutput(buf)
	cmd.SetArgs([]string{"man"})
	err := cmd.Execute()
	require.NoError(t, err)
	assert.NotEmpty(t, buf.Bytes())
}
