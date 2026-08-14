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
