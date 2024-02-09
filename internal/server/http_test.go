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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewHTTPServer(t *testing.T) {
	settings := &HTTPSettings{
		MaxHeaderSize:  1024,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   5 * time.Second,
		KeepAlive:      time.Minute,
		CleanupTimeout: 5 * time.Minute,
	}

	handler := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})
	server := NewHTTPServer(settings, handler)

	assert.Equal(t, 1024, server.MaxHeaderBytes)
	assert.Equal(t, 10*time.Second, server.ReadTimeout)
	assert.Equal(t, 5*time.Second, server.WriteTimeout)
	assert.Equal(t, 5*time.Minute, server.IdleTimeout)
}
