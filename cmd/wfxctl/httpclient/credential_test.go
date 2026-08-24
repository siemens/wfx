package httpclient

/*
 * SPDX-FileCopyrightText: 2026 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func credentialHelper(t *testing.T, response string) (string, string) {
	t.Helper()
	dir := t.TempDir()
	input := filepath.Join(dir, "input")
	helper := filepath.Join(dir, "helper")
	script := "#!/bin/sh\ncat >" + input + "\nprintf '%s' '" + response + "'\n"
	require.NoError(t, os.WriteFile(helper, []byte(script), 0o700))
	return helper, input
}

func TestCredentialRequestEditorBasic(t *testing.T) {
	helper, input := credentialHelper(t, "username=user\npassword=secret\n\n")
	editor, err := RequestEditor(nil, helper)
	require.NoError(t, err)
	req, err := http.NewRequest(http.MethodGet, "https://example.com:8443/api/wfx/v1/jobs", nil)
	require.NoError(t, err)

	require.NoError(t, editor(req.Context(), req))
	assert.Equal(t, "Basic dXNlcjpzZWNyZXQ=", req.Header.Get("Authorization"))
	payload, err := os.ReadFile(input)
	require.NoError(t, err)
	assert.Equal(t, "protocol=https\nhost=example.com:8443\npath=api/wfx/v1/jobs\n\n", string(payload))
}

func TestCredentialRequestEditorAuthType(t *testing.T) {
	helper, _ := credentialHelper(t, "authtype=Bearer\ncredential=token\n\n")
	editor, err := RequestEditor(nil, helper)
	require.NoError(t, err)
	req, err := http.NewRequest(http.MethodGet, "https://example.com/api", nil)
	require.NoError(t, err)

	require.NoError(t, editor(req.Context(), req))
	assert.Equal(t, "Bearer token", req.Header.Get("Authorization"))
}

func TestCredentialRequestEditorPreservesAuthorization(t *testing.T) {
	editor, err := RequestEditor(nil, "missing-helper")
	require.NoError(t, err)
	req, err := http.NewRequest(http.MethodGet, "https://example.com/api", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer explicit")

	require.NoError(t, editor(req.Context(), req))
	assert.Equal(t, "Bearer explicit", req.Header.Get("Authorization"))
}

func TestCredentialRequestEditorIncompleteCredential(t *testing.T) {
	helper, _ := credentialHelper(t, "username=user\n\n")
	editor, err := RequestEditor(nil, helper)
	require.NoError(t, err)
	req, err := http.NewRequest(http.MethodGet, "https://example.com/api", nil)
	require.NoError(t, err)

	require.ErrorContains(t, editor(req.Context(), req), "credential helper returned no complete credential")
}

func TestCredentialHelperName(t *testing.T) {
	assert.Equal(t, "wfxctl-credential-store", credentialHelperName("store"))
	assert.Equal(t, "/usr/bin/helper", credentialHelperName("/usr/bin/helper"))
}
