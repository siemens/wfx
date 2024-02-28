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

	"github.com/siemens/wfx/middleware"
)

func LoadNorthboundPlugins(chan error) ([]middleware.IntermediateMW, error) {
	return nil, errors.New("this binary was built without plugin support")
}

func LoadSouthboundPlugins(chan error) ([]middleware.IntermediateMW, error) {
	return nil, errors.New("this binary was built without plugin support")
}
