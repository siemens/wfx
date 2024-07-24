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

	"github.com/stretchr/testify/assert"
)

func TestDefaultConfigFiles(t *testing.T) {
	fnames := DefaultConfigFiles()
	assert.NotEmpty(t, fnames)
}
