package version

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"encoding/json"
	"fmt"

	"github.com/Southclaws/fault"
	"github.com/spf13/cobra"

	"github.com/siemens/wfx/cmd/wfxctl/errutil"
	"github.com/siemens/wfx/cmd/wfxctl/flags"
	"github.com/siemens/wfx/generated/api"
)

func NewCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Retrieve version of wfx",
		Long:  "Retrieve version of wfx",
		RunE: func(cmd *cobra.Command, _ []string) error {
			baseCmd := flags.NewBaseCmd(cmd.Flags())
			client := errutil.Must(baseCmd.CreateClient())
			resp, err := client.GetVersion(cmd.Context())
			if err != nil {
				return fault.Wrap(err)
			}
			var versionResp api.GetVersion200JSONResponse
			if err := json.NewDecoder(resp.Body).Decode(&versionResp); err != nil {
				return fault.Wrap(err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), `
   version: %s
    commit: %s
apiVersion: %s
`, versionResp.Version, versionResp.Commit, versionResp.ApiVersion)
			return nil
		},
	}
}
