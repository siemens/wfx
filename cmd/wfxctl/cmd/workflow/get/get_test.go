package get

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/siemens/wfx/cmd/wfxctl/flags"
	"github.com/siemens/wfx/generated/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetWorkflow(t *testing.T) {
	const expectedPath = "/api/wfx/v1/workflows/test"
	var actualPath string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := "test"
		wf := api.Workflow{Name: name}
		actualPath = r.URL.Path

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(wf)
	}))
	defer ts.Close()

	u, _ := url.Parse(ts.URL)
	t.Setenv("WFX_CLIENT_HOST", u.Hostname())
	t.Setenv("WFX_CLIENT_PORT", u.Port())

	cmd := NewCommand()
	cmd.SetArgs([]string{"--" + flags.NameFlag, "test"})

	err := cmd.Execute()

	require.NoError(t, err)
	assert.Equal(t, expectedPath, actualPath)
}
