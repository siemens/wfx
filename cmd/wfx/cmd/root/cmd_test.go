//go:build linux

package root

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/knadh/koanf/v2"
	"github.com/rs/zerolog"
	"github.com/siemens/wfx/cmd/wfxctl/flags"
	"github.com/stretchr/testify/require"
)

func TestCommand_ReloadConfig(t *testing.T) {
	dir, err := os.MkdirTemp("", "wfx-reload-test")
	require.NoError(t, err)

	cfgFile, err := os.CreateTemp("", "wfxconfig")
	require.NoError(t, err)

	var clientSocket, mgmtSocket string
	k.Read(func(k *koanf.Koanf) {
		clientSocket = k.String(flags.ClientUnixSocketFlag)
		mgmtSocket = k.String(flags.MgmtUnixSocketFlag)
	})

	t.Cleanup(func() {
		_ = cfgFile.Close()
		_ = os.Remove(cfgFile.Name())
		_ = os.Remove(clientSocket)
		_ = os.Remove(mgmtSocket)
		_ = os.RemoveAll(dir)
	})

	_, err = cfgFile.Write([]byte("log-level: warn"))
	require.NoError(t, err)

	Command.SetArgs([]string{
		"--config", cfgFile.Name(),
		"--storage-opt=file:wfx?mode=memory&cache=shared&_fk=1",
		"--scheme=unix",
	})

	err = errors.New("Should be set to nil by the goroutine")

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		err = Command.Execute()
	}()

	waitForLogLevel(t, zerolog.WarnLevel)

	// wait for inotify to be set up
	foundInotify := false
	for foundInotify {
		fddir := fmt.Sprintf("/proc/%d/fd", os.Getpid())
		files, err := os.ReadDir(fddir)
		require.NoError(t, err)
		for _, f := range files {
			fname, _ := os.Readlink(fmt.Sprintf("%s/%s", fddir, f.Name()))
			if strings.HasPrefix(fname, "anon_inode") {
				foundInotify = true
				break
			}
		}
		time.Sleep(10 * time.Millisecond)
	}

	// modify config file
	_, _ = cfgFile.Seek(0, 0)
	_, err = cfgFile.Write([]byte("log-level: trace"))
	require.NoError(t, err)

	waitForLogLevel(t, zerolog.TraceLevel)

	// tell go routine to stop
	signalChannel <- os.Interrupt

	wg.Wait()
	require.NoError(t, err)
}

func waitForLogLevel(t *testing.T, expected zerolog.Level) {
	for i := 0; i < 500; i++ {
		if zerolog.GlobalLevel() == expected {
			break
		}
		time.Sleep(time.Millisecond * 10)
	}
	require.Equal(t, expected.String(), zerolog.GlobalLevel().String())
}
