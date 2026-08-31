package main

/*
 * SPDX-FileCopyrightText: 2026 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"flag"
	"os"

	"github.com/siemens/wfx/cmd/wfx/cmd/config"
	"github.com/siemens/wfx/ui"
)

func main() {
	var data ui.TemplateData

	flag.StringVar(&data.AppCSS, "app-css", "app.css", "path to app CSS")
	flag.StringVar(&data.AppMJS, "app-mjs", "app.js", "path to app JS module")
	flag.StringVar(&data.WfxURL, "wfx-url", "http://127.0.0.1:8081/api/wfx/v1", "wfx API base URL")
	flag.StringVar(&data.BasePath, "base-path", "", "base path for UI routing")
	flag.StringVar(&data.OAuth.Issuer, "oauth-issuer", config.DefaultOAuthIssuer, "OIDC issuer URL")
	flag.StringVar(&data.OAuth.ClientID, "oauth-client-id", config.DefaultOAuthClientID, "OAuth client ID")
	flag.StringVar(&data.OAuth.Scope, "oauth-scope", config.DefaultOAuthScope, "OAuth scope")
	flag.Parse()

	if err := ui.RenderIndex(os.Stdout, data); err != nil {
		panic(err)
	}
}
