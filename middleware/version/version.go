package version

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/siemens/wfx/cmd/wfx/metadata"
)

type Version struct {
	Version    string `json:"version"`
	Commit     string `json:"commit"`
	BuildDate  string `json:"buildDate"`
	APIVersion string `json:"apiVersion"`
}

func NewVersionMiddleware(next http.Handler) http.Handler {
	v := Version{
		Version:    metadata.Version,
		Commit:     metadata.Commit,
		BuildDate:  metadata.Date,
		APIVersion: metadata.APIVersion,
	}
	info, _ := json.MarshalIndent(v, "", "  ")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/version") {
			next.ServeHTTP(w, r)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(info)
	})
}
