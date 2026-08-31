package ui

/*
 * SPDX-FileCopyrightText: 2026 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderIndex(t *testing.T) {
	var output bytes.Buffer
	err := RenderIndex(&output, TemplateData{
		AppCSS:   "app.css",
		AppMJS:   "app.js",
		WfxURL:   "http://127.0.0.1:8081/api/wfx/v1",
		BasePath: "",
		OAuth: OAuthSettings{
			IssuerURL: "https://id.example",
			ClientID:  "</script>",
			Scope:     "openid email profile",
		},
	})
	require.NoError(t, err)
	assert.Contains(t, output.String(), `href="app.css"`)
	assert.Contains(t, output.String(), `src="app.js"`)
	assert.Contains(t, output.String(), `"wfx_url": "http://127.0.0.1:8081/api/wfx/v1"`)
	assert.Contains(t, output.String(), `"base_path": ""`)
	assert.Contains(t, output.String(), `"oauth": {`)
	assert.Contains(t, output.String(), `"issuer_url": "https://id.example"`)
	assert.Contains(t, output.String(), `"client_id": "\u003c/script\u003e"`)
	assert.Contains(t, output.String(), `"scope": "openid email profile"`)

	output.Reset()
	err = RenderIndex(&output, TemplateData{})
	require.NoError(t, err)
	assert.Contains(t, output.String(), `"oauth": null`)
}
