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
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/siemens/wfx/cmd/wfxctl/flags"
	"github.com/stretchr/testify/assert"
)

func TestCreateJob(t *testing.T) {
	const data = `{"artifacts":[{"name":"example.swu","uri":"http://localhost:8080/download/example.swu"}]}`
	expected := fmt.Sprintf(`{"definition":%s,"clientId":"my_client","workflow":"wfx.workflow.dau.direct","tags":[]}`, data)

	var body []byte

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ = io.ReadAll(r.Body)

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("{}"))
	}))
	defer ts.Close()

	u, _ := url.Parse(ts.URL)
	_ = flags.Koanf.Set(flags.MgmtHostFlag, u.Hostname())
	port, _ := strconv.Atoi(u.Port())
	_ = flags.Koanf.Set(flags.MgmtPortFlag, port)

	_ = flags.Koanf.Set(clientIDFlag, "my_client")
	_ = flags.Koanf.Set(workflowFlag, "wfx.workflow.dau.direct")

	var stdin bytes.Buffer
	stdin.Write([]byte(data))

	Command.SetIn(&stdin)
	Command.SetArgs([]string{"-"})
	err := Command.Execute()
	assert.NoError(t, err)

	assert.JSONEq(t, expected, string(body))
}
