package create

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"encoding/json"
	"io"
	"os"

	"github.com/Southclaws/fault"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/siemens/wfx/cmd/wfxctl/errutil"
	"github.com/siemens/wfx/cmd/wfxctl/flags"
	"github.com/siemens/wfx/generated/api"
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new job",
		Long:  "Create a new job. You should provide the job definition (JSON) via stdin.",
		Example: `
echo '{ "title": "Task 1" }' | wfxctl job create --client-id=my_client --workflow=wfx.workflow.kanban -
	`,
		TraverseChildren: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			baseCmd := flags.NewBaseCmd(cmd.Flags())
			clientID := baseCmd.ClientID
			workflow := baseCmd.Workflow
			tags := baseCmd.Tags
			log.Debug().
				Str("clientID", clientID).
				Str("workflow", workflow).
				Strs("tags", tags).
				Msg("Creating new job")

			request := api.PostJobsJSONRequestBody{
				ClientID:   clientID,
				Workflow:   workflow,
				Tags:       tags,
				Definition: make(map[string]any),
			}

			n := len(args)
			switch n {
			case 0:
				log.Warn().Msg("No job definition supplied!")
			case 1:
				var b []byte
				var err error
				if args[0] == "-" {
					// stdin is a pipe, so we shall read from it and attach it to the Data field
					log.Debug().Msg("Reading job definition from stdin...")
					b, err = io.ReadAll(cmd.InOrStdin())
				} else {
					b, err = os.ReadFile(args[0])
				}
				if err != nil {
					return fault.Wrap(err)
				}
				if err := json.Unmarshal(b, &request.Definition); err != nil {
					return fault.Wrap(err)
				}
				log.Debug().RawJSON("definition", b).Msg("Parsed job definition")
			default:
				return errors.New("Too many arguments")
			}

			client := errutil.Must(baseCmd.CreateMgmtClient())
			resp, err := client.PostJobs(cmd.Context(), nil, request)
			if err != nil {
				return fault.Wrap(err)
			}
			return fault.Wrap(baseCmd.ProcessResponse(resp, cmd.OutOrStdout()))
		},
	}
	f := cmd.Flags()
	f.String(flags.ClientIDFlag, "", "clientID for the job")
	f.String(flags.WorkflowFlag, "", "workflow for the job")
	f.StringArray(flags.TagFlag, []string{}, "Tags to apply to the job")
	return cmd
}
