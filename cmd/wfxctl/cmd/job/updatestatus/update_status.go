package updatestatus

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/siemens/wfx/cmd/wfxctl/errutil"
	"github.com/siemens/wfx/cmd/wfxctl/flags"
	"github.com/siemens/wfx/generated/client"
	"github.com/siemens/wfx/generated/client/jobs"
	"github.com/siemens/wfx/generated/model"
)

const (
	idFlag       = "id"
	actorFlag    = "actor"
	clientIDFlag = "client-id"
	stateFlag    = "state"
	progressFlag = "progress"
	messageFlag  = "message"
)

func init() {
	f := Command.Flags()
	f.String(idFlag, "", "job which shall be updated")
	f.String(actorFlag, string(model.EligibleEnumCLIENT), "actor to use (eligible)")
	f.String(clientIDFlag, "", "client which sends the update")
	f.String(stateFlag, "", "Name of the new state")
	f.Int(progressFlag, 0, "progress value (0 <= progress <= 100)")
	f.String(messageFlag, "", "status message / info, free text from client")
}

var Command = &cobra.Command{
	Use:   "update-status",
	Short: "Update job status",
	Long:  `Update job status of an existing job`,
	Example: `
wfxctl job update-status --id=1 --client-id=client42 --state=DOWNLOAD
`,
	Run: func(cmd *cobra.Command, args []string) {
		baseCmd := flags.NewBaseCmd()
		req := model.JobStatus{
			ClientID: flags.Koanf.String(clientIDFlag),
			State:    flags.Koanf.String(stateFlag),
			Progress: int32(flags.Koanf.Int(progressFlag)),
			Message:  flags.Koanf.String(messageFlag),
		}

		cli := errutil.Must(baseCmd.CreateHTTPClient())

		params := jobs.NewPutJobsIDStatusParams().
			WithHTTPClient(cli).
			WithID(flags.Koanf.String(idFlag)).
			WithNewJobStatus(&req)

		var c *client.WorkflowExecutor
		if strings.ToUpper(flags.Koanf.String(actorFlag)) == string(model.EligibleEnumCLIENT) {
			c = baseCmd.CreateClient()
		} else {
			c = baseCmd.CreateMgmtClient()
		}
		resp, err := c.Jobs.PutJobsIDStatus(params)
		if err != nil {
			errutil.ProcessErrorResponse(cmd.OutOrStderr(), err)
			log.Fatal().Msg("Failed to update job")
		}

		log.Info().Str("id", params.ID).Str("state", req.State).Msg("Updated job status")
		if err := baseCmd.DumpResponse(cmd.OutOrStdout(), resp.GetPayload()); err != nil {
			log.Fatal().Err(err).Msg("Failed to dump response")
		}
	},
}
