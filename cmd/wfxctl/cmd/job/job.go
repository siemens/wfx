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
	"github.com/siemens/wfx/cmd/wfxctl/cmd/job/get"
	"github.com/siemens/wfx/cmd/wfxctl/cmd/job/getdefinition"
	"github.com/siemens/wfx/cmd/wfxctl/cmd/job/getstatus"
	"github.com/siemens/wfx/cmd/wfxctl/cmd/job/gettags"
	"github.com/siemens/wfx/cmd/wfxctl/cmd/job/query"
	"github.com/siemens/wfx/cmd/wfxctl/cmd/job/updatedefinition"
	"github.com/siemens/wfx/cmd/wfxctl/cmd/job/updatestatus"
	"github.com/spf13/cobra"
)

var Command = &cobra.Command{
	Use:              "job",
	Short:            "manage jobs",
	Long:             "subcommand to manage jobs (CRUD)",
	TraverseChildren: true,
	SilenceUsage:     true,
}

func init() {
	Command.AddCommand(create.Command)
	Command.AddCommand(delete.Command)
	Command.AddCommand(get.Command)
	Command.AddCommand(query.Command)
	Command.AddCommand(updatestatus.Command)
	Command.AddCommand(getstatus.Command)
	Command.AddCommand(updatedefinition.Command)
	Command.AddCommand(getdefinition.Command)
	Command.AddCommand(addtags.Command)
	Command.AddCommand(deltags.Command)
	Command.AddCommand(gettags.Command)
}
