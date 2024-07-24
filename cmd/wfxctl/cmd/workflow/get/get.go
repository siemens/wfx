package get

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"errors"

	"github.com/Southclaws/fault"
	"github.com/spf13/cobra"

	"github.com/siemens/wfx/cmd/wfxctl/errutil"
	"github.com/siemens/wfx/cmd/wfxctl/flags"
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:              "get",
		Short:            "Get an existing workflow",
		Long:             `Get an existing workflow`,
		Example:          "wfxctl workflow get --name=wfx.workflow.kanban",
		TraverseChildren: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			baseCmd := flags.NewBaseCmd(cmd.Flags())
			name := baseCmd.Name
			if name == "" {
				return errors.New("workflow name missing")
			}
			client := errutil.Must(baseCmd.CreateClient())
			resp, err := client.GetWorkflowsName(cmd.Context(), name, nil)
			if err != nil {
				return fault.Wrap(err)
			}
			return fault.Wrap(baseCmd.ProcessResponse(resp, cmd.OutOrStdout()))
		},
	}
	f := cmd.PersistentFlags()
	f.String(flags.NameFlag, "", "workflow name")
	return cmd
}
