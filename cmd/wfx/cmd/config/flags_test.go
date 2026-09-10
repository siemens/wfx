package config

/*
 * SPDX-FileCopyrightText: 2024 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfigFiles(t *testing.T) {
	fnames := DefaultConfigFiles()
	assert.NotEmpty(t, fnames)
}

func TestHostFlags(t *testing.T) {
	flags := NewFlagset()
	require.NoError(t, flags.Parse([]string{
		"--" + ClientHostFlag, "https://0.0.0.0:8443",
		"--" + ClientHostFlag, "unix:///tmp/wfx-client.sock",
		"--" + MgmtHostFlag, "http://127.0.0.1:8081",
	}))

	clientHosts, err := flags.GetStringSlice(ClientHostFlag)
	require.NoError(t, err)
	assert.Equal(t, []string{"https://0.0.0.0:8443", "unix:///tmp/wfx-client.sock"}, clientHosts)

	mgmtHosts, err := flags.GetStringSlice(MgmtHostFlag)
	require.NoError(t, err)
	assert.Equal(t, []string{"http://127.0.0.1:8081"}, mgmtHosts)

	for _, removed := range []string{
		"scheme", "client-port", "client-tls-host", "client-tls-port", "client-unix-socket",
		"mgmt-port", "mgmt-tls-host", "mgmt-tls-port", "mgmt-unix-socket",
	} {
		assert.Nil(t, flags.Lookup(removed))
	}
}

func TestJQFilterTimeoutFlag(t *testing.T) {
	flags := NewFlagset()
	require.NoError(t, flags.Parse([]string{"--" + JQFilterTimeoutFlag, "5s"}))
	timeout, err := flags.GetDuration(JQFilterTimeoutFlag)
	require.NoError(t, err)
	assert.Equal(t, 5*time.Second, timeout)
}

func TestJQFilterMaxResponseSizeFlag(t *testing.T) {
	flags := NewFlagset()
	require.NoError(t, flags.Parse([]string{"--" + JQFilterMaxResponseSizeFlag, "1024"}))
	maxResponseSize, err := flags.GetInt(JQFilterMaxResponseSizeFlag)
	require.NoError(t, err)
	assert.Equal(t, 1024, maxResponseSize)
}
