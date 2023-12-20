//go:build windows

package plugin

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"os"

	"github.com/Southclaws/fault"
)

func (p *Plugin) terminateProcess() error {
	pid := p.cmd.Process.Pid
	proc, err := os.FindProcess(pid)
	if err != nil {
		return fault.Wrap(err)
	}
	// note: this does not kill child processes
	return fault.Wrap(proc.Kill())
}
