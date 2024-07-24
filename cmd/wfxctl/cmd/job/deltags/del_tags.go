package deltags

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
	"github.com/siemens/wfx/generated/api"
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:              "del-tags",
		Short:            "Delete tags from a job",
		Example:          "wfxctl job del-tags --id=8ea1e9d7-28e6-4f1f-b444-a8d2d1ad7618 tag1 tag2",
		TraverseChildren: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("no tags provided")
			}
			baseCmd := flags.NewBaseCmd(cmd.Flags())
			tags := args

			client := errutil.Must(baseCmd.CreateMgmtClient())
			resp, err := client.DeleteJobsIdTags(cmd.Context(), baseCmd.ID, nil, api.DeleteJobsIdTagsJSONRequestBody(tags))
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
