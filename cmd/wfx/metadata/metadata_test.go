package metadata

/*
 * SPDX-FileCopyrightText: 2024 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildDate(t *testing.T) {
	_, err := BuildDate()
	require.NoError(t, err)
}
