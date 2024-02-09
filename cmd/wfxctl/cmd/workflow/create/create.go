package create

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/Southclaws/fault"
	"github.com/go-openapi/strfmt"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/siemens/wfx/cmd/wfxctl/errutil"
	"github.com/siemens/wfx/cmd/wfxctl/flags"
	"github.com/siemens/wfx/generated/client/workflows"
	"github.com/siemens/wfx/generated/model"
)

const kanbanExample = `
name: wfx.workflow.kanban

states:
  - name: BACKLOG
  - name: NEW
  - name: PROGRESS
  - name: DONE
  - name: CANCELED

transitions:
  - from: BACKLOG
    to: NEW
    eligible: WFX
  - from: BACKLOG
    to: CANCELED
    eligible: WFX
  - from: NEW
    to: PROGRESS
    eligible: CLIENT
  - from: PROGRESS
    to: DONE
    eligible: CLIENT
`

var Command = &cobra.Command{
	Use:   "create",
	Short: "Create a new workflow",
	Long:  `Create a new workflow. The workflow must be in YAML format.`,
	Example: fmt.Sprintf(`
cat <<EOF | wfxctl workflow create -
%s
EOF
`, kanbanExample),
	TraverseChildren: true,
	Args:             cobra.OnlyValidArgs,
	ValidArgsFunction: func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
		return []string{"yaml", "yml"}, cobra.ShellCompDirectiveFilterFileExt
	},
	Run: func(cmd *cobra.Command, args []string) {
		allWorkflows, err := readWorkflows(args, cmd.InOrStdin())
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to read workflows")
		}
		log.Info().Int("count", len(allWorkflows)).Msg("Creating workflows")
		for _, wf := range allWorkflows {
			log.Debug().Msg("Validating workflow")
			if err := wf.Validate(strfmt.Default); err != nil {
				log.Fatal().Msg(err.Error())
			}

			baseCmd := flags.NewBaseCmd()
			client := errutil.Must(baseCmd.CreateHTTPClient())
			params := workflows.NewPostWorkflowsParams().
				WithHTTPClient(client).
				WithWorkflow(wf)

			resp, err := baseCmd.CreateMgmtClient().Workflows.PostWorkflows(params)
			if err != nil {
				errutil.ProcessErrorResponse(cmd.OutOrStderr(), err)
				log.Fatal().Msg("Failed to create workflow")
			}

			log.Info().Str("name", wf.Name).Msg("Created new workflow")
			if err := baseCmd.DumpResponse(cmd.OutOrStdout(), resp.GetPayload()); err != nil {
				log.Fatal().Err(err).Msg("Failed to dump response")
			}
		}
	},
}

func readWorkflows(args []string, r io.Reader) ([]*model.Workflow, error) {
	n := len(args)
	if n == 0 {
		return nil, errors.New("workflow must ge given either via file or stdin")
	}
	allWorkflows := make([]*model.Workflow, 0, len(args))
	if n == 1 && args[0] == "-" {
		log.Debug().Msg("Reading workflow from stdin...")
		b, err := io.ReadAll(r)
		if err != nil {
			return nil, fault.Wrap(err)
		}
		wf, err := unmarshal(b)
		if err != nil {
			return nil, fault.Wrap(err)
		}
		allWorkflows = append(allWorkflows, wf)
	} else {
		for _, fname := range args {
			b, err := os.ReadFile(fname)
			if err != nil {
				return nil, fault.Wrap(err)
			}
			wf, err := unmarshal(b)
			if err != nil {
				return nil, fault.Wrap(err)
			}
			allWorkflows = append(allWorkflows, wf)
		}
	}
	return allWorkflows, nil
}

func unmarshal(raw []byte) (*model.Workflow, error) {
	// try YAML
	var wf model.Workflow
	err := yaml.Unmarshal(raw, &wf)
	if err != nil {
		return nil, fault.Wrap(err)
	}
	return &wf, nil
}
