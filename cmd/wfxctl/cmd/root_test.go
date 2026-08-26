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

func TestRootCmd_HostFlag(t *testing.T) {
	cmd := NewCommand()
	flag := cmd.PersistentFlags().Lookup("host")
	require.NotNil(t, flag)
	assert.Equal(t, "http://localhost:8081", flag.DefValue)

	for _, removed := range []string{
		"enable-tls",
		"client-host", "client-port", "client-tls-host", "client-tls-port", "client-unix-socket",
		"mgmt-host", "mgmt-port", "mgmt-tls-host", "mgmt-tls-port", "mgmt-unix-socket",
	} {
		assert.Nil(t, cmd.PersistentFlags().Lookup(removed))
	}
}
