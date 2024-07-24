package create

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/siemens/wfx/cmd/wfxctl/flags"
	"github.com/siemens/wfx/generated/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateJob(t *testing.T) {
	const data = `{"artifacts":[{"name":"example.swu","uri":"http://localhost:8080/download/example.swu"}]}`
	expected := fmt.Sprintf(`{"definition":%s,"clientId":"my_client","workflow":"wfx.workflow.dau.direct"}`, data)

	var body []byte

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ = io.ReadAll(r.Body)

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)

		id := "1"
		clientID := "my_client"
		job := api.Job{
			ID:       id,
			ClientID: clientID,
		}
		_ = json.NewEncoder(w).Encode(job)
	}))
	defer ts.Close()

	u, _ := url.Parse(ts.URL)
	t.Setenv("WFX_MGMT_HOST", u.Hostname())
	t.Setenv("WFX_MGMT_PORT", u.Port())

	var stdin bytes.Buffer
	stdin.Write([]byte(data))

	cmd := NewCommand()
	cmd.SetArgs([]string{"--" + flags.ClientIDFlag, "my_client", "--" + flags.WorkflowFlag, "wfx.workflow.dau.direct", "-"})
	cmd.SetIn(&stdin)

	err := cmd.Execute()
	require.NoError(t, err)

	assert.JSONEq(t, expected, string(body))
}
