package health

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alexliesenfeld/health"
	"github.com/siemens/wfx/cmd/wfxctl/flags"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

func TestCommand_Color(t *testing.T) {
	for _, c := range []string{colorAlways, colorAuto, colorNever, "foo"} {
		_ = flags.Koanf.Set(colorFlag, c)
		err := Command.Execute()
		require.NoError(t, err)
	}
}

func TestPrettyReport_Empty(t *testing.T) {
	buf := new(bytes.Buffer)
	prettyReport(buf, false, nil)
	prettyReport(buf, true, nil)
	assert.NotEmpty(t, buf)
}

func TestPrettyReport(t *testing.T) {
	buf := new(bytes.Buffer)

	allEndpoints := []endpoint{
		{Name: "Foo", URL: "http://127.0.0.1", Status: health.StatusUp},
		{Name: "Foo", URL: "http://127.0.0.2", Status: health.StatusDown},
		{Name: "Foo", URL: "http://127.0.0.3", Status: health.StatusUnknown},
	}

	prettyReport(buf, false, allEndpoints)
	prettyReport(buf, true, allEndpoints)
	assert.NotEmpty(t, buf)
}

func TestUpdateStatus_Down(t *testing.T) {
	ep := endpoint{Name: "Foo", URL: "http://127.0.0.1", Status: health.StatusUp}
	updateStatus(&ep, &http.Client{})
	assert.Equal(t, health.StatusDown, ep.Status)
}

type mockHTTPClient struct {
	Status      health.AvailabilityStatus
	InvalidJSON bool
}

func (m mockHTTPClient) Get(string) (resp *http.Response, err error) {
	result := health.CheckerResult{
		Status: m.Status,
	}
	body, _ := json.Marshal(result)

	fakeResp := httptest.NewRecorder()
	fakeResp.WriteHeader(http.StatusOK)
	fakeResp.Header().Set("Content-Type", "application/json")
	_, _ = fakeResp.Write(body)
	if m.InvalidJSON {
		_, _ = fakeResp.Write([]byte("invalid!!!"))
	}
	return fakeResp.Result(), nil
}

func TestUpdateStatus_Up(t *testing.T) {
	ep := endpoint{Name: "Foo", URL: "http://127.0.0.1", Status: health.StatusUnknown}
	mock := mockHTTPClient{Status: health.StatusUp}
	updateStatus(&ep, mock)
	assert.Equal(t, health.StatusUp, ep.Status)
}

func TestUpdateStatus_DownInvalidJSON(t *testing.T) {
	ep := endpoint{Name: "Foo", URL: "http://127.0.0.1", Status: health.StatusUnknown}
	mock := mockHTTPClient{Status: health.StatusUp, InvalidJSON: true}
	updateStatus(&ep, mock)
	assert.Equal(t, health.StatusDown, ep.Status)
}
