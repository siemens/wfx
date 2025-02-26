//go:build no_plugin

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

	"github.com/siemens/wfx/middleware/plugin"
)

func loadPlugins(dir string) ([]plugin.Plugin, error) {
	if dir != "" {
		return nil, errors.New("this binary was built without plugin support")
	}
	return []plugin.Plugin{}, nil
}
