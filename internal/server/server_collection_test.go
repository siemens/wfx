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
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Southclaws/fault"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/routers/gorillamux"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/siemens/wfx/api"
	"github.com/siemens/wfx/cmd/wfx/cmd/config"
	genAPI "github.com/siemens/wfx/generated/api"
	"github.com/siemens/wfx/internal/gron"
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

func TestQueryParameterNamesAreCaseInsensitive(t *testing.T) {
	dbMock := persistence.NewHealthyMockStorage(t)
	dbMock.EXPECT().QueryJobs(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(_ context.Context, filter persistence.FilterParams, _ persistence.SortParams, pagination persistence.PaginationParams) (*genAPI.PaginatedJobList, error) {
			require.NotNil(t, filter.State)
			require.Equal(t, "READY", *filter.State)
			assert.Equal(t, int32(7), pagination.Limit)
			return new(genAPI.PaginatedJobList), nil
		})
	sc, err := NewServerCollection(new(config.AppConfig), api.NewWfxServer(dbMock), dbMock)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/wfx/v1/jobs?LiMiT=7&StAtE=READY", nil)
	sc.North.Handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestQueryParameterNamesAreCaseInsensitiveWithJobID(t *testing.T) {
	dbMock := persistence.NewHealthyMockStorage(t)
	dbMock.EXPECT().GetJob(mock.Anything, "demo", persistence.FetchParams{History: true}).
		Return(&genAPI.Job{ID: "demo"}, nil).Twice()
	sc, err := NewServerCollection(new(config.AppConfig), api.NewWfxServer(dbMock), dbMock)
	require.NoError(t, err)

	for name, handler := range map[string]http.Handler{"north": sc.North.Handler, "south": sc.South.Handler} {
		t.Run(name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/api/wfx/v1/jobs/demo?HiStOrY=true", nil)
			handler.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code)
		})
	}
}

func TestCaseInsensitiveQuery(t *testing.T) {
	doc, err := openapi3.NewLoader().LoadFromData([]byte(`{
		"openapi": "3.0.3",
		"info": {"title": "test", "version": "1"},
		"paths": {"/test": {
			"parameters": [{"$ref": "#/components/parameters/tags"}],
			"get": {
				"parameters": [{"name": "tags", "in": "query", "schema": {"type": "string"}}],
				"responses": {"200": {"description": "OK"}}
			}
		}},
		"components": {"parameters": {
			"tags": {"name": "tags", "in": "query", "required": true, "schema": {"type": "string"}}
		}}
	}`))
	require.NoError(t, err)
	require.NoError(t, doc.Validate(context.Background()))
	router, err := gorillamux.NewRouter(doc)
	require.NoError(t, err)

	for _, tc := range []struct {
		name   string
		path   string
		raw    string
		want   string
		status int
	}{
		{name: "preserve wire format", path: "/test", raw: "z=%2f&TaGs=a%20b&&tags=c+d&TaGs&Unknown=%2B", want: "z=%2f&tags=a%20b&&tags=c+d&tags&Unknown=%2B", status: http.StatusOK},
		{name: "escaped name", path: "/test", raw: "%54aGs=x%2fy", want: "tags=x%2fy", status: http.StatusOK},
		{name: "already canonical", path: "/test", raw: "tags=a%20b&z=%2f", want: "tags=a%20b&z=%2f", status: http.StatusOK},
		{name: "unknown parameter", path: "/test", raw: "OTHER=1", want: "OTHER=1", status: http.StatusOK},
		{name: "empty query", path: "/test", status: http.StatusOK},
		{name: "unknown route", path: "/other", raw: "TaGs=%zz", want: "TaGs=%zz", status: http.StatusOK},
		{name: "invalid name escape", path: "/test", raw: "%zz=x", status: http.StatusBadRequest},
		{name: "invalid value escape", path: "/test", raw: "TaGs=%zz", status: http.StatusBadRequest},
		{name: "unescaped semicolon", path: "/test", raw: "TaGs=a;b", status: http.StatusBadRequest},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path+"?"+tc.raw, nil)
			originalURL := req.URL
			called := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				assert.Equal(t, tc.want, r.URL.RawQuery)
				if tc.want != tc.raw {
					assert.NotSame(t, req, r)
					assert.NotSame(t, originalURL, r.URL)
				} else {
					assert.Same(t, req, r)
				}
				if tc.name == "preserve wire format" {
					assert.Equal(t, []string{"a b", "c d", ""}, r.URL.Query()["tags"])
				}
				w.WriteHeader(http.StatusOK)
			})
			rec := httptest.NewRecorder()
			caseInsensitiveQuery(next, router).ServeHTTP(rec, req)

			assert.Equal(t, tc.status, rec.Code)
			assert.Equal(t, tc.status == http.StatusOK, called)
			assert.Same(t, originalURL, req.URL)
			assert.Equal(t, tc.raw, req.URL.RawQuery)
		})
	}

	doc.Paths.Value("/test").Parameters[0].Value = nil
	assert.Error(t, doc.Validate(context.Background()))
}

func TestCORSNorthboundOnly(t *testing.T) {
	dbMock := persistence.NewHealthyMockStorage(t)
	dbMock.EXPECT().QueryJobs(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(new(genAPI.PaginatedJobList), nil).Twice()
	wfx := api.NewWfxServer(dbMock)

	flags := config.NewFlagset()
	require.NoError(t, flags.Parse([]string{"--" + config.CORSEnabledFlag}))
	cfg, err := config.NewAppConfig(flags)
	require.NoError(t, err)
	t.Cleanup(cfg.Stop)

	sc, err := NewServerCollection(cfg, wfx, dbMock)
	require.NoError(t, err)

	for _, tc := range []struct {
		name     string
		handler  http.Handler
		expected string
	}{
		{name: "north", handler: sc.North.Handler, expected: "*"},
		{name: "south", handler: sc.South.Handler},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/api/wfx/v1/jobs", nil)
			req.Header.Set("Origin", "https://example.com")
			tc.handler.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, tc.expected, rec.Header().Get("Access-Control-Allow-Origin"))
		})
	}
}

func TestCORSReload(t *testing.T) {
	dbMock := persistence.NewHealthyMockStorage(t)
	wfx := api.NewWfxServer(dbMock)

	cfgFile := path.Join(t.TempDir(), "config.yaml")
	writeConfig := func(contents string) {
		// Replace atomically so the watcher never reloads between truncate and write.
		tmpFile := cfgFile + ".tmp"
		require.NoError(t, os.WriteFile(tmpFile, []byte(contents), 0o600))
		require.NoError(t, os.Rename(tmpFile, cfgFile))
	}
	writeConfig("cors-enabled: false\n")
	flags := config.NewFlagset()
	require.NoError(t, flags.Parse([]string{"--" + config.ConfigFlag, cfgFile}))
	cfg, err := config.NewAppConfig(flags)
	require.NoError(t, err)
	t.Cleanup(cfg.Stop)

	sc, err := NewServerCollection(cfg, wfx, dbMock)
	require.NoError(t, err)

	corsOrigin := func(origin string) string {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodOptions, "/api/wfx/v1/jobs", nil)
		req.Header.Set("Origin", origin)
		req.Header.Set("Access-Control-Request-Method", http.MethodGet)
		sc.North.Handler.ServeHTTP(rec, req)
		return rec.Header().Get("Access-Control-Allow-Origin")
	}

	assert.Empty(t, corsOrigin("https://one.example"))
	writeConfig("cors-enabled: true\ncors-allowed-origins: [https://one.example]\n")
	require.Eventually(t, func() bool {
		return corsOrigin("https://one.example") == "https://one.example"
	}, 5*time.Second, 10*time.Millisecond)

	writeConfig("cors-enabled: true\ncors-allowed-origins: [https://two.example]\n")
	require.Eventually(t, func() bool {
		return corsOrigin("https://one.example") == "" && corsOrigin("https://two.example") == "https://two.example"
	}, 5*time.Second, 10*time.Millisecond)

	writeConfig("cors-enabled: false\n")
	require.Eventually(t, func() bool {
		return corsOrigin("https://two.example") == ""
	}, 5*time.Second, 10*time.Millisecond)
}

func TestCORSDisabledByDefault(t *testing.T) {
	dbMock := persistence.NewHealthyMockStorage(t)
	dbMock.EXPECT().QueryJobs(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(new(genAPI.PaginatedJobList), nil)
	wfx := api.NewWfxServer(dbMock)

	flags := config.NewFlagset()
	require.NoError(t, flags.Parse(nil))
	cfg, err := config.NewAppConfig(flags)
	require.NoError(t, err)
	t.Cleanup(cfg.Stop)

	sc, err := NewServerCollection(cfg, wfx, dbMock)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/wfx/v1/jobs", nil)
	req.Header.Set("Origin", "https://example.com")
	sc.North.Handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Empty(t, rec.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORSConfigurableOrigins(t *testing.T) {
	dbMock := persistence.NewHealthyMockStorage(t)
	dbMock.EXPECT().QueryJobs(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(new(genAPI.PaginatedJobList), nil).Twice()
	wfx := api.NewWfxServer(dbMock)

	flags := config.NewFlagset()
	require.NoError(t, flags.Parse([]string{
		"--" + config.CORSEnabledFlag,
		"--" + config.CORSAllowedOriginsFlag, "https://example.com",
	}))
	cfg, err := config.NewAppConfig(flags)
	require.NoError(t, err)
	t.Cleanup(cfg.Stop)

	sc, err := NewServerCollection(cfg, wfx, dbMock)
	require.NoError(t, err)

	for _, tc := range []struct {
		origin   string
		expected string
	}{
		{origin: "https://example.com", expected: "https://example.com"},
		{origin: "https://evil.example.com"},
	} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/wfx/v1/jobs", nil)
		req.Header.Set("Origin", tc.origin)
		sc.North.Handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, tc.expected, rec.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestCORSPreflightAllowedMethods(t *testing.T) {
	dbMock := persistence.NewHealthyMockStorage(t)
	wfx := api.NewWfxServer(dbMock)

	flags := config.NewFlagset()
	require.NoError(t, flags.Parse([]string{
		"--" + config.CORSEnabledFlag,
		"--" + config.CORSAllowedMethodsFlag, http.MethodGet,
	}))
	cfg, err := config.NewAppConfig(flags)
	require.NoError(t, err)
	t.Cleanup(cfg.Stop)

	sc, err := NewServerCollection(cfg, wfx, dbMock)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/api/wfx/v1/jobs", nil)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", http.MethodDelete)
	sc.North.Handler.ServeHTTP(rec, req)

	assert.Empty(t, rec.Header().Get("Access-Control-Allow-Origin"))
	assert.Empty(t, rec.Header().Get("Access-Control-Allow-Methods"))
}

func TestCORSAllowCredentials(t *testing.T) {
	dbMock := persistence.NewHealthyMockStorage(t)
	dbMock.EXPECT().QueryJobs(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(new(genAPI.PaginatedJobList), nil)
	wfx := api.NewWfxServer(dbMock)

	flags := config.NewFlagset()
	require.NoError(t, flags.Parse([]string{
		"--" + config.CORSEnabledFlag,
		"--" + config.CORSAllowedOriginsFlag, "https://example.com",
		"--" + config.CORSAllowCredentialsFlag,
	}))
	cfg, err := config.NewAppConfig(flags)
	require.NoError(t, err)
	t.Cleanup(cfg.Stop)

	sc, err := NewServerCollection(cfg, wfx, dbMock)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/wfx/v1/jobs", nil)
	req.Header.Set("Origin", "https://example.com")
	sc.North.Handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "true", rec.Header().Get("Access-Control-Allow-Credentials"))

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodOptions, "/api/wfx/v1/jobs", nil)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", http.MethodDelete)
	sc.North.Handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.Equal(t, "true", rec.Header().Get("Access-Control-Allow-Credentials"))
}

func TestCORSMaxAge(t *testing.T) {
	dbMock := persistence.NewHealthyMockStorage(t)
	wfx := api.NewWfxServer(dbMock)

	flags := config.NewFlagset()
	require.NoError(t, flags.Parse([]string{
		"--" + config.CORSEnabledFlag,
		"--" + config.CORSMaxAgeFlag, "30s",
	}))
	cfg, err := config.NewAppConfig(flags)
	require.NoError(t, err)
	t.Cleanup(cfg.Stop)

	sc, err := NewServerCollection(cfg, wfx, dbMock)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/api/wfx/v1/jobs", nil)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", http.MethodGet)
	sc.North.Handler.ServeHTTP(rec, req)

	assert.Equal(t, "30", rec.Header().Get("Access-Control-Max-Age"))
}

func TestCORSMaxAgeDisabled(t *testing.T) {
	dbMock := persistence.NewHealthyMockStorage(t)
	wfx := api.NewWfxServer(dbMock)

	flags := config.NewFlagset()
	require.NoError(t, flags.Parse([]string{"--" + config.CORSEnabledFlag}))
	cfg, err := config.NewAppConfig(flags)
	require.NoError(t, err)
	t.Cleanup(cfg.Stop)

	sc, err := NewServerCollection(cfg, wfx, dbMock)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/api/wfx/v1/jobs", nil)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", http.MethodGet)
	sc.North.Handler.ServeHTTP(rec, req)

	assert.Empty(t, rec.Header().Get("Access-Control-Max-Age"))
}

func TestCORSPreflight(t *testing.T) {
	dbMock := persistence.NewHealthyMockStorage(t)
	wfx := api.NewWfxServer(dbMock)

	flags := config.NewFlagset()
	require.NoError(t, flags.Parse([]string{"--" + config.CORSEnabledFlag}))
	cfg, err := config.NewAppConfig(flags)
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

	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.Equal(t, "*", rec.Header().Get("Access-Control-Allow-Origin"))
	assert.Contains(t, rec.Header().Get("Access-Control-Allow-Headers"), "Authorization")
}

func TestCreateServer_UseMiddlewares(t *testing.T) {
	dbMock := persistence.NewHealthyMockStorage(t)
	dbMock.EXPECT().QueryJobs(context.Background(), mock.Anything, mock.Anything, mock.Anything).Return(new(genAPI.PaginatedJobList), nil).Twice()
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

	gronResponse := httptest.NewRecorder()
	gronReq := httptest.NewRequest(http.MethodGet, "/api/wfx/v1/jobs", nil)
	gronReq.Header.Set("Accept", "application/gron")
	gron.Middleware(server.Handler).ServeHTTP(gronResponse, gronReq)
	assert.Equal(t, "application/gron", gronResponse.Header().Get("Content-Type"))
	assert.Contains(t, gronResponse.Body.String(), "json = {};\n")

	assert.True(t, myMWCalled.Load())
}

type failingServer struct {
	genAPI.StrictServerInterface
	err error
}

func (server failingServer) GetHealth(context.Context, genAPI.GetHealthRequestObject) (genAPI.GetHealthResponseObject, error) {
	return nil, server.err
}

func TestCreateServer_InternalErrorIsLoggedWithoutLeaking(t *testing.T) {
	var logs syncBuffer
	originalLogger := log.Logger
	originalMarshaler := zerolog.ErrorStackMarshaler
	t.Cleanup(func() {
		log.Logger = originalLogger
		zerolog.ErrorStackMarshaler = originalMarshaler //nolint:reassign // restore global test state
	})
	log.Logger = zerolog.New(&logs)
	zerolog.ErrorStackMarshaler = func(err error) any { return fault.Flatten(err) } //nolint:reassign // configure global test state

	underlying := "database password rejected"
	server, err := createServer(new(config.AppConfig), failingServer{err: fault.Wrap(errors.New(underlying))}, http.NewServeMux(), nil, nil)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	server.Handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health", nil))

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Equal(t, "Internal Server Error\n", rec.Body.String())
	assert.NotContains(t, rec.Body.String(), underlying)
	assert.Contains(t, logs.String(), underlying)
	assert.Contains(t, logs.String(), `"stack"`)
	assert.Contains(t, logs.String(), "server_collection_test.go")
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

func TestTopLevelUsesForwardedProtocol(t *testing.T) {
	mux := createMux(new(config.AppConfig), "/api/wfx/v1", false)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-Proto", "https")
	mux.ServeHTTP(rec, req)

	assert.Equal(t, `<https://example.com/api/wfx/v1/openapi.json>; rel="service-desc"`, rec.Header().Get("Link"))
}

func TestTopLevelRejectsInvalidForwardedProtocol(t *testing.T) {
	mux := createMux(new(config.AppConfig), "/api/wfx/v1", false)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-Proto", "javascript")
	mux.ServeHTTP(rec, req)

	assert.Equal(t, `<http://example.com/api/wfx/v1/openapi.json>; rel="service-desc"`, rec.Header().Get("Link"))
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
