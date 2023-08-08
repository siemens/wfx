package version

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/steinfletcher/apitest"
	"github.com/stretchr/testify/assert"
)

func TestGetVersion(t *testing.T) {
	result := apitest.New().
		Handler(MW{}.Wrap(nil)).
		Get("/version").
		Expect(t).
		Status(http.StatusOK).
		End()

	actual, err := io.ReadAll(result.Response.Body)
	assert.NoError(t, err)

	var data map[string]string
	err = json.Unmarshal(actual, &data)
	assert.NoError(t, err)
}
