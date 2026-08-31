package ui

/*
 * SPDX-FileCopyrightText: 2026 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
)

type OAuthSettings struct {
	IssuerURL string
	ClientID  string
	Scope     string
}

type TemplateData struct {
	AppCSS   string
	AppMJS   string
	WfxURL   string
	BasePath string
	OAuth    OAuthSettings
}

//go:embed index.html.tmpl
var indexTmpl string

func RenderIndex(w io.Writer, data TemplateData) error {
	tmpl, err := template.New("index").Funcs(template.FuncMap{
		"json": func(value string) (template.JS, error) {
			encoded, err := json.Marshal(value)
			return template.JS(encoded), err
		},
	}).Parse(indexTmpl)
	if err != nil {
		return fmt.Errorf("parse index template: %w", err)
	}
	if err := tmpl.Execute(w, data); err != nil {
		return fmt.Errorf("execute index template: %w", err)
	}
	return nil
}
