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
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/knadh/koanf/v2"
	"github.com/rs/zerolog"
	"github.com/siemens/wfx/cmd/wfxctl/flags"
	"github.com/stretchr/testify/assert"
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

func TestAdoptListeners(t *testing.T) {
	assert.Empty(t, adoptListeners(nil, nil))
	assert.Empty(t, adoptListeners([]net.Listener{nil, nil, nil}, nil))

	sock1, _ := os.CreateTemp(os.TempDir(), "TestAdoptListeners*.sock")
	sock2, _ := os.CreateTemp(os.TempDir(), "TestAdoptListeners*.sock")
	t.Cleanup(func() {
		_ = os.Remove(sock1.Name())
		_ = os.Remove(sock2.Name())
	})

	ln1, _ := net.Listen("unix", sock1.Name())
	ln2, _ := net.Listen("unix", sock1.Name())

	collection := adoptListeners([]net.Listener{ln1, ln2}, nil)
	t.Cleanup(func() {
		for _, sc := range collection {
			sc.Shutdown(context.Background())
		}
	})
	assert.Len(t, collection, 2)
	assert.Len(t, collection[0].servers, 1)
	assert.Len(t, collection[1].servers, 1)

	actualLn1, err := collection[0].servers[0].Listener()
	assert.Nil(t, err)
	assert.Equal(t, ln1, actualLn1)

	actualLn2, err := collection[1].servers[0].Listener()
	assert.Nil(t, err)
	assert.Equal(t, ln2, actualLn2)
}

func TestLaunchServer_ListenerError(t *testing.T) {
	srv := myServer{
		Listener: func() (net.Listener, error) {
			return nil, errors.New("something went wrong")
		},
	}
	err := launchServer("test", srv)
	assert.NotNil(t, err)
}

func TestLaunchServer_UDS(t *testing.T) {
	f, _ := os.CreateTemp(os.TempDir(), "TestLaunchServer_UDS.*.sock")
	f.Close()
	_ = os.Remove(f.Name())
	ln, err := net.Listen("unix", f.Name())
	require.NoError(t, err)

	srv := myServer{
		Srv: &http.Server{Handler: http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			rw.WriteHeader(http.StatusOK)
		})},
		Listener: func() (net.Listener, error) {
			return ln, nil
		},
		Kind: kindUnix,
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := launchServer("notls", srv)
		assert.NoError(t, err)
	}()

	for i := 0; i < 30; i++ {
		time.Sleep(time.Millisecond * 10)
		conn, err := net.Dial("unix", f.Name())
		if err != nil {
			continue
		}
		defer conn.Close()

		req, err := http.NewRequest(http.MethodGet, "/", nil)
		require.NoError(t, err)
		client := &http.Client{
			Transport: &http.Transport{
				DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
					return conn, nil
				},
			},
		}
		resp, err := client.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			break
		}
	}

	_ = srv.Srv.Shutdown(context.Background())
	wg.Wait()
}

func TestLaunchServer_ListenerClosed(t *testing.T) {
	f, _ := os.CreateTemp(os.TempDir(), "TestLaunchServer_UDS.*.sock")
	f.Close()
	_ = os.Remove(f.Name())
	ln, _ := net.Listen("unix", f.Name())
	_ = ln.Close()

	srv := myServer{
		Srv: &http.Server{Handler: http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			rw.WriteHeader(http.StatusOK)
		})},
		Listener: func() (net.Listener, error) {
			return ln, nil
		},
		Kind: kindUnix,
	}

	err := launchServer("notls", srv)
	assert.Error(t, err)
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
