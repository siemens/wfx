//go:build !ui

package ui

/*
 * SPDX-FileCopyrightText: 2025 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import "net/http"

const Enabled = false

func Mux(string, string) *http.ServeMux {
	return http.NewServeMux()
}

func FaviconHandler() http.Handler {
	return nil
}
