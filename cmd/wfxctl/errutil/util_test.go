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

	"github.com/siemens/wfx/generated/client/jobs"
	"github.com/siemens/wfx/generated/model"
	"github.com/siemens/wfx/generated/northbound/restapi/operations/northbound"
	"github.com/stretchr/testify/assert"
)

var resp = model.ErrorResponse{
	Errors: []*model.Error{
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

func TestProcessErrorResponse(t *testing.T) {
	bad := jobs.NewPostJobsBadRequest()
	bad.Payload = &resp
	buf := new(bytes.Buffer)
	ProcessErrorResponse(buf, bad)
	msg := buf.String()
	assert.Equal(t, `ERROR: something went wrong (code=foo, logref=)
ERROR: oops (code=bar, logref=)
`, msg)
}

func TestExtractErrors(t *testing.T) {
	err := northbound.NewPostJobsBadRequest().WithPayload(&resp)
	messages := extractErrors(err)
	assert.NotEmpty(t, messages)
}

func TestMust(t *testing.T) {
	s := "hello world"
	assert.Equal(t, s, Must(s, nil))
}
