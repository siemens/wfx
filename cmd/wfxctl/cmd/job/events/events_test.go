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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/siemens/wfx/cmd/wfxctl/flags"
	"github.com/stretchr/testify/assert"
)

func TestSSETransportWriteEventGron(t *testing.T) {
	var out bytes.Buffer
	transport := SSETransport{out: &out, format: "gron"}

	assert.NoError(t, transport.writeEvent(`{"id":"1","action":"UPDATE_STATUS"}`))
	assert.Equal(t, "json = {};\njson.action = \"UPDATE_STATUS\";\njson.id = \"1\";\n\n", out.String())
}

func TestSubscribeJobStatusGron(t *testing.T) {
	for _, tc := range []struct {
		name    string
		payload string
		output  string
		error   string
	}{
		{
			name:    "two events",
			payload: "data: {\"id\":\"1\"}\n\n: keepalive\n\ndata: {\"id\":\"2\"}\n\n",
			output:  "json = {};\njson.id = \"1\";\n\njson = {};\njson.id = \"2\";\n\n",
			error:   "connection to server lost",
		},
		{
			name:    "invalid payload",
			payload: "data: invalid\n\ndata: {\"id\":\"2\"}\n\n",
			error:   "invalid character",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var accept string
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				accept = r.Header.Get("Accept")
				w.Header().Set("Content-Type", "text/event-stream")
				_, _ = w.Write([]byte(tc.payload))
			}))
			t.Cleanup(ts.Close)
			t.Setenv("WFX_HOST", ts.URL)
			t.Setenv("WFX_FORMAT", "gron")

			var out bytes.Buffer
			cmd := NewCommand()
			cmd.SetOut(&out)
			cmd.SetErr(&bytes.Buffer{})
			assert.ErrorContains(t, cmd.Execute(), tc.error)
			assert.Equal(t, "text/event-stream", accept)
			assert.Equal(t, tc.output, out.String())
		})
	}
}

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
