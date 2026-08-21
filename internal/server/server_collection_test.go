package server

/*
 * SPDX-FileCopyrightText: 2024 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/siemens/wfx/api"
	"github.com/siemens/wfx/cmd/wfx/cmd/config"
	genAPI "github.com/siemens/wfx/generated/api"
	"github.com/siemens/wfx/persistence"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewServerCollection(t *testing.T) {
	dbMock := persistence.NewHealthyMockStorage(t)
	sc, err := NewServerCollection(new(config.AppConfig), nil, dbMock)
	assert.NotNil(t, sc)
	assert.NoError(t, err)
}

func TestCORSConfigurableOrigins(t *testing.T) {
	dbMock := persistence.NewHealthyMockStorage(t)
	dbMock.EXPECT().QueryJobs(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(new(genAPI.PaginatedJobList), nil)
	wfx := api.NewWfxServer(dbMock)

	f := config.NewFlagset()
	_ = f.Parse([]string{"--" + config.CORSAllowedOriginsFlag, "https://example.com"})
	cfg, err := config.NewAppConfig(f)
	require.NoError(t, err)
	t.Cleanup(cfg.Stop)

	sc, err := NewServerCollection(cfg, wfx, dbMock)
	require.NoError(t, err)

	for _, tc := range []struct {
		origin   string
		expected string
	}{
		{origin: "https://example.com", expected: "https://example.com"},
		{origin: "https://evil.example.com", expected: ""},
	} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/wfx/v1/jobs", nil)
		req.Header.Set("Origin", tc.origin)
		sc.North.Handler.ServeHTTP(rec, req)

		result := rec.Result()
		assert.Equal(t, http.StatusOK, result.StatusCode)
		assert.Equal(t, tc.expected, result.Header.Get("Access-Control-Allow-Origin"))
	}
}

func TestCORSPreflightAllowedMethods(t *testing.T) {
	dbMock := persistence.NewHealthyMockStorage(t)
	wfx := api.NewWfxServer(dbMock)

	f := config.NewFlagset()
	_ = f.Parse([]string{"--" + config.CORSAllowedMethodsFlag, "GET"})
	cfg, err := config.NewAppConfig(f)
	require.NoError(t, err)
	t.Cleanup(cfg.Stop)

	sc, err := NewServerCollection(cfg, wfx, dbMock)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/api/wfx/v1/jobs", nil)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", "DELETE")
	sc.North.Handler.ServeHTTP(rec, req)

	// disallowed method in preflight: cors middleware answers without CORS headers,
	// so the browser rejects the actual request
	result := rec.Result()
	assert.Empty(t, result.Header.Get("Access-Control-Allow-Origin"))
	assert.Empty(t, result.Header.Get("Access-Control-Allow-Methods"))
}

func TestCORSAllowCredentials(t *testing.T) {
	dbMock := persistence.NewHealthyMockStorage(t)
	dbMock.EXPECT().QueryJobs(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(new(genAPI.PaginatedJobList), nil)
	wfx := api.NewWfxServer(dbMock)

	f := config.NewFlagset()
	_ = f.Parse([]string{
		"--" + config.CORSAllowedOriginsFlag, "https://example.com",
		"--" + config.CORSAllowCredentialsFlag,
	})
	cfg, err := config.NewAppConfig(f)
	require.NoError(t, err)
	t.Cleanup(cfg.Stop)

	sc, err := NewServerCollection(cfg, wfx, dbMock)
	require.NoError(t, err)

	// actual request
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/wfx/v1/jobs", nil)
	req.Header.Set("Origin", "https://example.com")
	sc.North.Handler.ServeHTTP(rec, req)
	result := rec.Result()
	assert.Equal(t, http.StatusOK, result.StatusCode)
	assert.Equal(t, "true", result.Header.Get("Access-Control-Allow-Credentials"))

	// preflight request
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodOptions, "/api/wfx/v1/jobs", nil)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", http.MethodDelete)
	sc.North.Handler.ServeHTTP(rec, req)
	result = rec.Result()
	assert.Equal(t, http.StatusNoContent, result.StatusCode)
	assert.Equal(t, "true", result.Header.Get("Access-Control-Allow-Credentials"))
}

func TestCORSMaxAge(t *testing.T) {
	dbMock := persistence.NewHealthyMockStorage(t)
	wfx := api.NewWfxServer(dbMock)

	f := config.NewFlagset()
	_ = f.Parse([]string{"--" + config.CORSMaxAgeFlag, "30s"})
	cfg, err := config.NewAppConfig(f)
	require.NoError(t, err)
	t.Cleanup(cfg.Stop)

	sc, err := NewServerCollection(cfg, wfx, dbMock)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/api/wfx/v1/jobs", nil)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", http.MethodGet)
	sc.North.Handler.ServeHTTP(rec, req)

	assert.Equal(t, "30", rec.Result().Header.Get("Access-Control-Max-Age"))
}

func TestCORSMaxAgeDisabled(t *testing.T) {
	dbMock := persistence.NewHealthyMockStorage(t)
	wfx := api.NewWfxServer(dbMock)

	// default: no --cors-max-age flag, header must be omitted
	f := config.NewFlagset()
	_ = f.Parse(nil)
	cfg, err := config.NewAppConfig(f)
	require.NoError(t, err)
	t.Cleanup(cfg.Stop)

	sc, err := NewServerCollection(cfg, wfx, dbMock)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/api/wfx/v1/jobs", nil)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", http.MethodGet)
	sc.North.Handler.ServeHTTP(rec, req)

	assert.Empty(t, rec.Result().Header.Get("Access-Control-Max-Age"))
}

func TestCORSDefaultAllowsAnyHeader(t *testing.T) {
	dbMock := persistence.NewHealthyMockStorage(t)
	wfx := api.NewWfxServer(dbMock)

	f := config.NewFlagset()
	_ = f.Parse([]string{})
	cfg, err := config.NewAppConfig(f)
	require.NoError(t, err)
	t.Cleanup(cfg.Stop)

	sc, err := NewServerCollection(cfg, wfx, dbMock)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/api/wfx/v1/jobs", nil)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", http.MethodGet)
	req.Header.Set("Access-Control-Request-Headers", "Authorization")
	sc.North.Handler.ServeHTTP(rec, req)

	result := rec.Result()
	assert.Equal(t, "*", result.Header.Get("Access-Control-Allow-Origin"))
	assert.Contains(t, strings.ToLower(result.Header.Get("Access-Control-Allow-Headers")), "authorization")
}

func TestCreateServer_UseMiddlewares(t *testing.T) {
	dbMock := persistence.NewHealthyMockStorage(t)
	dbMock.EXPECT().QueryJobs(context.Background(), mock.Anything, mock.Anything, mock.Anything).Return(new(genAPI.PaginatedJobList), nil)
	wfx := api.NewWfxServer(dbMock)

	var myMWCalled atomic.Bool
	myMW := func(next http.Handler) http.Handler {
		myMWCalled.Store(true)
		return next
	}

	middlewares := []genAPI.MiddlewareFunc{myMW}
	cfg := new(config.AppConfig)
	mux := createMux(cfg, "/api/wfx/v1", false)
	server, err := createServer(cfg, NewNorthboundServer(wfx), mux, middlewares, nil)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("", "/api/wfx/v1/jobs", nil)

	server.Handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Result().StatusCode)

	assert.True(t, myMWCalled.Load())
}

func TestOpenAPIJSON(t *testing.T) {
	mux := createMux(new(config.AppConfig), "/api/wfx/v1", false)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/wfx/v1/openapi.json", nil))
	result := rec.Result()
	assert.Equal(t, http.StatusOK, result.StatusCode)
}

func TestTopLevelNotFound(t *testing.T) {
	mux := createMux(new(config.AppConfig), "/api/wfx/v1", false)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	result := rec.Result()
	assert.Equal(t, http.StatusNoContent, result.StatusCode)
	assert.Equal(t, strings.HasSuffix(result.Header.Get("Link"), `/api/wfx/v1/openapi.json>; rel="service-desc"`), true)
}

func TestDownloadRedirect(t *testing.T) {
	mux := createMux(new(config.AppConfig), "/api/wfx/v1", false)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/download", nil))
	result := rec.Result()
	// this has changed from MovedPermanently to StatusTemporaryRedirect in Go 1.26, see https://go.dev/doc/go1.26
	assert.True(t, result.StatusCode == http.StatusTemporaryRedirect || result.StatusCode == http.StatusMovedPermanently)
	b, _ := io.ReadAll(result.Body)
	assert.Contains(t, string(b), "/download/")
}

func TestDownload(t *testing.T) {
	tmp := os.TempDir()
	tmpFile, _ := os.CreateTemp(tmp, "TestDownload.*")
	_, _ = tmpFile.Write([]byte("hello world"))
	_ = tmpFile.Close()
	t.Cleanup(func() { _ = os.Remove(tmpFile.Name()) })

	f := config.NewFlagset()
	_ = f.Parse([]string{"--" + config.SimpleFileServerFlag, tmp})
	cfg, err := config.NewAppConfig(f)
	require.NotEmpty(t, cfg.SimpleFileserver())
	require.NoError(t, err)
	t.Cleanup(cfg.Stop)

	mux := createMux(cfg, "/api/wfx/v1", false)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, fmt.Sprintf("/download/%s", path.Base(tmpFile.Name())), nil))
	result := rec.Result()
	assert.Equal(t, http.StatusOK, result.StatusCode)
	b, _ := io.ReadAll(result.Body)
	assert.Contains(t, string(b), "hello world")
}

func TestDownload_NotFound(t *testing.T) {
	f := config.NewFlagset()
	cfg, err := config.NewAppConfig(f)
	t.Cleanup(cfg.Stop)
	require.NoError(t, err)
	mux := createMux(cfg, "/api/wfx/v1", false)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/download/", nil))
	result := rec.Result()
	assert.Equal(t, http.StatusNotFound, result.StatusCode)
}
