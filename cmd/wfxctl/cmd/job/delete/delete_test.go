package delete

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
	"net/url"
	"strconv"
	"testing"

	"github.com/siemens/wfx/cmd/wfxctl/flags"
	"github.com/stretchr/testify/assert"
)

func TestCreateJob(t *testing.T) {
	var actualPath string
	var actualMethod string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actualPath = r.URL.Path
		actualMethod = r.Method
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	u, _ := url.Parse(ts.URL)
	_ = flags.Koanf.Set(flags.MgmtHostFlag, u.Hostname())
	port, _ := strconv.Atoi(u.Port())
	_ = flags.Koanf.Set(flags.MgmtPortFlag, port)
	_ = flags.Koanf.Set(idFlag, "1")

	err := Command.Execute()
	assert.NoError(t, err)

	const expectedPath = "/api/wfx/v1/jobs/1"
	assert.Equal(t, expectedPath, actualPath)
	assert.Equal(t, http.MethodDelete, actualMethod)
}
