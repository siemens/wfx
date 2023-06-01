package deltags

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"fmt"

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
	Use:              "del-tags",
	Short:            "Delete tags from a job",
	Example:          "wfxctl del-tags --id=1 tag1 tag2",
	TraverseChildren: true,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			log.Fatal().Msg("No tags were provided")
		}

		baseCmd := flags.NewBaseCmd()
		client := errutil.Must(baseCmd.CreateHTTPClient())
		params := jobs.NewDeleteJobsIDTagsParams().
			WithHTTPClient(client).
			WithID(flags.Koanf.String(idFlag)).
			WithTags(args)

		resp, err := baseCmd.CreateMgmtClient().Jobs.DeleteJobsIDTags(params)
		if err != nil {
			errutil.ProcessErrorResponse(cmd.OutOrStderr(), err)
			log.Fatal().Msg("Failed to delete tags")
		}
		fmt.Fprintln(cmd.OutOrStdout(), resp.Payload)
	},
}
