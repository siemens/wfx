package gettags

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
	cmd := &cobra.Command{
		Use:              "get-tags",
		Short:            "Get tags of a job",
		Example:          "wfxctl job get-tags --id=8ea1e9d7-28e6-4f1f-b444-a8d2d1ad7618",
		TraverseChildren: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			baseCmd := flags.NewBaseCmd(cmd.Flags())
			id := baseCmd.ID
			client := errutil.Must(baseCmd.CreateClient())
			resp, err := client.GetJobsIdTags(cmd.Context(), id, nil)
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
