package query

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/siemens/wfx/cmd/wfxctl/flags"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueryJobs(t *testing.T) {
	const expectedPath = "/api/wfx/v1/jobs"
	var actualPath string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actualPath = r.URL.Path

		values := r.URL.Query()

		assert.Equal(t, "my_client", values.Get("clientId"))
		assert.Equal(t, []string{"OPEN"}, values["group"])
		assert.Equal(t, "DOWNLOAD", values.Get("state"))
		assert.Equal(t, "wfx.workflow.dau.direct", values.Get("workflow"))

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{}"))
	}))
	defer ts.Close()

	u, _ := url.Parse(ts.URL)
	t.Setenv("WFX_CLIENT_HOST", u.Hostname())
	t.Setenv("WFX_CLIENT_PORT", u.Port())

	cmd := NewCommand()
	cmd.SetArgs([]string{"--" + flags.ClientIDFlag, "my_client", "--" + flags.GroupFlag, "OPEN", "--" + flags.StateFlag, "DOWNLOAD", "--" + flags.WorkflowFlag, "wfx.workflow.dau.direct"})
	err := cmd.Execute()
	require.NoError(t, err)
	assert.Equal(t, expectedPath, actualPath)
}
