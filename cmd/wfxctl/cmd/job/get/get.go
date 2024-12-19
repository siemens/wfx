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
	"github.com/siemens/wfx/generated/api"
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get an existing job",
		Long:  `Get an existing job.`,
		Example: `
wfxctl job get --id=8ea1e9d7-28e6-4f1f-b444-a8d2d1ad7618
`,
		TraverseChildren: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			baseCmd := flags.NewBaseCmd(cmd.Flags())
			id := baseCmd.ID
			if id == "" {
				return errors.New("id missing")
			}
			history := baseCmd.History
			client := errutil.Must(baseCmd.CreateClient())
			resp, err := client.GetJobsId(cmd.Context(), id, &api.GetJobsIdParams{ParamHistory: &history})
			if err != nil {
				return fault.Wrap(err)
			}
			return fault.Wrap(baseCmd.ProcessResponse(resp, cmd.OutOrStdout()))
		},
	}

	f := cmd.Flags()
	f.String(flags.IDFlag, "", "id of the job to be fetched")
	f.Bool(flags.HistoryFlag, false, "whether to fetch the job history")
	return cmd
}
