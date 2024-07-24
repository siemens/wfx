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

func TestCreateJob(t *testing.T) {
	var actualPath string
	var actualMethod string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actualPath = r.URL.Path
		actualMethod = r.Method
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	u, _ := url.Parse(ts.URL)
	t.Setenv("WFX_MGMT_HOST", u.Hostname())
	t.Setenv("WFX_MGMT_PORT", u.Port())

	cmd := NewCommand()
	cmd.SetArgs([]string{"--id=1"})

	err := cmd.Execute()
	require.NoError(t, err)

	const expectedPath = "/api/wfx/v1/jobs/1"
	assert.Equal(t, expectedPath, actualPath)
	assert.Equal(t, http.MethodDelete, actualMethod)
}
