package getdefinition

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
	"testing"

	"github.com/siemens/wfx/cmd/wfxctl/flags"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetJobStatus(t *testing.T) {
	const expectedPath = "/api/wfx/v1/jobs/1/definition"
	var actualPath string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actualPath = r.URL.Path

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{}"))
	}))
	defer ts.Close()

	t.Setenv("WFX_HOST", ts.URL)

	cmd := NewCommand()
	cmd.SetArgs([]string{"--" + flags.IDFlag, "1"})
	err := cmd.Execute()
	require.NoError(t, err)
	assert.Equal(t, expectedPath, actualPath)
}
