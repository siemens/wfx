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
	"github.com/siemens/wfx/generated/client/workflows"
)

const (
	nameFlag = "name"
)

func init() {
	f := Command.PersistentFlags()
	f.String(nameFlag, "", "workflow name")
}

var Command = &cobra.Command{
	Use:              "delete",
	Short:            "Delete an existing workflow",
	Long:             `Delete an existing workflow`,
	TraverseChildren: true,
	Example:          "wfxctl workflow delete --name=wfx.workflow.kanban",
	Run: func(cmd *cobra.Command, args []string) {
		baseCmd := flags.NewBaseCmd()
		client := errutil.Must(baseCmd.CreateHTTPClient())
		params := workflows.NewDeleteWorkflowsNameParams().
			WithHTTPClient(client).
			WithName(flags.Koanf.String(nameFlag))

		// no content
		if _, err := baseCmd.CreateMgmtClient().Workflows.DeleteWorkflowsName(params); err != nil {
			errutil.ProcessErrorResponse(cmd.OutOrStderr(), err)
			log.Fatal().Msg("Failed to delete workflow")
		}
	},
}
