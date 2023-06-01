package root

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
)

type myServer struct {
	Srv  *http.Server
	Kind serverKind
}

type serverKind int

const (
	kindHTTP serverKind = iota
	kindHTTPS
	kindUnix
)

func (k serverKind) String() string {
	switch k {
	case kindHTTP:
		return "http"
	case kindHTTPS:
		return "https"
	case kindUnix:
		return "unix"
	}
	return ""
}

func parseServerKind(scheme string) (*serverKind, error) {
	var result serverKind
	switch scheme {
	case "http":
		result = kindHTTP
	case "https":
		result = kindHTTPS
	case "unix":
		result = kindUnix
	default:
		return nil, fmt.Errorf("unknown scheme: %s", scheme)
	}
	return &result, nil
}
