package updatestatus

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/siemens/wfx/cmd/wfxctl/flags"
	"github.com/siemens/wfx/generated/model"
	"github.com/stretchr/testify/assert"
)

func TestUpdateJobStatus(t *testing.T) {
	var actualPath string
	var body []byte

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actualPath = r.URL.Path
		body, _ = io.ReadAll(r.Body)

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{}"))
	}))
	defer ts.Close()

	u, _ := url.Parse(ts.URL)
	_ = flags.Koanf.Set(flags.ClientHostFlag, u.Hostname())
	port, _ := strconv.Atoi(u.Port())
	_ = flags.Koanf.Set(flags.ClientPortFlag, port)

	_ = flags.Koanf.Set(clientIDFlag, "foo")
	_ = flags.Koanf.Set(messageFlag, "this is a test")
	_ = flags.Koanf.Set(progressFlag, int32(42))
	_ = flags.Koanf.Set(stateFlag, "DOWNLOADED")
	_ = flags.Koanf.Set(idFlag, "1")
	_ = flags.Koanf.Set(actorFlag, model.EligibleEnumCLIENT)

	err := Command.Execute()
	assert.NoError(t, err)

	assert.Equal(t, "/api/wfx/v1/jobs/1/status", actualPath)
	assert.JSONEq(t, `{"clientId": "foo", "message":"this is a test","progress":42,"state":"DOWNLOADED"}`, string(body))
}
