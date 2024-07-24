package spec

/*
 * SPDX-FileCopyrightText: 2024 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"

	"gopkg.in/yaml.v3"
)

//go:embed wfx.openapiv3.yml
var openapiv3YAML string

func init() {
	var yamlObj map[string]any
	if err := yaml.Unmarshal([]byte(openapiv3YAML), &yamlObj); err != nil {
		panic(err)
	}
	servers := yamlObj["servers"].([]any)
	servers2 := servers[0].(map[string]any)
	basePath := servers2["url"]
	specEndpoint := fmt.Sprintf("%s/openapiv3.json", basePath)

	jsonData, err := json.Marshal(yamlObj)
	if err != nil {
		panic(err)
	}

	Handlers[fmt.Sprintf("GET %s", specEndpoint)] = http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(jsonData)
	})

	Handlers["GET /"] = http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(fmt.Sprintf("The requested resource could not be found.\n\nHint: Check %s to see available endpoints.\n", specEndpoint)))
	})
}
