package fileserver

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"

	"github.com/knadh/koanf/v2"
	"github.com/siemens/wfx/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileServerMiddleware_Fallback(t *testing.T) {
	dir, err := os.MkdirTemp(os.TempDir(), "wfx-fileservertest")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	k := config.New()
	k.Write(func(k *koanf.Koanf) {
		_ = k.Set(SimpleFileServerFlag, dir)
	})

	handler, err := NewFileServerMiddleware(k, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, client")
	}))
	require.Nil(t, err)

	req := httptest.NewRequest(http.MethodGet, "/foo", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	res := w.Result()
	defer res.Body.Close()
	b, _ := ioutil.ReadAll(res.Body)
	assert.Equal(t, "Hello, client\n", string(b))
}

func TestFileServerMiddleware_Download(t *testing.T) {
	dir, err := os.MkdirTemp(os.TempDir(), "wfx-fileservertest")
	require.NoError(t, err)
	defer os.RemoveAll(dir)
	err = os.WriteFile(path.Join(dir, "hello"), []byte("world"), 0o644)
	require.NoError(t, err)

	k := config.New()
	k.Write(func(k *koanf.Koanf) {
		_ = k.Set(SimpleFileServerFlag, dir)
	})

	handler, err := NewFileServerMiddleware(k, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, client")
	}))
	require.Nil(t, err)

	req := httptest.NewRequest(http.MethodGet, "/download/hello", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	res := w.Result()
	defer res.Body.Close()
	b, _ := ioutil.ReadAll(res.Body)
	assert.Equal(t, "world", string(b))
}

func TestFileServerMiddleware_DirNotExist(t *testing.T) {
	k := config.New()
	k.Write(func(k *koanf.Koanf) {
		_ = k.Set(SimpleFileServerFlag, "/this/dir/does/not/exist")
	})
	handler, err := NewFileServerMiddleware(k, nil)
	assert.Nil(t, handler)
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "no such file or directory")
}

func TestFileServerMiddleware_DirIsFile(t *testing.T) {
	tmpFile, _ := os.CreateTemp(os.TempDir(), "wfx-cert-")
	_, _ = tmpFile.Write([]byte("hello world"))
	defer os.Remove(tmpFile.Name())

	k := config.New()
	k.Write(func(k *koanf.Koanf) {
		_ = k.Set(SimpleFileServerFlag, tmpFile.Name())
	})
	handler, err := NewFileServerMiddleware(k, nil)
	assert.Nil(t, handler)
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "not a directory")
}
