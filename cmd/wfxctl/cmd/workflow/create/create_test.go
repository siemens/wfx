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
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/siemens/wfx/generated/api"
	"github.com/siemens/wfx/workflow/dau"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateWorkflow_YAML(t *testing.T) {
	var body []byte

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ = io.ReadAll(r.Body)

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("{}"))
	}))
	defer ts.Close()

	u, _ := url.Parse(ts.URL)
	t.Setenv("WFX_MGMT_HOST", u.Hostname())
	t.Setenv("WFX_MGMT_PORT", u.Port())

	f, err := os.CreateTemp("", "wfx-")
	require.NoError(t, err)
	defer func() {
		_ = f.Close()
		_ = os.Remove(f.Name())
	}()

	b, err := yaml.Marshal(dau.DirectWorkflow())
	require.NoError(t, err)
	workflowYaml := string(b)

	_, _ = f.WriteString(workflowYaml)
	cmd := NewCommand()
	cmd.SetArgs([]string{f.Name()})

	err = cmd.Execute()
	require.NoError(t, err)

	var wf api.Workflow
	_ = yaml.Unmarshal([]byte(workflowYaml), &wf)
	b, _ = json.Marshal(wf)
	assert.JSONEq(t, string(b), string(body))
}

func TestReadWorkflows_None(t *testing.T) {
	_, err := readWorkflows(nil, nil)
	require.Error(t, err)
}

func TestReadWorkflows_Stdin(t *testing.T) {
	buf := new(bytes.Buffer)
	buf.WriteString(kanbanExample)

	result, err := readWorkflows([]string{"-"}, buf)
	require.NoError(t, err)
	assert.Len(t, result, 1)
}

func TestReadWorkflows_FileNotExist(t *testing.T) {
	_, err := readWorkflows([]string{"this file does not exist"}, nil)
	require.Error(t, err)
}
