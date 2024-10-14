package man

/*
 * SPDX-FileCopyrightText: 2024 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"os"

	"github.com/Southclaws/fault"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var (
	manDir  string
	Command = &cobra.Command{
		Use:   "man",
		Short: "Generate man pages and exit",
		RunE: func(cmd *cobra.Command, _ []string) error {
			rootCmd := cmd.Root()
			disableAutoGenTag(rootCmd)
			if err := os.MkdirAll(manDir, os.FileMode(0o755)); err != nil {
				return fault.Wrap(err)
			}
			if err := doc.GenManTree(rootCmd, nil, manDir); err != nil {
				return fault.Wrap(err)
			}
			return nil
		},
	}
)

func init() {
	Command.Flags().StringVar(&manDir, "dir", "man", "directory to store the man page files")
}

func disableAutoGenTag(cmd *cobra.Command) {
	cmd.DisableAutoGenTag = true
	for _, cmd := range cmd.Commands() {
		disableAutoGenTag(cmd)
	}
}
