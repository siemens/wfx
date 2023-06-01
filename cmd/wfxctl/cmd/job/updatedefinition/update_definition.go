package updatedefinition

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"bufio"
	"encoding/json"
	"io"

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
	Use:   "update-definition",
	Short: "Update job definition",
	Long:  `Update definition of an existing job using data provided via stdin`,
	Example: `
wfxctl job update-definition
`,
	TraverseChildren: true,
	Run: func(cmd *cobra.Command, args []string) {
		baseCmd := flags.NewBaseCmd()

		b, err := io.ReadAll(bufio.NewReader(cmd.InOrStdin()))
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to read from stdin")
		}

		var definition map[string]any
		err = json.Unmarshal(b, &definition)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to parse data from stdin")
		}
		client := errutil.Must(baseCmd.CreateHTTPClient())
		params := jobs.NewPutJobsIDDefinitionParams().
			WithHTTPClient(client).
			WithID(flags.Koanf.String(idFlag)).
			WithJobDefinition(definition)

		resp, err := baseCmd.CreateMgmtClient().Jobs.PutJobsIDDefinition(params)
		if err != nil {
			errutil.ProcessErrorResponse(cmd.OutOrStderr(), err)
			log.Fatal().Msg("Failed to update job definition")
		}

		log.Info().Str("id", params.ID).Msg("Updated job definition")
		if err := baseCmd.DumpResponse(cmd.OutOrStdout(), resp.GetPayload()); err != nil {
			log.Fatal().Err(err).Msg("Failed to print response")
		}
	},
}
