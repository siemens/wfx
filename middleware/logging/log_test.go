package logging

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
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

func TestLog(t *testing.T) {
	mw := MW{}
	handler := mw.Wrap(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintln(w, "Hello, client")
	}))

	ts := httptest.NewServer(handler)
	defer ts.Close()

	res, err := http.Get(ts.URL)
	assert.NoError(t, err)

	greeting, err := io.ReadAll(res.Body)
	defer res.Body.Close()
	assert.NoError(t, err)

	assert.Equal(t, "Hello, client\n", string(greeting))
	mw.Shutdown()
}

func TestLogDebug(t *testing.T) {
	handler := MW{}.Wrap(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintln(w, "Hello, client")
	}))

	ts := httptest.NewServer(handler)
	defer ts.Close()

	res, err := http.Get(ts.URL)
	assert.NoError(t, err)

	greeting, err := io.ReadAll(res.Body)
	defer res.Body.Close()
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

	handler := MW{}.Wrap(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintln(w, "Hello, client")
	}))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, r)
}
