package getstatus

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

func TestGetJobStatus(t *testing.T) {
	const expectedPath = "/api/wfx/v1/jobs/1/status"
	var actualPath string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actualPath = r.URL.Path

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{}"))
	}))
	defer ts.Close()

	u, _ := url.Parse(ts.URL)
	_ = flags.Koanf.Set(flags.ClientHostFlag, u.Hostname())
	port, _ := strconv.Atoi(u.Port())
	_ = flags.Koanf.Set(flags.ClientPortFlag, port)

	_ = flags.Koanf.Set(idFlag, "1")

	err := Command.Execute()
	assert.NoError(t, err)

	assert.Equal(t, expectedPath, actualPath)
}
