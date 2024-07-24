package man

/*
 * SPDX-FileCopyrightText: 2024 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"fmt"
	"time"

	mcobra "github.com/muesli/mango-cobra"
	"github.com/muesli/roff"
	"github.com/spf13/cobra"
)

// NewCommand returns a new cobra command that generates man pages.
func NewCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "man",
		Short: "Generate man page and exit",
		RunE: func(cmd *cobra.Command, _ []string) error {
			manPage, _ := mcobra.NewManPage(1, cmd.Root())
			year := time.Now().Year()
			manPage = manPage.WithSection("Copyright", fmt.Sprintf(`(C) %d Siemens AG.
Licensed under the Apache License, Version 2.0`, year))
			fmt.Fprintln(cmd.OutOrStdout(), manPage.Build(roff.NewDocument()))
			return nil
		},
	}
}
