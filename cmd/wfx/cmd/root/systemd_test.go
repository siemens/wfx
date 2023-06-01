package root

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestAdoptSystemdSockets_None(t *testing.T) {
	err := adoptSystemdSockets(nil, nil, nil)
	require.NoError(t, err)
}

func TestAdoptSystemdSockets_InvalidFdCount(t *testing.T) {
	err := adoptSystemdSockets([]net.Listener{nil}, nil, nil)
	require.Error(t, err)
}

func TestAdoptSystemdSockets(t *testing.T) {
	errChan := make(chan error)

	tempDir, err := ioutil.TempDir("", "wfx-test-systemd")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	clientSock := path.Join(tempDir, "client.sock")
	clientListener, err := net.Listen("unix", clientSock)
	require.NoError(t, err)
	defer clientListener.Close()

	mgmtSock := path.Join(tempDir, "mgmt.sock")
	mgmtListener, err := net.Listen("unix", mgmtSock)
	require.NoError(t, err)
	defer mgmtListener.Close()

	err = adoptSystemdSockets([]net.Listener{clientListener, mgmtListener}, nil, errChan)
	require.NoError(t, err)

	for _, sock := range []string{clientSock, mgmtSock} {
		dialer := net.Dialer{Timeout: 5 * time.Second}
		conn, err := dialer.Dial("unix", sock)
		require.NoError(t, err)
		defer conn.Close()

		client := &http.Client{
			Transport: &http.Transport{
				Dial: func(_, _ string) (net.Conn, error) {
					return conn, nil
				},
			},
		}

		request, err := http.NewRequest(http.MethodGet, "http://localhost/version", nil)
		require.NoError(t, err)

		response, err := client.Do(request)
		require.NoError(t, err)
		defer response.Body.Close()
	}
}
