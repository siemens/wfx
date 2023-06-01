package query

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

func TestQueryWorkflows(t *testing.T) {
	const expectedPath = "/api/wfx/v1/workflows"
	var actualPath string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actualPath = r.URL.Path
		assert.Equal(t, "0", r.URL.Query().Get("offset"))
		assert.Equal(t, "10", r.URL.Query().Get("limit"))

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{}"))
	}))
	defer ts.Close()

	u, _ := url.Parse(ts.URL)
	_ = flags.Koanf.Set(flags.ClientHostFlag, u.Hostname())
	port, _ := strconv.Atoi(u.Port())
	_ = flags.Koanf.Set(flags.ClientPortFlag, port)

	_ = flags.Koanf.Set(offsetFlag, 0)
	_ = flags.Koanf.Set(limitFlag, 10)

	err := Command.Execute()
	assert.NoError(t, err)

	assert.Equal(t, expectedPath, actualPath)
}
