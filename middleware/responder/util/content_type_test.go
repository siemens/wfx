package util

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
	"testing"

	"github.com/siemens/wfx/internal/producer"
	"github.com/stretchr/testify/assert"
)

func TestForceContentType(t *testing.T) {
	resp := map[string]string{"hello": "world"}
	f := ForceJSONResponse(http.StatusNotFound, resp)
	rec := httptest.NewRecorder()
	f.WriteResponse(rec, producer.JSONProducer())

	result := rec.Result()
	assert.Equal(t, "application/json", result.Header.Get("Content-Type"))
	assert.Equal(t, http.StatusNotFound, result.StatusCode)
	b, _ := io.ReadAll(result.Body)
	body := string(b)
	assert.JSONEq(t, `{"hello":"world"}`, body)
}
