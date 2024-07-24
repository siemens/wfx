package workflow

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"github.com/siemens/wfx/cmd/wfxctl/cmd/workflow/create"
	"github.com/siemens/wfx/cmd/wfxctl/cmd/workflow/delete"
	"github.com/siemens/wfx/cmd/wfxctl/cmd/workflow/get"
	"github.com/siemens/wfx/cmd/wfxctl/cmd/workflow/query"
	"github.com/siemens/wfx/cmd/wfxctl/cmd/workflow/validate"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:              "workflow",
		Short:            "manage workflows",
		Long:             "subcommand to manage workflows (CRUD)",
		TraverseChildren: true,
		SilenceUsage:     true,
	}
	cmd.AddCommand(create.NewCommand())
	cmd.AddCommand(delete.NewCommand())
	cmd.AddCommand(get.NewCommand())
	cmd.AddCommand(query.NewCommand())
	cmd.AddCommand(validate.NewCommand())
	return cmd
}
