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

	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/require"
)

func TestReloadConfig_Invalid(t *testing.T) {
	var inner *koanf.Koanf
	k.Write(func(k *koanf.Koanf) {
		_ = k.Set(logLevelFlag, "foo")
		// only do this in a test!
		inner = k
	})
	err := reloadConfig(inner)
	require.Error(t, err)
}
