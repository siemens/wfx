package subscribestatus

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/siemens/wfx/api"
	"github.com/siemens/wfx/cmd/wfxctl/errutil"
	"github.com/siemens/wfx/cmd/wfxctl/flags"
	"github.com/siemens/wfx/generated/model"
	"github.com/stretchr/testify/assert"
)

func TestSubscribeJobStatus(t *testing.T) {
	const expectedPath = "/api/wfx/v1/jobs/1/status/subscribe"
	var actualPath string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actualPath = r.URL.Path

		w.Header().Add("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`data: "hello world"

`))
	}))
	defer ts.Close()

	u, _ := url.Parse(ts.URL)
	_ = flags.Koanf.Set(flags.ClientHostFlag, u.Hostname())
	port, _ := strconv.Atoi(u.Port())
	_ = flags.Koanf.Set(flags.ClientPortFlag, port)

	_ = flags.Koanf.Set(idFlag, "1")

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

	errResp := model.ErrorResponse{
		Errors: []*model.Error{
			&api.JobTerminalState,
		},
	}

	rec := httptest.NewRecorder()
	rec.WriteHeader(http.StatusBadRequest)
	_, _ = rec.Write(errutil.Must(json.Marshal(&errResp)))

	resp := rec.Result()
	err := validator(out)(resp)
	assert.NotNil(t, err)

	assert.Equal(t, "ERROR: The request was invalid because the job is in a terminal state (code=wfx.jobTerminalState, logref=916f0a913a3e4a52a96bd271e029c201)\n", out.String())
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
