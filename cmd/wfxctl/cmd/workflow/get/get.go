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
	"github.com/siemens/wfx/generated/client/workflows"
)

const (
	nameFlag = ("name")
)

func init() {
	f := Command.PersistentFlags()
	f.String(nameFlag, "", "workflow name")
}

var Command = &cobra.Command{
	Use:              "get",
	Short:            "Get an existing workflow",
	Long:             `Get an existing workflow`,
	Example:          "wfxctl workflow get --name=wfx.workflow.kanban",
	TraverseChildren: true,
	Run: func(cmd *cobra.Command, _ []string) {
		baseCmd := flags.NewBaseCmd()
		client := errutil.Must(baseCmd.CreateHTTPClient())
		params := workflows.NewGetWorkflowsNameParams().
			WithHTTPClient(client).
			WithName(flags.Koanf.String(nameFlag))

		if params.Name == "" {
			log.Fatal().Msg("No workflow name provided")
		}

		resp, err := baseCmd.CreateMgmtClient().Workflows.GetWorkflowsName(params)
		if err != nil {
			errutil.ProcessErrorResponse(cmd.OutOrStderr(), err)
			log.Fatal().Msg("Failed to get workflow")
		}
		if err := baseCmd.DumpResponse(cmd.OutOrStdout(), resp.GetPayload()); err != nil {
			log.Fatal().Err(err).Msg("Failed to dump response")
		}
	},
}
