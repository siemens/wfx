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

	"github.com/go-openapi/strfmt"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/siemens/wfx/cmd/wfxctl/errutil"
	"github.com/siemens/wfx/cmd/wfxctl/flags"
	"github.com/siemens/wfx/generated/client/jobs"
	"github.com/siemens/wfx/generated/model"
)

const (
	clientIDFlag = ("client-id")
	workflowFlag = ("workflow")
	tagFlag      = ("tag")
)

func init() {
	f := Command.Flags()
	f.String(clientIDFlag, "", "clientID for the job")
	f.String(workflowFlag, "", "workflow for the job")
	f.StringArray(tagFlag, []string{}, "Tags to apply to the job")
}

var Command = &cobra.Command{
	Use:   "create",
	Short: "Create a new job",
	Long:  "Create a new job. You should provide the job definition (JSON) via stdin.",
	Example: `
echo '{ "title": "Task 1" }' | wfxctl job create --client-id=my_client --workflow=wfx.workflow.kanban -
	`,
	TraverseChildren: true,
	Run: func(cmd *cobra.Command, args []string) {
		clientID := flags.Koanf.String(clientIDFlag)
		workflow := flags.Koanf.String(workflowFlag)
		tags := flags.Koanf.Strings(tagFlag)
		log.Debug().
			Str("clientID", clientID).
			Str("workflow", workflow).
			Strs("tags", tags).
			Msg("Creating new job")

		baseCmd := flags.NewBaseCmd()
		client := errutil.Must(baseCmd.CreateHTTPClient())
		params := jobs.NewPostJobsParams().
			WithHTTPClient(client).
			WithJob(&model.JobRequest{
				ClientID:   clientID,
				Workflow:   workflow,
				Tags:       tags,
				Definition: map[string]any{},
			})

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
				log.Fatal().Err(err).Msg("Failed to read from stdin")
			}
			if err := json.Unmarshal(b, &params.Job.Definition); err != nil {
				log.Fatal().Err(err).Msg("Failed to unmarshal JSON")
			}
			log.Debug().RawJSON("definition", b).Msg("Parsed job definition")
		default:
			log.Fatal().Int("n", n).Msg("Too many arguments")
		}

		if err := params.Job.Validate(strfmt.Default); err != nil {
			log.Fatal().Msg(err.Error())
		}

		resp, err := baseCmd.CreateMgmtClient().Jobs.PostJobs(params)
		if err != nil {
			errutil.ProcessErrorResponse(cmd.OutOrStderr(), err)
			log.Fatal().Msg("Failed to create job")
		}
		job := resp.GetPayload()
		var state string
		if job.Status != nil {
			state = job.Status.State
		}
		log.Info().Str("id", job.ID).Str("state", state).Msg("Created new job")
		if err := baseCmd.DumpResponse(cmd.OutOrStdout(), resp.GetPayload()); err != nil {
			log.Fatal().Err(err).Msg("Failed to print response")
		}
	},
}
