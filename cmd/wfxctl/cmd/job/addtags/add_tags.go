package addtags

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
	Use:              "add-tags",
	Short:            "Add tags to a job",
	Example:          "wfxctl add-tags --id=1 tag1 tag2",
	TraverseChildren: true,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			log.Fatal().Msg("No tags were provided")
		}

		baseCmd := flags.NewBaseCmd()
		client := errutil.Must(baseCmd.CreateHTTPClient())
		params := jobs.NewPostJobsIDTagsParams().
			WithHTTPClient(client).
			WithID(flags.Koanf.String(idFlag)).
			WithTags(args)

		resp := errutil.Must(baseCmd.CreateMgmtClient().Jobs.PostJobsIDTags(params))
		fmt.Fprintln(cmd.OutOrStdout(), resp.Payload)
	},
}
