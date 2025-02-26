//go:build !windows

package main

/*
 * SPDX-FileCopyrightText: 2024 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"log"
	"os/exec"
	"syscall"
	"time"
)

func createTtyCmd(args []string) *exec.Cmd {
	args = append(args, "bash", "-l")
	ttydCmd = exec.Command("ttyd", args...)
	ttydCmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	return ttydCmd
}

func terminate() {
	if ttydCmd != nil {
		// ensure ttyd and children are stopped
		pid := ttydCmd.Process.Pid
		ttydCmd = nil
		if err := syscall.Kill(-pid, 0); err == nil {
			// process still running
			log.Println("Sending SIGTERM to", pid)
			_ = syscall.Kill(-pid, syscall.SIGTERM)
		}
		time.Sleep(3 * time.Second)
		if err := syscall.Kill(-pid, 0); err == nil {
			// process still running
			log.Println("Sending SIGKILL to", pid)
			_ = syscall.Kill(-pid, syscall.SIGKILL)
		}
	}
}
