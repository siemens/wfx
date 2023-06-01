package addtags

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
	"github.com/stretchr/testify/assert"
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
	_ = flags.Koanf.Set(flags.MgmtHostFlag, u.Hostname())
	port, _ := strconv.Atoi(u.Port())
	_ = flags.Koanf.Set(flags.MgmtPortFlag, port)

	_ = flags.Koanf.Set(idFlag, "1")
	Command.SetArgs([]string{"foo", "bar"})

	err := Command.Execute()
	assert.NoError(t, err)

	assert.Equal(t, "/api/wfx/v1/jobs/1/tags", actualPath)
	assert.JSONEq(t, `["foo", "bar"]`, string(body))
}
