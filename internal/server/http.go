package server

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"net/http"
	"time"

	"github.com/go-openapi/runtime/flagext"
)

type HTTPSettings struct {
	MaxHeaderSize  flagext.ByteSize
	KeepAlive      time.Duration
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	CleanupTimeout time.Duration
}

func NewHTTPServer(settings *HTTPSettings, handler http.Handler) *http.Server {
	server := new(http.Server)
	server.MaxHeaderBytes = int(settings.MaxHeaderSize)
	server.ReadTimeout = settings.ReadTimeout
	server.WriteTimeout = settings.WriteTimeout
	server.SetKeepAlivesEnabled(int64(settings.KeepAlive) > 0)

	if int64(settings.CleanupTimeout) > 0 {
		server.IdleTimeout = settings.CleanupTimeout
	}

	if int64(settings.CleanupTimeout) > 0 {
		server.IdleTimeout = settings.CleanupTimeout
	}

	server.Handler = handler
	return server
}
