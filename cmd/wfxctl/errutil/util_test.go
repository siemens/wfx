package errutil

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"bytes"
	"testing"

	"github.com/siemens/wfx/generated/api"
	"github.com/stretchr/testify/assert"
)

func TestProcessErrorResponse(t *testing.T) {
	resp := api.ErrorResponse{
		Errors: &[]api.Error{
			{
				Code:    "foo",
				Message: "something went wrong",
			},
			{
				Code:    "bar",
				Message: "oops",
			},
		},
	}
	buf := new(bytes.Buffer)
	ProcessErrorResponse(buf, resp)
	msg := buf.String()
	assert.Equal(t, `ERROR: something went wrong (code=foo, logref=)
ERROR: oops (code=bar, logref=)
`, msg)
}

func TestMust(t *testing.T) {
	s := "hello world"
	assert.Equal(t, s, Must(s, nil))
}
