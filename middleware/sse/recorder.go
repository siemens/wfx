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
)

// MockResponseRecorder extends httptest.ResponseRecorder and implements http.Hijacker.
type MockResponseRecorder struct {
	*httptest.ResponseRecorder
	hijackedConn net.Conn
	ChResponse   chan string
}

// Ensure MockResponseRecorder implements http.Hijacker.
var _ http.Hijacker = &MockResponseRecorder{}

// NewMockResponseRecorder creates a new MockResponseRecorder.
func NewMockResponseRecorder() *MockResponseRecorder {
	return &MockResponseRecorder{
		ResponseRecorder: httptest.NewRecorder(),
		ChResponse:       make(chan string),
	}
}

// Hijack implements the http.Hijacker interface.
func (c *MockResponseRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if c.hijackedConn == nil {
		// Create a dummy connection for testing purposes.
		var r net.Conn
		c.hijackedConn, r = net.Pipe()
		go func() {
			response, _ := io.ReadAll(r)
			c.ChResponse <- string(response)
		}()
	}
	rw := bufio.NewReadWriter(bufio.NewReader(c.hijackedConn), bufio.NewWriter(c.hijackedConn))
	return c.hijackedConn, rw, nil
}
