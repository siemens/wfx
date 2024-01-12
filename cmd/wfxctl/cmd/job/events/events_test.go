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
	"net/url"
	"strconv"
	"testing"

	"github.com/siemens/wfx/cmd/wfxctl/flags"
	"github.com/stretchr/testify/assert"
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

	u, _ := url.Parse(ts.URL)
	_ = flags.Koanf.Set(flags.ClientHostFlag, u.Hostname())
	port, _ := strconv.Atoi(u.Port())
	_ = flags.Koanf.Set(flags.ClientPortFlag, port)

	_ = flags.Koanf.Set(jobIDFlag, "1")

	err := Command.Execute()
	assert.NoError(t, err)

	assert.Equal(t, expectedPath, actualPath)
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
