package query

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
		Use:   "query",
		Short: "Query existing jobs",
		Long:  `Query existing jobs`,
		Example: `
wfxctl job query --state=CREATED
`,
		TraverseChildren: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			baseCmd := flags.NewBaseCmd(cmd.Flags())

			params := new(api.GetJobsParams)
			if clientID := baseCmd.ClientID; clientID != "" {
				params.ParamClientID = &clientID
			}
			if state := baseCmd.State; state != "" {
				params.ParamState = &state
			}
			if workflow := baseCmd.Workflow; workflow != "" {
				params.ParamWorkflow = &workflow
			}
			{
				var err error
				params.ParamSort, err = baseCmd.SortParam()
				if err != nil {
					return fault.Wrap(err)
				}
			}
			if groups := baseCmd.Groups; len(groups) > 0 {
				params.ParamGroup = &groups
			}
			if tags := baseCmd.Tags; len(tags) > 0 {
				params.ParamTag = &tags
			}

			params.ParamOffset = &baseCmd.Offset
			params.ParamLimit = &baseCmd.Limit

			client := errutil.Must(baseCmd.CreateClient())
			resp, err := client.GetJobs(cmd.Context(), params)
			if err != nil {
				return fault.Wrap(err)
			}
			return fault.Wrap(baseCmd.ProcessResponse(resp, cmd.OutOrStdout()))
		},
	}
	f := cmd.Flags()
	f.String(flags.ClientIDFlag, "", "Filter jobs belonging to a specific client with clientId")
	f.StringSlice(flags.GroupFlag, []string{}, "Filter jobs based on the group they belong to")
	f.String(flags.StateFlag, "", "Filter jobs based on the current state value")
	f.String(flags.WorkflowFlag, "", "Filter jobs based on workflow name")
	f.StringSlice(flags.TagFlag, []string{}, "Filter jobs by tags")
	f.Int64(flags.OffsetFlag, 0, "0-based index of the page")
	f.Int32(flags.LimitFlag, 10, "maximum number of elements returned in one page ")
	f.String(flags.SortFlag, "", "sort order. possible values: asc, desc")
	return cmd
}
