package svg

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/siemens/wfx/workflow/dau"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerate(t *testing.T) {
	var body []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Hello world"))
	}))
	defer server.Close()

	gen := NewGenerator()
	f := pflag.NewFlagSet("", pflag.ContinueOnError)
	gen.RegisterFlags(f)
	_ = f.Set(krokiURLFlag, server.URL)

	buf := new(bytes.Buffer)
	err := gen.Generate(buf, dau.DirectWorkflow())
	require.NoError(t, err)
	assert.Equal(t, "Hello world\n", buf.String())
	assert.Contains(t, string(body), "@startuml")
}
