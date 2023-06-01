package delete

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
	f.String(idFlag, "", "id of the job which shall be deleted")
}

var Command = &cobra.Command{
	Use:              "delete",
	Short:            "Delete an existing job",
	Long:             `Delete an existing job`,
	TraverseChildren: true,
	Run: func(cmd *cobra.Command, args []string) {
		baseCmd := flags.NewBaseCmd()
		client := errutil.Must(baseCmd.CreateHTTPClient())
		params := jobs.NewDeleteJobsIDParams().
			WithHTTPClient(client).
			WithID(flags.Koanf.String(idFlag))

		// no content
		if _, err := baseCmd.CreateMgmtClient().Jobs.DeleteJobsID(params); err != nil {
			errutil.ProcessErrorResponse(cmd.OutOrStderr(), err)
			log.Fatal().Msg("Failed to delete job")
		}
	},
}
