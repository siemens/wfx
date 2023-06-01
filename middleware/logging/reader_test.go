package logging

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReader(t *testing.T) {
	body := new(bytes.Buffer)
	body.WriteString("Hello world")
	req := httptest.NewRequest(http.MethodGet, "http://localhost", body)

	out := make([]byte, 1024)
	reader := newMyRequestReader(req)
	defer reader.Close()
	_, err := reader.Read(out)
	assert.NoError(t, err)
	assert.Equal(t, "Hello world", reader.requestBody.String())
}
