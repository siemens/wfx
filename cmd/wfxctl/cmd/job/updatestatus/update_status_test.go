package updatestatus

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/siemens/wfx/cmd/wfxctl/flags"
	"github.com/stretchr/testify/assert"
)

func TestUpdateJobStatus(t *testing.T) {
	var actualPath string
	var body []byte

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actualPath = r.URL.Path
		body, _ = io.ReadAll(r.Body)

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{}"))
	}))
	defer ts.Close()

	t.Setenv("WFX_HOST", ts.URL)

	cmd := NewCommand()
	cmd.SetArgs([]string{
		"--" + flags.ClientIDFlag, "foo",
		"--" + flags.MessageFlag, "this is a test",
		"--" + flags.ProgressFlag, "42",
		"--" + flags.StateFlag, "DOWNLOADED",
		"--" + flags.IDFlag, "1",
	})
	err := cmd.Execute()
	assert.NoError(t, err)

	assert.Equal(t, "/api/wfx/v1/jobs/1/status", actualPath)
	assert.JSONEq(t, `{"clientId": "foo", "message":"this is a test","progress":42,"state":"DOWNLOADED"}`, string(body))
}
