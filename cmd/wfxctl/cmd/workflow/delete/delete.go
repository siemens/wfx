package delete

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"github.com/Southclaws/fault"
	"github.com/spf13/cobra"

	"github.com/siemens/wfx/cmd/wfxctl/errutil"
	"github.com/siemens/wfx/cmd/wfxctl/flags"
)

func NewCommand() *cobra.Command {
	return &cobra.Command{
		Use:              "delete",
		Short:            "Delete an existing workflow",
		TraverseChildren: true,
		Example:          "wfxctl workflow delete wfx.workflow.kanban",
		RunE: func(cmd *cobra.Command, args []string) error {
			baseCmd := flags.NewBaseCmd(cmd.Flags())
			client := errutil.Must(baseCmd.CreateMgmtClient())
			for _, name := range args {
				resp, err := client.DeleteWorkflowsName(cmd.Context(), name)
				if err != nil {
					return fault.Wrap(err)
				}
				if err := baseCmd.ProcessResponse(resp, cmd.OutOrStdout()); err != nil {
					return fault.Wrap(err)
				}
			}
			return nil
		},
	}
}
