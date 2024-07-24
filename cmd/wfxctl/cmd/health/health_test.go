package health

/*
 * SPDX-FileCopyrightText: 2024 Siemens AG
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
	"testing"

	"github.com/siemens/wfx/generated/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCommand(t *testing.T) {
	cmd := NewCommand()
	err := cmd.Execute()
	require.NoError(t, err)
}

func TestNewCommand_Up(t *testing.T) {
	var actualPath string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actualPath = r.URL.Path

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := api.GetHealth200JSONResponse{
			Body: api.CheckerResult{
				Status: api.Up,
			},
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer ts.Close()

	u, _ := url.Parse(ts.URL)
	t.Setenv("WFX_CLIENT_HOST", u.Hostname())
	t.Setenv("WFX_CLIENT_PORT", u.Port())

	cmd := NewCommand()
	err := cmd.Execute()
	require.NoError(t, err)
	assert.Equal(t, "/api/wfx/v1/health", actualPath)
}

func TestNewCommand_ColorModes(t *testing.T) {
	for _, mode := range []string{colorAlways, colorAuto, colorNever} {
		cmd := NewCommand()
		cmd.SetArgs([]string{"--color", mode})
		err := cmd.Execute()
		require.NoError(t, err)
	}
	cmd := NewCommand()
	cmd.SetArgs([]string{"--color", "foo"})
	err := cmd.Execute()
	assert.ErrorContains(t, err, "unsupported color mode: foo")
}

func TestPrettyReport_Empty(t *testing.T) {
	buf := new(bytes.Buffer)
	prettyReport(buf, false, nil)
	prettyReport(buf, true, nil)
	assert.NotEmpty(t, buf)
}

func TestPrettyReport(t *testing.T) {
	buf := new(bytes.Buffer)

	allEndpoints := []Endpoint{
		{Name: "Foo", Server: "http://127.0.0.1", Response: &api.GetHealthResponse{JSON200: &api.CheckerResult{Status: api.Up}}},
		{Name: "Foo", Server: "http://127.0.0.2", Response: &api.GetHealthResponse{JSON503: &api.CheckerResult{Status: api.Down}}},
		{Name: "Foo", Server: "http://127.0.0.3", Response: &api.GetHealthResponse{JSON503: &api.CheckerResult{Status: api.Unknown}}},
	}

	prettyReport(buf, false, allEndpoints)
	prettyReport(buf, true, allEndpoints)
	assert.NotEmpty(t, buf)
}
