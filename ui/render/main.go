package main

/*
 * SPDX-FileCopyrightText: 2026 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"os"

	"github.com/siemens/wfx/cmd/wfx/cmd/config"
	"github.com/siemens/wfx/ui"
)

func main() {
	if err := ui.RenderIndex(os.Stdout, ui.TemplateData{
		AppCSS:   "app.css",
		AppMJS:   "app.js",
		WfxURL:   "http://127.0.0.1:8081/api/wfx/v1",
		BasePath: "",
		OAuth: ui.OAuthSettings{
			IssuerURL: config.DefaultOAuthIssuerURL,
			ClientID:  config.DefaultOAuthClientID,
			Scope:     config.DefaultOAuthScope,
		},
	}); err != nil {
		panic(err)
	}
}
