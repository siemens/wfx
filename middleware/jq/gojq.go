package jq

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

type MW struct{}

func (mw MW) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		filter := r.Header.Get("X-Response-Filter")
		if filter != "" {
			w.Header().Set("X-Response-Filter", filter)
		}
		next.ServeHTTP(w, r)
	})
}

func (mw MW) Shutdown() {}
