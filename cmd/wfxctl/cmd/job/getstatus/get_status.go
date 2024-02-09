package getstatus

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
	idFlag = "id"
)

func init() {
	f := Command.Flags()
	f.String(idFlag, "", "job id")
}

var Command = &cobra.Command{
	Use:   "get-status",
	Short: "Get status of an existing job",
	Long:  "Get status of an existing job",
	Example: `
wfxctl job get-status --id=1
`,
	TraverseChildren: true,
	Run: func(cmd *cobra.Command, _ []string) {
		baseCmd := flags.NewBaseCmd()
		client := errutil.Must(baseCmd.CreateHTTPClient())
		params := jobs.NewGetJobsIDStatusParams().
			WithHTTPClient(client).
			WithID(flags.Koanf.String(idFlag))

		if params.ID == "" {
			log.Fatal().Msg("Job ID missing")
		}

		resp, err := baseCmd.CreateClient().Jobs.GetJobsIDStatus(params)
		if err != nil {
			errutil.ProcessErrorResponse(cmd.OutOrStderr(), err)
			log.Fatal().Msg("Failed to get job status")
		}
		if err := baseCmd.DumpResponse(cmd.OutOrStdout(), resp.GetPayload()); err != nil {
			log.Fatal().Err(err).Msg("Failed to get job status")
		}
	},
}
