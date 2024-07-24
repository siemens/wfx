package addtags

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/siemens/wfx/cmd/wfxctl/flags"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddTags(t *testing.T) {
	var actualPath string
	var body []byte

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actualPath = r.URL.Path
		body, _ = io.ReadAll(r.Body)

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(body)
	}))
	defer ts.Close()

	u, _ := url.Parse(ts.URL)
	t.Setenv("WFX_MGMT_HOST", u.Hostname())
	t.Setenv("WFX_MGMT_PORT", u.Port())

	cmd := NewCommand()
	cmd.SetArgs([]string{
		fmt.Sprintf("--%s=%s", flags.IDFlag, "1"),
		"foo", "bar",
	})

	err := cmd.Execute()
	require.NoError(t, err)

	assert.Equal(t, "/api/wfx/v1/jobs/1/tags", actualPath)
	assert.JSONEq(t, `["foo", "bar"]`, string(body))
}
