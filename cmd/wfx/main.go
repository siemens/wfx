package main

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

	"github.com/siemens/wfx/cmd/wfx/cmd/root"
)

func main() {
	if err := root.NewCommand().Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "wfx encountered an error: %+v\n", err)
		os.Exit(1)
	}
}
