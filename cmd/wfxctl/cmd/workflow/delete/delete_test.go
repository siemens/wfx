package delete

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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeleteWorkflow(t *testing.T) {
	const expectedPath = "/api/wfx/v1/workflows/wfx.workflow.dau.direct"
	var actualPath string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actualPath = r.URL.Path

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	u, _ := url.Parse(ts.URL)
	t.Setenv("WFX_MGMT_HOST", u.Hostname())
	t.Setenv("WFX_MGMT_PORT", u.Port())

	cmd := NewCommand()
	cmd.SetArgs([]string{"wfx.workflow.dau.direct"})

	err := cmd.Execute()
	require.NoError(t, err)

	assert.Equal(t, expectedPath, actualPath)
}
