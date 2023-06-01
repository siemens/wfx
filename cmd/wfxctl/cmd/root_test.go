package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

func TestRootCmd_ManPage(t *testing.T) {
	buf := new(bytes.Buffer)
	RootCmd.SetOutput(buf)
	RootCmd.SetArgs([]string{"man"})
	err := RootCmd.Execute()
	require.NoError(t, err)
	assert.NotEmpty(t, buf.Bytes())
}
