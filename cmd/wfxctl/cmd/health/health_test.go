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

	"github.com/siemens/wfx/cmd/wfxctl/flags"
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

func TestNewCommand_Headers(t *testing.T) {
	requests := make(chan *http.Request, 4)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests <- r
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"body":{"status":"up"}}`))
	}))
	defer ts.Close()

	u, err := url.Parse(ts.URL)
	require.NoError(t, err)
	t.Setenv("WFX_CLIENT_HOST", u.Hostname())
	t.Setenv("WFX_CLIENT_PORT", u.Port())
	t.Setenv("WFX_MGMT_HOST", u.Hostname())
	t.Setenv("WFX_MGMT_PORT", u.Port())

	cmd := NewCommand()
	cmd.Flags().StringArray(flags.ClientHeaderFlag, nil, "")
	cmd.Flags().StringArray(flags.MgmtHeaderFlag, nil, "")
	cmd.SetArgs([]string{"--client-hdr", "X-Client: custom", "--mgmt-hdr", "X-Mgmt: custom"})
	require.NoError(t, cmd.Execute())

	seenClient, seenMgmt := false, false
	for range 2 {
		req := <-requests
		if req.Header.Get("X-Client") != "" {
			assert.Empty(t, req.Header.Get("X-Mgmt"))
			seenClient = true
		} else {
			assert.Equal(t, "custom", req.Header.Get("X-Mgmt"))
			seenMgmt = true
		}
	}
	assert.True(t, seenClient)
	assert.True(t, seenMgmt)
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
