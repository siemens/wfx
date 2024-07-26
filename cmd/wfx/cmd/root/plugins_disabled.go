//go:build !plugin

package root

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"errors"

	"github.com/knadh/koanf/v2"
	"github.com/siemens/wfx/middleware"
)

func loadPluginSet(flag string, _ chan error) ([]middleware.IntermediateMW, error) {
	var pluginsDir string
	k.Read(func(k *koanf.Koanf) {
		pluginsDir = k.String(flag)
	})
	if pluginsDir != "" {
		return nil, errors.New("this binary was built without plugin support")
	}
	return []middleware.IntermediateMW{}, nil
}
