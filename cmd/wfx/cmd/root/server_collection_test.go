package root

/*
 * SPDX-FileCopyrightText: 2024 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
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

func TestNewServerCollection(t *testing.T) {
	dbMock := persistence.NewHealthyMockStorage(t)
	sc, err := NewServerCollection(new(config.AppConfig), dbMock)
	assert.NotNil(t, sc)
	assert.NoError(t, err)
}

func TestOpenAPIJSON(t *testing.T) {
	mux := createMux(new(config.AppConfig), nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/wfx/v1/openapiv3.json", nil))
	result := rec.Result()
	assert.Equal(t, http.StatusOK, result.StatusCode)
}

func TestTopLevelNotFound(t *testing.T) {
	mux := createMux(new(config.AppConfig), nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	result := rec.Result()
	assert.Equal(t, http.StatusNotFound, result.StatusCode)
	b, _ := io.ReadAll(result.Body)
	assert.Equal(t, "The requested resource could not be found.\n\nHint: Check /api/wfx/v1/openapiv3.json to see available endpoints.\n", string(b))
}

func TestDownloadRedirect(t *testing.T) {
	mux := createMux(new(config.AppConfig), nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/download", nil))
	result := rec.Result()
	assert.Equal(t, http.StatusMovedPermanently, result.StatusCode)
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

	mux := createMux(cfg, nil)
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
	mux := createMux(cfg, nil)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/download/", nil))
	result := rec.Result()
	assert.Equal(t, http.StatusNotFound, result.StatusCode)
}
