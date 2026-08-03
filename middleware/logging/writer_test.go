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
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriter(t *testing.T) {
	recorder := httptest.NewRecorder()

	w := newMyResponseWriter(recorder, true)
	w.WriteHeader(http.StatusOK)

	assert.NotNil(t, w.Header())

	_, err := w.Write([]byte("hello world"))
	require.NoError(t, err)

	result := recorder.Result()
	assert.Equal(t, http.StatusOK, result.StatusCode)

	body, err := io.ReadAll(result.Body)
	require.NoError(t, err)
	defer func() { _ = result.Body.Close() }()

	assert.Equal(t, "hello world", string(body))
	assert.Equal(t, "hello world", w.responseBody.String())
}

func TestWriterImplementsFlusher(t *testing.T) {
	recorder := httptest.NewRecorder()
	var w http.ResponseWriter = newMyResponseWriter(recorder, true)
	flusher, ok := w.(http.Flusher)
	assert.True(t, ok)
	flusher.Flush()
}

func TestWriterIgnoreBody(t *testing.T) {
	recorder := httptest.NewRecorder()
	w := newMyResponseWriter(recorder, false)
	_, err := w.Write([]byte("hello world"))
	require.NoError(t, err)
	assert.Empty(t, w.responseBody.String())
}

// hijackableRecorder augments httptest.ResponseRecorder with a (fake) http.Hijacker
// implementation, which the recorder itself does not provide.
type hijackableRecorder struct {
	*httptest.ResponseRecorder
	conn net.Conn
	err  error
}

func (r *hijackableRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if r.err != nil {
		return nil, nil, r.err
	}
	return r.conn, bufio.NewReadWriter(bufio.NewReader(r.conn), bufio.NewWriter(r.conn)), nil
}

func TestWriterHijackNotSupported(t *testing.T) {
	// httptest.ResponseRecorder does not implement http.Hijacker
	w := newMyResponseWriter(httptest.NewRecorder(), true)
	_, _, err := w.Hijack()
	require.Error(t, err)
	assert.False(t, w.hijacked)
}

func TestWriterHijackFails(t *testing.T) {
	recorder := &hijackableRecorder{ResponseRecorder: httptest.NewRecorder(), err: errors.New("nope")}
	w := newMyResponseWriter(recorder, true)

	_, _, err := w.Hijack()
	require.Error(t, err)
	assert.False(t, w.hijacked, "failed hijack must not mark the writer as hijacked")

	// writer must still be usable
	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte("hello world"))
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, w.statusCode)
	assert.Equal(t, "hello world", recorder.Body.String())
}

func TestWriterHijacked(t *testing.T) {
	client, server := net.Pipe()
	defer func() { _ = client.Close() }()
	defer func() { _ = server.Close() }()

	recorder := &hijackableRecorder{ResponseRecorder: httptest.NewRecorder(), conn: server}
	w := newMyResponseWriter(recorder, true)

	conn, bw, err := w.Hijack()
	require.NoError(t, err)
	assert.Same(t, server, conn)
	assert.NotNil(t, bw)
	assert.True(t, w.hijacked)

	// subsequent calls must be no-ops / errors and must not touch the recorder
	w.WriteHeader(http.StatusTeapot)
	assert.Zero(t, w.statusCode)

	n, err := w.Write([]byte("hello world"))
	require.ErrorIs(t, err, http.ErrHijacked)
	assert.Zero(t, n)
	assert.Empty(t, recorder.Body.String())
	assert.Empty(t, w.responseBody.String())

	w.Flush()
	assert.False(t, recorder.Flushed)
}
