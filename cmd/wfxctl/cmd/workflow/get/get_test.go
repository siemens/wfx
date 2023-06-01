package get

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/siemens/wfx/cmd/wfxctl/flags"
	"github.com/siemens/wfx/generated/model"
	"github.com/stretchr/testify/assert"
)

func TestGetWorkflow(t *testing.T) {
	const expectedPath = "/api/wfx/v1/workflows/test"
	var actualPath string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := "test"
		wf := model.Workflow{Name: name}
		actualPath = r.URL.Path

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		model, _ := json.Marshal(wf)
		_, _ = w.Write(model)
	}))
	defer ts.Close()

	u, _ := url.Parse(ts.URL)
	_ = flags.Koanf.Set(flags.MgmtHostFlag, u.Hostname())
	port, _ := strconv.Atoi(u.Port())

	_ = flags.Koanf.Set(flags.MgmtPortFlag, port)
	_ = flags.Koanf.Set(nameFlag, "test")

	err := Command.Execute()
	assert.NoError(t, err)

	assert.Equal(t, expectedPath, actualPath)
}
