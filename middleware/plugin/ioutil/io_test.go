package ioutil

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"bytes"
	"encoding/binary"
	"errors"
	"testing"
	"testing/iotest"

	"github.com/siemens/wfx/generated/plugin"
	"github.com/siemens/wfx/generated/plugin/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteAndReadRequest(t *testing.T) {
	t.Parallel()

	expected := plugin.PluginRequestT{
		Cookie: 1,
		Request: &client.RequestT{
			Action:      client.ActionRead,
			Destination: "http://localhost/foo/bar/offset=0&limit=42",
			Envelope: []*client.EnvelopeT{
				{Name: "Foo", Values: []string{"Bar", "Baz"}},
			},
		},
	}

	buf := new(bytes.Buffer)
	err := WriteRequest(buf, &expected)
	require.NoError(t, err)

	actual, err := ReadRequest(buf)

	require.NoError(t, err)
	assert.EqualValues(t, expected, *actual)
}

func TestWriteAndReadResponse(t *testing.T) {
	t.Parallel()

	expected := plugin.PluginResponseT{
		Cookie: 1,
		Payload: &plugin.PayloadT{
			Type: plugin.Payloadgenerated_plugin_client_Request,
			Value: &client.RequestT{
				Action:      client.ActionRead,
				Destination: "http://localhost/foo/bar/offset=0&limit=42",
				Envelope: []*client.EnvelopeT{
					{Name: "Foo", Values: []string{"Bar", "Baz"}},
				},
			},
		},
	}

	buf := new(bytes.Buffer)
	err := WriteResponse(buf, &expected)
	require.NoError(t, err)

	actual, err := ReadResponse(buf)

	require.NoError(t, err)
	assert.EqualValues(t, expected, *actual)
}

func TestFaultyReader(t *testing.T) {
	t.Parallel()

	myErr := errors.New("this is a fake error")
	r := iotest.ErrReader(myErr)

	t.Run("ReadRequest", func(t *testing.T) {
		t.Parallel()

		req, err := ReadRequest(r)
		assert.ErrorIs(t, err, myErr)
		assert.Nil(t, req)
	})

	t.Run("ReadResponse", func(t *testing.T) {
		t.Parallel()

		req, err := ReadResponse(r)
		assert.ErrorIs(t, err, myErr)
		assert.Nil(t, req)
	})
}

func TestIncompleteRead(t *testing.T) {
	var myInt int32 = 12345678

	buf := new(bytes.Buffer)
	_ = binary.Write(buf, binary.LittleEndian, myInt)

	_, err := readBytes(buf)
	assert.NotNil(t, err)
}
