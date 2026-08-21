package httpclient

/*
 * SPDX-FileCopyrightText: 2026 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseHeader(t *testing.T) {
	for _, tc := range []struct {
		input string
		name  string
		value string
		err   bool
	}{
		{input: "Authorization: Bearer foo", name: "Authorization", value: "Bearer foo"},
		{input: "X-Foo:bar", name: "X-Foo", value: "bar"},
		{input: "X-Foo:", name: "X-Foo", value: ""},
		{input: "X-Time: 12:30", name: "X-Time", value: "12:30"},
		{input: "no-colon", err: true},
		{input: ": value", err: true},
		{input: "Bad Name: value", err: true},
		{input: "Bad(Name): value", err: true},
		{input: "X-Foo: value\r\nInjected: true", err: true},
		{input: "X-Foo: value\x00", err: true},
	} {
		t.Run(tc.input, func(t *testing.T) {
			name, value, err := parseHeader(tc.input)
			if tc.err {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.name, name)
			assert.Equal(t, tc.value, value)
		})
	}
}

func TestRequestEditor(t *testing.T) {
	editor, err := RequestEditor([]string{"X-Foo: bar", "X-Foo: baz", "X-Other:quux", "Host: example.com"})
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodGet, "http://localhost", nil)
	require.NoError(t, err)
	req.Header.Set("X-Foo", "generated")
	require.NoError(t, editor(req.Context(), req))

	assert.Equal(t, []string{"bar", "baz"}, req.Header.Values("X-Foo"))
	assert.Equal(t, "quux", req.Header.Get("X-Other"))
	assert.Equal(t, "example.com", req.Host)
	assert.Empty(t, req.Header.Values("Host"))
}

func TestRequestEditor_InvalidDoesNotLeakHeader(t *testing.T) {
	const header = "Authorization Bearer secret"
	editor, err := RequestEditor([]string{header})
	assert.Error(t, err)
	assert.NotContains(t, err.Error(), header)
	assert.Nil(t, editor)
}
