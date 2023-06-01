package config

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"path"

	"github.com/OpenPeeDeeP/xdg"
)

func DefaultConfigFiles() []string {
	configFiles := []string{
		// current directory
		"wfx.yml",
		// user home
		path.Join(xdg.ConfigHome(), "wfx", "config.yml"),
		path.Join("/etc/wfx/wfx.yml"),
	}
	return configFiles
}
