//go:build !windows

package plugin

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"fmt"
	"os/exec"
	"syscall"
	"time"

	"github.com/Southclaws/fault"
	"github.com/rs/zerolog/log"
)

var gracefulTimeout = 15 * time.Second

func createCmd(path string) *exec.Cmd {
	cmd := exec.Command(path)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	return cmd
}

func (p *FBPlugin) terminateProcess() error {
	pid := p.cmd.Process.Pid
	if err := syscall.Kill(-pid, 0); err == nil {
		pid = -pid // this is the pid of the process group
	} else {
		log.Warn().Err(err).Int("pid", pid).Msg("Process group not found")
		if err := syscall.Kill(pid, 0); err != nil {
			return fmt.Errorf("plugin pid %d not found", pid)
		}
	}

	// signal is sent to *every* process in the process group
	log.Debug().Int("pid", pid).Msg("Sending SIGTERM")
	if err := syscall.Kill(pid, syscall.SIGTERM); err != nil {
		return fault.Wrap(err)
	}

	done := make(chan bool)
	go func() {
		for i := 0; i < 100; i++ {
			if p.waited.Load() {
				done <- true
				break
			}
			time.Sleep(10 * time.Millisecond)
		}
	}()

	for {
		select {
		case <-done:
			log.Debug().Msg("Process terminated")
			return nil
		case <-time.After(gracefulTimeout):
			// check if process is still alive
			if err := syscall.Kill(pid, 0); err == nil {
				// process is still alive
				log.Warn().Int("pid", pid).Msg("Process still alive, sending SIGKILL")
				_ = syscall.Kill(pid, syscall.SIGKILL)
			}
		}
	}
}
