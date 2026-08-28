package logging

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"bufio"
	"bytes"
	"io"
	"net"
	"net/http"

	"github.com/Southclaws/fault"
)

type responseWriter struct {
	responseBody bytes.Buffer
	bodyWriter   io.Writer
	httpWriter   http.ResponseWriter
	statusCode   int
	hijacked     bool
}

// responseWriter implements the following interfaces (compile-time check):
var (
	_ http.Flusher  = (*responseWriter)(nil)
	_ http.Hijacker = (*responseWriter)(nil)
)

func newMyResponseWriter(w http.ResponseWriter, interceptBody bool) *responseWriter {
	var result responseWriter
	if interceptBody {
		result.bodyWriter = io.MultiWriter(w, &result.responseBody)
	} else {
		result.bodyWriter = w
	}
	result.httpWriter = w
	return &result
}

func (w *responseWriter) Header() http.Header {
	return w.httpWriter.Header()
}

func (w *responseWriter) WriteHeader(statusCode int) {
	if w.hijacked {
		return
	}
	w.statusCode = statusCode
	w.httpWriter.WriteHeader(statusCode)
}

func (w *responseWriter) Write(b []byte) (int, error) {
	if w.hijacked {
		return 0, fault.Wrap(http.ErrHijacked)
	}
	n, err := w.bodyWriter.Write(b)
	return n, fault.Wrap(err)
}

// Flush is used by the server-sent events implementation to flush a single event to the client.
// This is part of the http.Flusher interface.
func (w *responseWriter) Flush() {
	if w.hijacked {
		return
	}
	if flusher, ok := w.httpWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

// Hijack allows an HTTP handler to take over the underlying connection.
// This is used by the server-sent events implementation to keep the long-running (idle) connection open.
// NOTE: The "funny" name comes from Golang's http.Hijacker interface.
func (w *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hj, ok := w.httpWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fault.New("hijacker interface not supported")
	}
	conn, bw, err := hj.Hijack()
	if err == nil {
		w.hijacked = true
	}
	return conn, bw, fault.Wrap(err)
}
