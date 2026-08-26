package cmd

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"fmt"

	"github.com/rs/zerolog"
	"github.com/siemens/wfx/cmd/wfx/cmd/config"
	"github.com/siemens/wfx/cmd/wfxctl/cmd/health"
	"github.com/siemens/wfx/cmd/wfxctl/cmd/job"
	"github.com/siemens/wfx/cmd/wfxctl/cmd/version"
	"github.com/siemens/wfx/cmd/wfxctl/cmd/workflow"
	"github.com/siemens/wfx/cmd/wfxctl/flags"
	"github.com/siemens/wfx/cmd/wfxctl/metadata"
	"github.com/siemens/wfx/internal/cmd/man"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "wfxctl",
		Short: "wfxctl is a command-line tool to interact with the wfx REST API",
		Long: `wfxctl can be used for management or diagnostic purposes.

Tip: Shell completion is available for Bash, Fish and Zsh. See wfxctl completion --help for more information.
`,
		Version:          fmt.Sprintf("%s (commit %s)", metadata.Version, metadata.Commit),
		SilenceUsage:     true,
		TraverseChildren: true,
	}
	cmd.AddCommand(man.NewCommand())
	cmd.AddCommand(job.NewCommand())
	cmd.AddCommand(workflow.NewCommand())
	cmd.AddCommand(version.NewCommand())
	cmd.AddCommand(health.NewCommand())

	f := cmd.PersistentFlags()
	f.StringSlice(flags.ConfigFlag, config.DefaultConfigFiles(), "path to one or more .yaml config files; if this option is not set, then the default paths are tried")
	_ = cmd.MarkPersistentFlagFilename(flags.ConfigFlag, "yml", "yaml")

	f.String(flags.HostFlag, "http://localhost:8081", "server URL (http://, https://, or unix:///path/to/socket)")
	f.String(flags.TLSCaFlag, "", "ca bundle (PEM)")

	f.StringArray(flags.HeaderFlag, nil, "add an HTTP header, e.g. --header 'Authorization: Bearer $TOKEN' (may be given multiple times)")
	f.String(flags.CredentialHelperFlag, "", "credential helper to invoke for HTTP authentication (wfxctl-credential-<name>, or path containing '/')")

	f.String(flags.FilterFlag, "", "output filter (jq-expression). example: '.id'")
	f.Bool(flags.RawFlag, false, "output raw strings, not JSON texts; use --filter to select a single entity")

	f.String(flags.LogLevelFlag, "info", fmt.Sprintf("set log level. one of: %s,%s,%s,%s,%s,%s,%s",
		zerolog.TraceLevel.String(),
		zerolog.DebugLevel.String(),
		zerolog.InfoLevel.String(),
		zerolog.WarnLevel.String(),
		zerolog.ErrorLevel.String(),
		zerolog.FatalLevel.String(),
		zerolog.PanicLevel.String()))

	return cmd
}
