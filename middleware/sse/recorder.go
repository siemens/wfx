//go:build testing

package sse

/*
 * SPDX-FileCopyrightText: 2025 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"bufio"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

// MockResponseRecorder extends httptest.ResponseRecorder and implements http.Hijacker.
type MockResponseRecorder struct {
	t *testing.T
	*httptest.ResponseRecorder
	hijackedConn net.Conn
	lines        []string
	muLines      sync.Mutex
}

// Ensure MockResponseRecorder implements http.Hijacker.
var _ http.Hijacker = &MockResponseRecorder{}

// NewMockResponseRecorder creates a new MockResponseRecorder.
func NewMockResponseRecorder(t *testing.T) *MockResponseRecorder {
	return &MockResponseRecorder{
		t:                t,
		ResponseRecorder: httptest.NewRecorder(),
		lines:            make([]string, 0),
	}
}

// Hijack implements the http.Hijacker interface.
func (c *MockResponseRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if c.hijackedConn == nil {
		// Create a dummy connection for testing purposes.
		var r net.Conn
		c.hijackedConn, r = net.Pipe()
		bufReader := bufio.NewReader(r)
		go func() {
			for {
				line, err := bufReader.ReadString('\n')
				if err != nil {
					// Break the loop if EOF or another error occurs
					if err != io.EOF {
						require.NoError(c.t, err)
					}
					break
				}
				c.muLines.Lock()
				c.lines = append(c.lines, line)
				c.muLines.Unlock()
			}
		}()
	}
	rw := bufio.NewReadWriter(bufio.NewReader(c.hijackedConn), bufio.NewWriter(c.hijackedConn))
	return c.hijackedConn, rw, nil
}

func (c *MockResponseRecorder) Response() string {
	c.muLines.Lock()
	defer c.muLines.Unlock()
	return strings.Join(c.lines, "")
}
