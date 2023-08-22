package logging

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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriter(t *testing.T) {
	recorder := httptest.NewRecorder()

	w := newMyResponseWriter(recorder)
	w.WriteHeader(http.StatusOK)

	assert.NotNil(t, w.Header())

	_, err := w.Write([]byte("hello world"))
	require.NoError(t, err)

	result := recorder.Result()
	assert.Equal(t, http.StatusOK, result.StatusCode)

	body, err := io.ReadAll(result.Body)
	require.NoError(t, err)
	defer result.Body.Close()

	assert.Equal(t, "hello world", string(body))
}
