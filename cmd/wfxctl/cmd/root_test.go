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

	"github.com/siemens/wfx/cmd/wfxctl/flags"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRootCmd_ManPage(t *testing.T) {
	cmd, _, err := RootCmd.Find([]string{"man"})
	require.NoError(t, err)
	assert.NotNil(t, cmd)
}

func TestPersistentPreRunE(t *testing.T) {
	t.Setenv("WFX_LOG_LEVEL", "trace")
	err := RootCmd.PersistentPreRunE(RootCmd, nil)
	require.NoError(t, err)
	assert.Equal(t, "trace", flags.Koanf.String(logLevelFlag))
}
