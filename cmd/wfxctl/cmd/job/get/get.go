package get

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
	idFlag      = "id"
	historyFlag = "history"
)

func init() {
	f := Command.Flags()
	f.String(idFlag, "", "id of the job to be fetched")
	f.Bool(historyFlag, false, "whether to fetch the job history")
}

var Command = &cobra.Command{
	Use:   "get",
	Short: "Get an existing job",
	Long:  `Get an existing job.`,
	Example: `
wfxctl job get --id=1
`,
	TraverseChildren: true,
	Run: func(cmd *cobra.Command, _ []string) {
		baseCmd := flags.NewBaseCmd()
		history := flags.Koanf.Bool(historyFlag)
		client := errutil.Must(baseCmd.CreateHTTPClient())
		params := jobs.NewGetJobsIDParams().
			WithHTTPClient(client).
			WithID(flags.Koanf.String(idFlag)).
			WithHistory(&history)

		if params.ID == "" {
			log.Fatal().Msg("Job ID missing")
		}

		resp, err := baseCmd.CreateClient().Jobs.GetJobsID(params)
		if err != nil {
			errutil.ProcessErrorResponse(cmd.OutOrStderr(), err)
			log.Fatal().Msg("Failed to get job")
		}
		if err := baseCmd.DumpResponse(cmd.OutOrStdout(), resp.GetPayload()); err != nil {
			log.Fatal().Err(err).Msg("Failed to dump response")
		}
	},
}
