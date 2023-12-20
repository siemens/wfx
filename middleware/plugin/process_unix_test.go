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
	"io"
	"os"
	"os/exec"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPerminateProcess_NoProcessGroup(t *testing.T) {
	f, _ := os.CreateTemp("", "TestPerminateProcess_NoProcessGroup.*.sh")
	fname := f.Name()
	t.Cleanup(func() { _ = os.Remove(fname) })
	_, _ = io.WriteString(f, `#!/usr/bin/env bash
while true; do
    sleep 60
done`)
	f.Close()
	_ = os.Chmod(fname, 0o755)

	gracefulTimeout = time.Microsecond

	cmd := exec.Command(fname)
	p := FBPlugin{cmd: cmd}
	err := cmd.Start()
	require.NoError(t, err)
	require.NotEqual(t, 0, p.cmd.Process.Pid)

	var g sync.WaitGroup
	awaitKillTestHelper(&p, &g)

	err = p.Stop()
	assert.NoError(t, err)

	g.Wait()
}

func TestPerminateProcess_KillStuckProcess(t *testing.T) {
	f, _ := os.CreateTemp("", "TestPerminateProcess_KillStuckProcess.*.sh")
	fname := f.Name()
	t.Cleanup(func() { _ = os.Remove(fname) })
	_, _ = io.WriteString(f, `#!/usr/bin/env bash
trap '' SIGTERM
while true; do
    sleep 60
done`)
	f.Close()
	_ = os.Chmod(fname, 0o755)

	gracefulTimeout = time.Millisecond

	cmd := exec.Command(fname)
	p := FBPlugin{cmd: cmd}
	err := cmd.Start()
	require.NoError(t, err)
	require.NotEqual(t, 0, p.cmd.Process.Pid)

	var g sync.WaitGroup
	awaitKillTestHelper(&p, &g)

	err = p.Stop()
	assert.NoError(t, err)
}

func awaitKillTestHelper(p *FBPlugin, g *sync.WaitGroup) {
	g.Add(1)
	go func() {
		defer g.Done()
		_ = p.cmd.Wait()
		p.waited.Store(true)
	}()
}
