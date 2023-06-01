package root

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServerKind(t *testing.T) {
	assert.Equal(t, "http", kindHTTP.String())
	assert.Equal(t, "https", kindHTTPS.String())
	assert.Equal(t, "unix", kindUnix.String())
}
