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
		Short: "Query existing workflows",
		Long:  `Query existing workflows`,
		Example: `
wfxctl workflow query
`,
		TraverseChildren: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			baseCmd := flags.NewBaseCmd(cmd.Flags())

			params := new(api.GetWorkflowsParams)
			params.ParamOffset = &baseCmd.Offset
			params.ParamLimit = &baseCmd.Limit

			{
				var err error
				params.ParamSort, err = baseCmd.SortParam()
				if err != nil {
					return fault.Wrap(err)
				}
			}

			client := errutil.Must(baseCmd.CreateClient())
			resp, err := client.GetWorkflows(cmd.Context(), params)
			if err != nil {
				return fault.Wrap(err)
			}
			return fault.Wrap(baseCmd.ProcessResponse(resp, cmd.OutOrStdout()))
		},
	}
	f := cmd.PersistentFlags()
	f.Int64(flags.OffsetFlag, 0, "the number of items to skip before starting to return results")
	f.Int32(flags.LimitFlag, 10, "the maximum number of items to return")
	f.String(flags.SortFlag, "", "sort order. possible values: asc, desc")
	return cmd
}
