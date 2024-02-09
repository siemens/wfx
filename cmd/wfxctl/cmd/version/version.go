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
	"io"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/siemens/wfx/cmd/wfxctl/errutil"
	"github.com/siemens/wfx/cmd/wfxctl/flags"
	"github.com/siemens/wfx/middleware/version"
)

var Command = &cobra.Command{
	Use:   "version",
	Short: "Retrieve version of wfx",
	Long:  "Retrieve version of wfx",
	RunE: func(cmd *cobra.Command, _ []string) error {
		baseCmd := flags.NewBaseCmd()
		client := errutil.Must(baseCmd.CreateHTTPClient())

		var url string
		if baseCmd.EnableTLS {
			url = fmt.Sprintf("https://%s:%d/version", baseCmd.TLSHost, baseCmd.TLSPort)
		} else {
			url = fmt.Sprintf("http://%s:%d/version", baseCmd.Host, baseCmd.Port)
		}

		resp, err := client.Get(url)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to retrieve version")
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to read body")
		}

		var v version.Version
		if err = json.Unmarshal(body, &v); err != nil {
			log.Fatal().Err(err).Msg("Failed to parse response")
		}

		fmt.Fprintf(cmd.OutOrStdout(), `
   version: %s
    commit: %s
 buildDate: %s
apiVersion: %s
`, v.Version, v.Commit, v.BuildDate, v.APIVersion)

		return nil
	},
}
