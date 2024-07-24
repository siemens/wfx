package updatedefinition

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/siemens/wfx/cmd/wfxctl/flags"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateJobDefinition(t *testing.T) {
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

	u, _ := url.Parse(ts.URL)
	t.Setenv("WFX_MGMT_HOST", u.Hostname())
	t.Setenv("WFX_MGMT_PORT", u.Port())

	var stdin bytes.Buffer
	stdin.Write([]byte("{}"))

	cmd := NewCommand()
	cmd.SetArgs([]string{"--" + flags.IDFlag, "1"})
	cmd.SetIn(&stdin)
	err := cmd.Execute()
	require.NoError(t, err)
	assert.Equal(t, "/api/wfx/v1/jobs/1/definition", actualPath)
	assert.JSONEq(t, "{}", string(body))
}
