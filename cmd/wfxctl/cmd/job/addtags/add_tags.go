package addtags

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
	"github.com/siemens/wfx/generated/api"
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:              "add-tags",
		Short:            "Add tags to a job",
		Example:          "wfxctl job add-tags --id=8ea1e9d7-28e6-4f1f-b444-a8d2d1ad7618 tag1 tag2",
		TraverseChildren: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			baseCmd := flags.NewBaseCmd(cmd.Flags())
			tags := args

			client := errutil.Must(baseCmd.CreateMgmtClient())
			resp, err := client.PostJobsIdTags(cmd.Context(), baseCmd.ID, nil, api.PostJobsIdTagsJSONRequestBody(tags))
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
