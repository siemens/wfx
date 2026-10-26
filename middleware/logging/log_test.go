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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLog(t *testing.T) {
	mw := NewLoggingMiddleware()
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintln(w, "Hello, client")
	}))

	ts := httptest.NewServer(handler)
	defer ts.Close()

	res, err := http.Get(ts.URL)
	assert.NoError(t, err)

	greeting, err := io.ReadAll(res.Body)
	defer func() { _ = res.Body.Close() }()
	assert.NoError(t, err)

	assert.Equal(t, "Hello, client\n", string(greeting))
}

func TestLogDebug(t *testing.T) {
	handler := NewLoggingMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintln(w, "Hello, client")
	}))

	ts := httptest.NewServer(handler)
	defer ts.Close()

	res, err := http.Get(ts.URL)
	assert.NoError(t, err)

	greeting, err := io.ReadAll(res.Body)
	defer func() { _ = res.Body.Close() }()
	assert.NoError(t, err)

	assert.Equal(t, "Hello, client\n", string(greeting))
}

func TestLoggerFomCtx(t *testing.T) {
	logger := zerolog.New(io.Discard)
	ctx := context.WithValue(context.Background(), KeyRequestLogger, logger)
	actual := LoggerFromCtx(ctx)
	assert.Equal(t, logger, actual)
}

func TestLoggerFomCtx_Default(t *testing.T) {
	actual := LoggerFromCtx(context.Background())
	assert.Equal(t, log.Logger, actual)
}

type FaultyReadCloser struct{}

func (r FaultyReadCloser) Read([]byte) (n int, err error) {
	return 0, errors.New("failed to read")
}

func (r FaultyReadCloser) Close() error {
	return nil
}

func TestPeekBody_ReadFailure(t *testing.T) {
	var body FaultyReadCloser
	r := &http.Request{Body: body}
	_, err := PeekBody(r)
	assert.NotNil(t, err)

	handler := NewLoggingMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintln(w, "Hello, client")
	}))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, r)
}

func TestRequestID(t *testing.T) {
	var logs bytes.Buffer
	previous := log.Logger
	log.Logger = zerolog.New(&logs).Level(zerolog.DebugLevel)
	t.Cleanup(func() { log.Logger = previous })
	handler := NewLoggingMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := LoggerFromCtx(r.Context())
		logger.Info().Msg("handler")
		w.WriteHeader(http.StatusBadRequest)
	}))
	ids := make(map[string]bool)
	for range 2 {
		logs.Reset()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-Request-ID", "caller-provided")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		id := rec.Result().Header.Get("X-Request-ID")
		parsed, err := uuid.Parse(id)
		require.NoError(t, err)
		assert.Equal(t, uuid.Version(4), parsed.Version())
		assert.False(t, ids[id])
		ids[id] = true
		decoder := json.NewDecoder(&logs)
		count := 0
		for decoder.More() {
			var entry map[string]any
			require.NoError(t, decoder.Decode(&entry))
			assert.Equal(t, id, entry["reqID"])
			count++
		}
		assert.Equal(t, 3, count)
	}
}
