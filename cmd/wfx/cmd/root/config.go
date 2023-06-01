package root

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"fmt"
	"os"

	"github.com/Southclaws/fault"
	"github.com/knadh/koanf/v2"
	"github.com/rs/zerolog"
)

func reloadConfig(k *koanf.Koanf) error {
	fmt.Fprintln(os.Stderr, "Reloading config")

	lvlString := k.String(logLevelFlag)
	lvl, err := zerolog.ParseLevel(lvlString)
	if err != nil {
		return fault.Wrap(err)
	}
	fmt.Fprintln(os.Stderr, "Setting global log level:", lvl)
	zerolog.SetGlobalLevel(lvl)
	return nil
}
