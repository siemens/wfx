package swagger

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"net/http"
	"path"
)

type MW struct {
	swaggerJSON []byte
	basePath    string
}

func (mw MW) Shutdown() {}

func (mw MW) Wrap(next http.Handler) http.Handler {
	swaggerPath := path.Join(mw.basePath, "swagger.json")
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			rw.WriteHeader(http.StatusNotFound)
			_, _ = rw.Write([]byte(`The requested resource could not be found.

Hint: Check /swagger.json to see available endpoints.
`))
			return
		} else if r.URL.Path == swaggerPath {
			rw.Header().Set("Content-Type", "application/json")
			rw.WriteHeader(http.StatusOK)
			_, _ = rw.Write(mw.swaggerJSON)
			return
		}
		next.ServeHTTP(rw, r)
	})
}

func NewSpecMiddleware(basePath string, swaggerJSON []byte) MW {
	return MW{basePath: basePath, swaggerJSON: swaggerJSON}
}
