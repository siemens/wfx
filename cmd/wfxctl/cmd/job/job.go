package job

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"github.com/siemens/wfx/cmd/wfxctl/cmd/job/addtags"
	"github.com/siemens/wfx/cmd/wfxctl/cmd/job/create"
	"github.com/siemens/wfx/cmd/wfxctl/cmd/job/delete"
	"github.com/siemens/wfx/cmd/wfxctl/cmd/job/deltags"
	"github.com/siemens/wfx/cmd/wfxctl/cmd/job/events"
	"github.com/siemens/wfx/cmd/wfxctl/cmd/job/get"
	"github.com/siemens/wfx/cmd/wfxctl/cmd/job/getdefinition"
	"github.com/siemens/wfx/cmd/wfxctl/cmd/job/getstatus"
	"github.com/siemens/wfx/cmd/wfxctl/cmd/job/gettags"
	"github.com/siemens/wfx/cmd/wfxctl/cmd/job/query"
	"github.com/siemens/wfx/cmd/wfxctl/cmd/job/updatedefinition"
	"github.com/siemens/wfx/cmd/wfxctl/cmd/job/updatestatus"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:              "job",
		Short:            "manage jobs",
		Long:             "subcommand to manage jobs (CRUD)",
		TraverseChildren: true,
		SilenceUsage:     true,
	}
	cmd.AddCommand(create.NewCommand())
	cmd.AddCommand(delete.NewCommand())
	cmd.AddCommand(get.NewCommand())
	cmd.AddCommand(query.NewCommand())
	cmd.AddCommand(updatestatus.NewCommand())
	cmd.AddCommand(getstatus.NewCommand())
	cmd.AddCommand(updatedefinition.NewCommand())
	cmd.AddCommand(getdefinition.NewCommand())
	cmd.AddCommand(addtags.NewCommand())
	cmd.AddCommand(deltags.NewCommand())
	cmd.AddCommand(gettags.NewCommand())
	cmd.AddCommand(events.NewCommand())
	return cmd
}
