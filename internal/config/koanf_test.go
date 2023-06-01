package config

import (
	"testing"

	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
)

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

func TestReadAndWrite(t *testing.T) {
	cfg := New()
	cfg.Write(func(k *koanf.Koanf) {
		_ = k.Set("hello", "world")
	})
	var actual string
	cfg.Read(func(k *koanf.Koanf) {
		actual = k.String("hello")
	})
	assert.Equal(t, "world", actual)
}
