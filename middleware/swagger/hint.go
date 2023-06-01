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
)

func NewSpecMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`The requested resource could not be found.

Hint: Check /swagger.json to see available endpoints.
`))
			return
		}
		next.ServeHTTP(w, r)
	})
}
