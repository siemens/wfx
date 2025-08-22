package version

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

	"github.com/siemens/wfx/generated/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVersion(t *testing.T) {
	const expectedPath = "/api/wfx/v1/version"
	var actualPath string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actualPath = r.URL.Path

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		version := api.GetVersion200JSONResponse{
			ApiVersion: "1.0",
			Commit:     "036d0ad9eae6e3d6e4cead4558c630a49a116219",
			Version:    "0.0.0",
		}
		_ = json.NewEncoder(w).Encode(version)
	}))
	defer ts.Close()

	u, _ := url.Parse(ts.URL)
	t.Setenv("WFX_CLIENT_HOST", u.Hostname())
	t.Setenv("WFX_CLIENT_PORT", u.Port())

	cmd := NewCommand()
	err := cmd.Execute()
	require.NoError(t, err)
	assert.Equal(t, expectedPath, actualPath)
}
