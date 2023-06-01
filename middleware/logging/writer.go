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

func newMyResponseWriter(w http.ResponseWriter) *responseWriter {
	var result responseWriter
	result.bodyWriter = io.MultiWriter(w, &result.responseBody)
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
