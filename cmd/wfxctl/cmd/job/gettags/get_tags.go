package gettags

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
	Use:              "get-tags",
	Short:            "Get tags of a job",
	Example:          "wfxctl get-tags --id=1",
	TraverseChildren: true,
	Run: func(cmd *cobra.Command, args []string) {
		baseCmd := flags.NewBaseCmd()
		client := errutil.Must(baseCmd.CreateHTTPClient())
		params := jobs.NewGetJobsIDTagsParams().
			WithHTTPClient(client).
			WithID(flags.Koanf.String(idFlag))

		resp, err := baseCmd.CreateClient().Jobs.GetJobsIDTags(params)
		if err != nil {
			errutil.ProcessErrorResponse(cmd.OutOrStderr(), err)
			log.Fatal().Msg("Failed to retrieve tags")
		}
		fmt.Fprintln(cmd.OutOrStdout(), resp.Payload)
	},
}
