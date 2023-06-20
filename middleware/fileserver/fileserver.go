package fileserver

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/Southclaws/fault"
	"github.com/knadh/koanf/v2"
	"github.com/siemens/wfx/internal/config"
)

const (
	SimpleFileServerFlag = "simple-fileserver"
)

func NewFileServerMiddleware(k *config.ThreadSafeKoanf, next http.Handler) (http.Handler, error) {
	var rootDir string
	k.Read(func(k *koanf.Koanf) {
		rootDir = k.String(SimpleFileServerFlag)
	})
	if rootDir != "" {
		info, err := os.Stat(rootDir)
		if err != nil {
			return nil, fault.Wrap(err)
		}
		if !info.IsDir() {
			return nil, fmt.Errorf("%s is not a directory", rootDir)
		}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/download") {
			var rootDir string
			k.Read(func(k *koanf.Koanf) {
				rootDir = k.String(SimpleFileServerFlag)
			})
			if rootDir != "" {
				http.StripPrefix("/download", http.FileServer(http.Dir(rootDir))).ServeHTTP(w, r)
				return
			}
		}
		next.ServeHTTP(w, r)
	}), nil
}
