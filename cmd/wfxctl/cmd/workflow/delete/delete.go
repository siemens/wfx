package delete

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/siemens/wfx/cmd/wfxctl/errutil"
	"github.com/siemens/wfx/cmd/wfxctl/flags"
	"github.com/siemens/wfx/generated/client/workflows"
)

var Command = &cobra.Command{
	Use:              "delete",
	Short:            "Delete an existing workflow",
	TraverseChildren: true,
	Example:          "wfxctl workflow delete wfx.workflow.kanban",
	RunE: func(cmd *cobra.Command, args []string) error {
		baseCmd := flags.NewBaseCmd()
		client := errutil.Must(baseCmd.CreateHTTPClient())
		failedWfs := make([]string, 0)
		for _, workflow := range args {
			params := workflows.NewDeleteWorkflowsNameParams().WithHTTPClient(client).WithName(workflow)
			// no content
			if _, err := baseCmd.CreateMgmtClient().Workflows.DeleteWorkflowsName(params); err != nil {
				errutil.ProcessErrorResponse(cmd.OutOrStderr(), err)
				failedWfs = append(failedWfs, workflow)
			} else {
				log.Info().Str("workflow", workflow).Msg("Deleted workflow")
			}
		}
		if len(failedWfs) > 0 {
			return fmt.Errorf("failed to delete workflows: %s", strings.Join(failedWfs, ","))
		}
		return nil
	},
}
