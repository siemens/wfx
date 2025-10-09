//go:build ui

package ui

/*
 * SPDX-FileCopyrightText: 2025 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/steinfletcher/apitest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/html"
)

func TestMux(t *testing.T) {
	t.Parallel()
	mux := Mux("http://localhost:1234", "/ui")
	t.Run("Index", func(t *testing.T) {
		t.Parallel()
		resp := apitest.New().
			Handler(mux).
			Get("/").
			Expect(t).
			Status(http.StatusOK).
			Header("Content-Type", "text/html; charset=utf-8").
			End()
		_, err := html.Parse(resp.Response.Body)
		assert.NoError(t, err)
	})
	t.Run("Logo", func(t *testing.T) {
		t.Parallel()
		resp := apitest.New().
			Handler(mux).
			Get("/logo.svg").
			Expect(t).
			Status(http.StatusOK).
			Header("Content-Type", "image/svg+xml").
			End()
		svgBody, err := io.ReadAll(resp.Response.Body)
		require.NoError(t, err)
		var svg any
		err = xml.Unmarshal(svgBody, &svg)
		require.NoError(t, err)
	})
	t.Run("AppCSS", func(t *testing.T) {
		t.Parallel()
		apitest.New().
			Handler(mux).
			Get(fmt.Sprintf("/%s.css", appCssHash)).
			Expect(t).
			Status(http.StatusOK).
			Header("Content-Type", "text/css").
			End()
	})
	t.Run("AppJS", func(t *testing.T) {
		t.Parallel()
		apitest.New().
			Handler(mux).
			Get(fmt.Sprintf("/%s.mjs", appJsHash)).
			Expect(t).
			Status(http.StatusOK).
			Header("Content-Type", "application/javascript").
			End()
	})
}

func TestZstdHandler(t *testing.T) {
	handler := zstHandler([]byte("this is not compressed"), "text/plain")
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestFaviconHandler(t *testing.T) {
	apitest.New().
		Handler(FaviconHandler()).
		Get("/logo.svg").
		Expect(t).
		Status(http.StatusOK).
		Header("Content-Type", "image/svg+xml").
		End()
}
