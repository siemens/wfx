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
	"github.com/siemens/wfx/generated/client/jobs"
)

const (
	clientIDFlag = "client-id"
	stateFlag    = "state"
	workflowFlag = "workflow"
	sortFlag     = "sort"
	groupFlag    = "group"
	offsetFlag   = "offset"
	limitFlag    = "limit"
	tagFlag      = "tag"
)

func init() {
	f := Command.Flags()
	f.String(clientIDFlag, "", "Filter jobs belonging to a specific client with clientId")
	f.StringSlice(groupFlag, []string{}, "Filter jobs based on the group they belong to")
	f.String(stateFlag, "", "Filter jobs based on the current state value")
	f.String(workflowFlag, "", "Filter jobs based on workflow name")
	f.StringSlice(tagFlag, []string{}, "Filter jobs by tags")
	f.Int32(offsetFlag, 0, "0-based index of the page")
	f.Int32(limitFlag, 10, "maximum number of elements returned in one page ")
	f.String(sortFlag, "", "sort order. possible values: asc, desc")
}

var Command = &cobra.Command{
	Use:   "query",
	Short: "Query existing jobs",
	Long:  `Query existing jobs`,
	Example: `
wfxctl job query --state=CREATED
`,
	TraverseChildren: true,
	Run: func(cmd *cobra.Command, args []string) {
		baseCmd := flags.NewBaseCmd()
		client := errutil.Must(baseCmd.CreateHTTPClient())
		params := jobs.NewGetJobsParams().
			WithHTTPClient(client)

		if clientID := flags.Koanf.String(clientIDFlag); clientID != "" {
			params = params.WithClientID(&clientID)
		}
		if state := flags.Koanf.String(stateFlag); state != "" {
			params = params.WithState(&state)
		}
		if modelKey := flags.Koanf.String(workflowFlag); modelKey != "" {
			params = params.WithWorkflow(&modelKey)
		}
		if sort := flags.Koanf.String(sortFlag); sort != "" {
			params = params.WithSort(&sort)
		}
		groups := flags.Koanf.Strings(groupFlag)
		tags := flags.Koanf.Strings(tagFlag)

		offset := flags.Koanf.Int64(offsetFlag)
		limit := int32(flags.Koanf.Int(limitFlag))
		params = params.
			WithOffset(&offset).
			WithLimit(&limit).
			WithTag(tags).
			WithGroup(groups)

		resp, err := baseCmd.CreateClient().Jobs.GetJobs(params)
		if err != nil {
			errutil.ProcessErrorResponse(cmd.OutOrStderr(), err)
		} else {
			if err := baseCmd.DumpResponse(cmd.OutOrStdout(), resp.GetPayload()); err != nil {
				log.Fatal().Err(err).Msg("Failed to dump response")
			}
		}
	},
}
