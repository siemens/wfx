package query

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/siemens/wfx/cmd/wfxctl/errutil"
	"github.com/siemens/wfx/cmd/wfxctl/flags"
	"github.com/siemens/wfx/generated/client/workflows"
)

const (
	offsetFlag = "offset"
	limitFlag  = "limit"
)

func init() {
	flags := Command.PersistentFlags()

	flags.Int32(offsetFlag, 0, "the number of items to skip before starting to return results")
	flags.Int32(limitFlag, 10, "the maximum number of items to return")
}

var Command = &cobra.Command{
	Use:   "query",
	Short: "Query existing workflows",
	Long:  `Query existing workflows`,
	Example: `
wfxctl workflow query
`,
	TraverseChildren: true,
	Run: func(cmd *cobra.Command, args []string) {
		baseCmd := flags.NewBaseCmd()

		offset := flags.Koanf.Int64(offsetFlag)
		limit := int32(flags.Koanf.Int(limitFlag))

		client := errutil.Must(baseCmd.CreateHTTPClient())
		params := workflows.NewGetWorkflowsParams().
			WithHTTPClient(client).
			WithOffset(&offset).
			WithLimit(&limit)

		resp, err := baseCmd.CreateClient().Workflows.GetWorkflows(params)
		if err != nil {
			errutil.ProcessErrorResponse(cmd.OutOrStderr(), err)
			log.Fatal().Msg("Failed to query workflows")
		}
		if err := baseCmd.DumpResponse(cmd.OutOrStdout(), resp.GetPayload()); err != nil {
			log.Fatal().Err(err).Msg("Failed to dump response")
		}
	},
}
