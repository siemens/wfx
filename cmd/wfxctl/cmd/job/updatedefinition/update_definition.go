package updatedefinition

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"bufio"
	"errors"

	"github.com/Southclaws/fault"
	"github.com/spf13/cobra"

	"github.com/siemens/wfx/cmd/wfxctl/errutil"
	"github.com/siemens/wfx/cmd/wfxctl/flags"
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-definition",
		Short: "Update job definition",
		Long:  `Update definition of an existing job using data provided via stdin`,
		Example: `
wfxctl job update-definition
`,
		TraverseChildren: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			baseCmd := flags.NewBaseCmd(cmd.Flags())
			id := baseCmd.ID
			if id == "" {
				return errors.New("job id missing")
			}
			client := errutil.Must(baseCmd.CreateMgmtClient())
			resp, err := client.PutJobsIdDefinitionWithBody(cmd.Context(), id, nil, "application/json", bufio.NewReader(cmd.InOrStdin()))
			if err != nil {
				return fault.Wrap(err)
			}
			return fault.Wrap(baseCmd.ProcessResponse(resp, cmd.OutOrStdout()))
		},
	}

	f := cmd.Flags()
	f.String(flags.IDFlag, "", "job id")
	return cmd
}
