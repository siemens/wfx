package root

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
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
	"testing"

	"github.com/siemens/wfx/cmd/wfx/cmd/config"
	"github.com/siemens/wfx/persistence"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenAPIJSON(t *testing.T) {
	storage := persistence.NewHealthyMockStorage(t)
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	sc, err := NewServerCollection(ctx, &config.AppConfig{}, storage, make(chan error))
	require.NoError(t, err)

	for _, srv := range []*http.Server{sc.North, sc.South} {

		rec := httptest.NewRecorder()
		srv.Handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/openapi.json", nil))
		result := rec.Result()
		assert.Equal(t, http.StatusOK, result.StatusCode)
	}
}

func TestTopLevelNotFound(t *testing.T) {
	storage := persistence.NewHealthyMockStorage(t)
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	sc, err := NewServerCollection(ctx, &config.AppConfig{}, storage, make(chan error))
	require.NoError(t, err)

	for _, srv := range []*http.Server{sc.North, sc.South} {

		rec := httptest.NewRecorder()
		srv.Handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
		result := rec.Result()
		assert.Equal(t, http.StatusNotFound, result.StatusCode)
		b, _ := io.ReadAll(result.Body)
		assert.Equal(t, "The requested resource could not be found.\n\nHint: Check /openapi.json to see available endpoints.\n", string(b))
	}
}

func TestDownloadRedirect(t *testing.T) {
	storage := persistence.NewHealthyMockStorage(t)
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	sc, err := NewServerCollection(ctx, &config.AppConfig{}, storage, make(chan error))
	require.NoError(t, err)

	for _, srv := range []*http.Server{sc.North, sc.South} {

		rec := httptest.NewRecorder()
		srv.Handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/download", nil))
		result := rec.Result()
		assert.Equal(t, http.StatusMovedPermanently, result.StatusCode)
		b, _ := io.ReadAll(result.Body)
		assert.Contains(t, string(b), "/download/")
	}
}

func TestDownload(t *testing.T) {
	storage := persistence.NewHealthyMockStorage(t)
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	tmp := os.TempDir()
	tmpFile, _ := os.CreateTemp(tmp, "TestDownload.*")
	_, _ = tmpFile.Write([]byte("hello world"))
	_ = tmpFile.Close()
	t.Cleanup(func() { _ = os.Remove(tmpFile.Name()) })

	f := config.NewFlagset()
	_ = f.Parse([]string{"--" + config.SimpleFileServerFlag, tmp})
	cfg, err := config.NewAppConfig(ctx, f)
	require.NotEmpty(t, cfg.SimpleFileserver())
	require.NoError(t, err)

	sc, err := NewServerCollection(ctx, cfg, storage, make(chan error))
	require.NoError(t, err)

	for _, srv := range []*http.Server{sc.North, sc.South} {
		rec := httptest.NewRecorder()
		srv.Handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, fmt.Sprintf("/download/%s", path.Base(tmpFile.Name())), nil))
		result := rec.Result()
		assert.Equal(t, http.StatusOK, result.StatusCode)
		b, _ := io.ReadAll(result.Body)
		assert.Contains(t, string(b), "hello world")
	}
}

func TestDownload_NotFound(t *testing.T) {
	storage := persistence.NewHealthyMockStorage(t)
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	f := config.NewFlagset()
	cfg, err := config.NewAppConfig(ctx, f)
	require.NoError(t, err)

	sc, err := NewServerCollection(ctx, cfg, storage, make(chan error))
	require.NoError(t, err)

	for _, srv := range []*http.Server{sc.North, sc.South} {
		rec := httptest.NewRecorder()
		srv.Handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/download/", nil))
		result := rec.Result()
		assert.Equal(t, http.StatusNotFound, result.StatusCode)
	}
}
