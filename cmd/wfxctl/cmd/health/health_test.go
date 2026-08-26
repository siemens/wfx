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
	"testing"

	"github.com/siemens/wfx/cmd/wfxctl/flags"
	"github.com/siemens/wfx/generated/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCommand_Down(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(api.CheckerResult{Status: api.Down})
	}))
	defer ts.Close()
	t.Setenv("WFX_HOST", ts.URL)

	cmd := NewCommand()
	err := cmd.Execute()
	require.ErrorContains(t, err, "wfx is not healthy")
}

func TestNewCommand_RequestError(t *testing.T) {
	ts := httptest.NewServer(nil)
	ts.Close()
	t.Setenv("WFX_HOST", ts.URL)

	cmd := NewCommand()
	require.Error(t, cmd.Execute())
}

func TestNewCommand_Up(t *testing.T) {
	var actualPath string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actualPath = r.URL.Path

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(api.CheckerResult{Status: api.Up})
	}))
	defer ts.Close()

	t.Setenv("WFX_HOST", ts.URL)

	cmd := NewCommand()
	err := cmd.Execute()
	require.NoError(t, err)
	assert.Equal(t, "/api/wfx/v1/health", actualPath)
}

func TestNewCommand_Headers(t *testing.T) {
	var actualHeader string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actualHeader = r.Header.Get("X-Custom")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"up"}`))
	}))
	defer ts.Close()

	t.Setenv("WFX_HOST", ts.URL)

	cmd := NewCommand()
	cmd.Flags().StringArray(flags.HeaderFlag, nil, "")
	cmd.SetArgs([]string{"--header", "X-Custom: value"})
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "value", actualHeader)
}

func TestNewCommand_HidesBasicAuth(t *testing.T) {
	var authorization string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorization = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"up"}`))
	}))
	defer ts.Close()

	host := "http://user:secret@" + ts.Listener.Addr().String()
	t.Setenv("WFX_HOST", host)
	var output bytes.Buffer
	cmd := NewCommand()
	cmd.SetOut(&output)
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "Basic dXNlcjpzZWNyZXQ=", authorization)
	assert.NotContains(t, output.String(), "user")
	assert.NotContains(t, output.String(), "secret")
	assert.Contains(t, output.String(), ts.Listener.Addr().String())
}

func TestNewCommand_ColorModes(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(api.CheckerResult{Status: api.Up})
	}))
	defer ts.Close()
	t.Setenv("WFX_HOST", ts.URL)

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
	prettyReport(buf, false, Endpoint{})
	prettyReport(buf, true, Endpoint{})
	assert.NotEmpty(t, buf)
}

func TestPrettyReport(t *testing.T) {
	for _, endpoint := range []Endpoint{
		{Name: "Foo", Server: "http://127.0.0.1", Response: &api.GetHealthResponse{JSON200: &api.CheckerResult{Status: api.Up}}},
		{Name: "Foo", Server: "http://127.0.0.2", Response: &api.GetHealthResponse{JSON503: &api.CheckerResult{Status: api.Down}}},
		{Name: "Foo", Server: "http://127.0.0.3", Response: &api.GetHealthResponse{JSON503: &api.CheckerResult{Status: api.Unknown}}},
	} {
		buf := new(bytes.Buffer)
		prettyReport(buf, false, endpoint)
		prettyReport(buf, true, endpoint)
		assert.NotEmpty(t, buf)
	}
}
