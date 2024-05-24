//go:build windows

package main

/*
 * SPDX-FileCopyrightText: 2024 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"os"
	"os/exec"
)

func createTtyCmd(args []string) *exec.Cmd {
	args = append(args, "cmd")
	return exec.Command("ttyd", args...)
}

func terminate() {
	if ttydCmd != nil {
		// ensure ttyd and children are stopped
		pid := ttydCmd.Process.Pid

		proc, _ := os.FindProcess(pid)
		if proc != nil {
			proc.Kill()
		}
	}
}
