package swagger

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
	"testing"

	"github.com/steinfletcher/apitest"
	"github.com/stretchr/testify/assert"
)

func TestHint(t *testing.T) {
	mw := NewSpecMiddleware("/api", []byte(`{ "hello": "world" }`))
	result := apitest.New().
		Handler(mw.Wrap(nil)).
		Get("/").
		Expect(t).
		Status(http.StatusNotFound).
		End()
	actual, err := io.ReadAll(result.Response.Body)
	assert.NoError(t, err)
	assert.Equal(t, `The requested resource could not be found.

Hint: Check /swagger.json to see available endpoints.
`, string(actual))
}

func TestSwaggerJSON(t *testing.T) {
	mw := NewSpecMiddleware("/api", []byte(`{ "hello": "world" }`))
	result := apitest.New().
		Handler(mw.Wrap(nil)).
		Get("/api/swagger.json").
		Expect(t).
		Status(http.StatusOK).
		End()
	actual, err := io.ReadAll(result.Response.Body)
	assert.NoError(t, err)
	assert.JSONEq(t, `{ "hello": "world" }`, string(actual))
}
