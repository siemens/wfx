package events

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/siemens/wfx/cmd/wfxctl/flags"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmaxmax/go-sse"
)

func TestSubscribeJobStatus(t *testing.T) {
	const expectedPath = "/api/wfx/v1/jobs/events"
	var actualPath string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actualPath = r.URL.Path

		w.Header().Add("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`data: "hello world"

`))
	}))
	t.Cleanup(ts.Close)

	t.Setenv("WFX_HOST", ts.URL)

	cmd := NewCommand()
	cmd.SetArgs([]string{"--" + flags.JobIDFlag, "1"})
	err := cmd.Execute()
	assert.ErrorContains(t, err, "connection to server lost")
	assert.Equal(t, expectedPath, actualPath)
}

func TestSubscribeJobStatusHeaders(t *testing.T) {
	var actualHeader string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actualHeader = r.Header.Get("X-Custom")
		w.Header().Add("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(ts.Close)

	t.Setenv("WFX_HOST", ts.URL)

	cmd := NewCommand()
	cmd.Flags().StringArray(flags.HeaderFlag, nil, "")
	cmd.SetArgs([]string{"--" + flags.HeaderFlag, "X-Custom: value"})
	err := cmd.Execute()
	assert.ErrorContains(t, err, "connection to server lost")
	assert.Equal(t, "value", actualHeader)
}

func TestValidator_OK(t *testing.T) {
	out := new(bytes.Buffer)
	resp := http.Response{StatusCode: http.StatusOK}
	err := validator(out)(&resp)
	assert.Nil(t, err)
}

func TestValidator_Error(t *testing.T) {
	out := new(bytes.Buffer)
	resp := http.Response{StatusCode: http.StatusInternalServerError}
	err := validator(out)(&resp)
	assert.NotNil(t, err)
}

func TestValidator_BadRequest(t *testing.T) {
	out := new(bytes.Buffer)

	rec := httptest.NewRecorder()
	rec.WriteHeader(http.StatusBadRequest)

	resp := rec.Result()
	err := validator(out)(resp)
	assert.NotNil(t, err)
}

func TestValidator_BadRequestInvalidJson(t *testing.T) {
	out := new(bytes.Buffer)

	rec := httptest.NewRecorder()
	rec.WriteHeader(http.StatusBadRequest)
	_, _ = rec.WriteString("data: foo")

	resp := rec.Result()
	err := validator(out)(resp)
	assert.NotNil(t, err)
}

func TestSSERequestIDErrors(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status int
		body   string
	}{
		{"disconnect", http.StatusOK, "data: hello\n\n"},
		{"json error", http.StatusBadRequest, `{"errors":[{"code":"bad","message":"bad request","logref":"legacy"}]}`},
		{"plain error", http.StatusBadGateway, "bad gateway"},
		{"empty error", http.StatusServiceUnavailable, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("X-Request-ID", "server-id")
				w.WriteHeader(tc.status)
				_, _ = io.WriteString(w, tc.body)
			}))
			defer ts.Close()
			t.Setenv("WFX_HOST", ts.URL)
			cmd := NewCommand()
			cmd.SetOut(io.Discard)
			cmd.SetErr(io.Discard)
			assert.ErrorContains(t, cmd.Execute(), `X-Request-ID: "server-id"`)
		})
	}
}

func TestSSERetryRequestID(t *testing.T) {
	attempts := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.Header().Set("X-Request-ID", fmt.Sprintf("server-%d", attempts))
		if attempts == 2 {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
	}))
	defer ts.Close()
	client := *sse.DefaultClient
	client.ResponseValidator = validator(io.Discard)
	client.Backoff.MaxRetries = 1
	client.Backoff.InitialInterval = time.Millisecond
	var retryErr error
	client.OnRetry = func(err error, _ time.Duration) { retryErr = err }
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, ts.URL, nil)
	require.NoError(t, err)
	transport := SSETransport{sseClient: &client, out: io.Discard}
	_, err = transport.Do(req)
	assert.ErrorContains(t, retryErr, `X-Request-ID: "server-1"`)
	assert.ErrorContains(t, err, `X-Request-ID: "server-2"`)
	assert.NotContains(t, err.Error(), "server-1")
}
