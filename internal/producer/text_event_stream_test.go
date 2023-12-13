package producer

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

	"github.com/siemens/wfx/generated/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTextEventStreamProducer(t *testing.T) {
	prod := TextEventStreamProducer()
	event := model.JobStatus{
		ClientID: "foo",
		Message:  "hello world",
		State:    "INSTALLING",
	}

	buf := new(bytes.Buffer)
	err := prod.Produce(buf, event)
	require.NoError(t, err)
	assert.Equal(t, `data: {"clientId":"foo","message":"hello world","state":"INSTALLING"}

`, buf.String())
}
