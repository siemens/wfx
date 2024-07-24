package updatestatus

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"errors"
	"strings"

	"github.com/Southclaws/fault"
	"github.com/spf13/cobra"

	"github.com/siemens/wfx/cmd/wfxctl/errutil"
	"github.com/siemens/wfx/cmd/wfxctl/flags"
	"github.com/siemens/wfx/generated/api"
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-status",
		Short: "Update job status",
		Long:  `Update job status of an existing job`,
		Example: `
wfxctl job update-status --id=8ea1e9d7-28e6-4f1f-b444-a8d2d1ad7618 --client-id=client42 --state=DOWNLOAD
`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			baseCmd := flags.NewBaseCmd(cmd.Flags())
			id := baseCmd.ID
			if id == "" {
				return errors.New("id missing")
			}
			clientID := baseCmd.ClientID
			progress := int32(baseCmd.Progress)
			message := baseCmd.Message
			status := api.JobStatus{
				ClientID: clientID,
				State:    baseCmd.State,
				Progress: &progress,
				Message:  message,
			}

			var client *api.Client
			if strings.ToUpper(baseCmd.Actor) == string(api.CLIENT) {
				client = errutil.Must(baseCmd.CreateClient())
			} else {
				client = errutil.Must(baseCmd.CreateMgmtClient())
			}
			resp, err := client.PutJobsIdStatus(cmd.Context(), id, nil, api.PutJobsIdStatusJSONRequestBody(status))
			if err != nil {
				return fault.Wrap(err)
			}
			return fault.Wrap(baseCmd.ProcessResponse(resp, cmd.OutOrStdout()))
		},
	}
	f := cmd.Flags()
	f.String(flags.IDFlag, "", "job which shall be updated")
	f.String(flags.ActorFlag, string(api.CLIENT), "actor to use (eligible)")
	f.String(flags.ClientIDFlag, "", "client which sends the update")
	f.String(flags.StateFlag, "", "name of the new state")
	f.Int(flags.ProgressFlag, 0, "progress value (0 <= progress <= 100)")
	f.String(flags.MessageFlag, "", "status message / info, free text from client")
	return cmd
}
