//go:build !plugin

package root

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import "github.com/siemens/wfx/middleware"

func LoadNorthboundPlugins(chan error) ([]middleware.IntermediateMW, error) {
	return []middleware.IntermediateMW{}, nil
}

func LoadSouthboundPlugins(chan error) ([]middleware.IntermediateMW, error) {
	return []middleware.IntermediateMW{}, nil
}
