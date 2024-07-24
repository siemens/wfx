package cmd

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
	"github.com/stretchr/testify/require"
)

func TestRootCmd_ManPage(t *testing.T) {
	cmd := NewCommand()
	cmd, _, err := cmd.Find([]string{"man"})
	require.NoError(t, err)
	assert.NotNil(t, cmd)
}
