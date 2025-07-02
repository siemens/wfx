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
	"github.com/goccy/go-yaml"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/siemens/wfx/cmd/wfxctl/errutil"
	"github.com/siemens/wfx/cmd/wfxctl/flags"
	"github.com/siemens/wfx/generated/api"
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

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
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
		RunE: func(cmd *cobra.Command, args []string) error {
			baseCmd := flags.NewBaseCmd(cmd.Flags())

			allWorkflows, err := readWorkflows(args, cmd.InOrStdin())
			if err != nil {
				return fault.Wrap(err)
			}
			log.Info().Int("count", len(allWorkflows)).Msg("Creating workflows")
			client := errutil.Must(baseCmd.CreateMgmtClient())
			for _, wf := range allWorkflows {
				resp, err := client.PostWorkflows(cmd.Context(), nil, api.PostWorkflowsJSONRequestBody(wf))
				if err != nil {
					return fault.Wrap(err)
				}
				if err := baseCmd.ProcessResponse(resp, cmd.OutOrStdout()); err != nil {
					return fault.Wrap(err)
				}
			}
			return nil
		},
	}
	return cmd
}

func readWorkflows(args []string, r io.Reader) ([]api.Workflow, error) {
	n := len(args)
	if n == 0 {
		return nil, errors.New("workflow must ge given either via file or stdin")
	}
	allWorkflows := make([]api.Workflow, 0, len(args))
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
		allWorkflows = append(allWorkflows, *wf)
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
			allWorkflows = append(allWorkflows, *wf)
		}
	}
	return allWorkflows, nil
}

func unmarshal(raw []byte) (*api.Workflow, error) {
	// try YAML
	var wf api.Workflow
	err := yaml.Unmarshal(raw, &wf)
	if err != nil {
		return nil, fault.Wrap(err)
	}
	return &wf, nil
}
