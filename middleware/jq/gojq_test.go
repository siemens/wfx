package jq

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewJqMiddleware(t *testing.T) {
	mw := MW{}
	handler := mw.Wrap(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintln(w, "Hello, client")
	}))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/foo", nil)
	req.Header["X-Response-Filter"] = []string{".name"}
	handler.ServeHTTP(w, req)

	resp := w.Result()
	assert.Equal(t, []string{".name"}, resp.Header["X-Response-Filter"])
}
