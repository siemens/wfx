package main

/*
 * SPDX-FileCopyrightText: 2026 Siemens AG
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

func TestRootCmd_HostFlags(t *testing.T) {
	flags := NewCommand().PersistentFlags()

	clientHost := flags.Lookup("client-host")
	require.NotNil(t, clientHost)
	assert.Equal(t, "http://localhost:8080", clientHost.DefValue)

	mgmtHost := flags.Lookup("mgmt-host")
	require.NotNil(t, mgmtHost)
	assert.Equal(t, "http://localhost:8081", mgmtHost.DefValue)

	assert.Nil(t, flags.Lookup("client-port"))
	assert.Nil(t, flags.Lookup("mgmt-port"))
}
