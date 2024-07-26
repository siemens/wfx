package root

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"github.com/siemens/wfx/middleware"
)

func LoadNorthboundPlugins(chQuit chan error) ([]middleware.IntermediateMW, error) {
	return loadPluginSet(mgmtPluginsDirFlag, chQuit)
}

func LoadSouthboundPlugins(chQuit chan error) ([]middleware.IntermediateMW, error) {
	return loadPluginSet(clientPluginsDirFlag, chQuit)
}
