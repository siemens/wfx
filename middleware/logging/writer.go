package logging

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"bytes"
	"io"
	"net/http"

	"github.com/Southclaws/fault"
)

type responseWriter struct {
	responseBody bytes.Buffer
	bodyWriter   io.Writer
	httpWriter   http.ResponseWriter
	statusCode   int
}

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
	w.statusCode = statusCode
	w.httpWriter.WriteHeader(statusCode)
}

func (w *responseWriter) Write(b []byte) (int, error) {
	n, err := w.bodyWriter.Write(b)
	return n, fault.Wrap(err)
}

// Flush implements the http.Flusher interface.
// This is used by the server-sent events implementation to flush a single event to the client.
func (w *responseWriter) Flush() {
	flusher := w.httpWriter.(http.Flusher)
	flusher.Flush()
}
