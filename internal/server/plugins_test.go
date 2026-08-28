//go:build !no_plugin

package server

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/siemens/wfx/cmd/wfx/cmd/config"
	genAPI "github.com/siemens/wfx/generated/api"
	genPlugin "github.com/siemens/wfx/generated/plugin"
	"github.com/siemens/wfx/middleware/plugin"
	pluginioutil "github.com/siemens/wfx/middleware/plugin/ioutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadPluginsEmpty(t *testing.T) {
	t.Parallel()

	dir, _ := os.MkdirTemp("", t.Name())
	t.Cleanup(func() {
		_ = os.Remove(dir)
	})
	plugins, err := loadPlugins(dir)
	require.NoError(t, err)
	assert.Empty(t, plugins)
}

func TestLoadPlugins(t *testing.T) {
	t.Parallel()

	dir, _ := os.MkdirTemp("", t.Name())
	t.Cleanup(func() {
		_ = os.RemoveAll(dir)
	})

	f, _ := os.CreateTemp(dir, "plugin")
	_ = f.Close()
	_ = os.Chmod(f.Name(), os.FileMode(0o700))

	plugins, err := loadPlugins(dir)
	require.NoError(t, err)
	assert.Len(t, plugins, 1)
	expected, _ := filepath.EvalSymlinks(f.Name())
	is, _ := filepath.EvalSymlinks(plugins[0].Name())
	assert.Equal(t, expected, is)
}

func TestLoadPluginsIgnoreNonExecutable(t *testing.T) {
	t.Parallel()

	dir, _ := os.MkdirTemp("", t.Name())
	t.Cleanup(func() {
		_ = os.RemoveAll(dir)
	})

	f, _ := os.CreateTemp(dir, "plugin")
	_ = f.Close()

	plugins, err := loadPlugins(dir)
	require.NoError(t, err)
	assert.Len(t, plugins, 0)
}

func TestLoadPluginsSymlink(t *testing.T) {
	t.Parallel()

	baseDir, _ := os.MkdirTemp("", t.Name())
	t.Cleanup(func() {
		_ = os.RemoveAll(baseDir)
	})

	first, _ := os.MkdirTemp(baseDir, "first")
	second, _ := os.MkdirTemp(baseDir, "second")

	f, _ := os.CreateTemp(first, "plugin")
	_ = f.Close()
	_ = os.Chmod(f.Name(), os.FileMode(0o700))

	// create symlink
	dest := path.Join(second, "example")
	_ = os.Symlink(f.Name(), dest)

	plugins, err := loadPlugins(second)
	require.NoError(t, err)
	assert.Len(t, plugins, 1)
	expected, _ := filepath.EvalSymlinks(f.Name())
	is, _ := filepath.EvalSymlinks(plugins[0].Name())
	assert.Equal(t, expected, is)
}

func TestLoadPluginsSymlinkIgnoreNonExecutable(t *testing.T) {
	t.Parallel()

	baseDir, _ := os.MkdirTemp("", t.Name())
	t.Cleanup(func() {
		_ = os.RemoveAll(baseDir)
	})

	first, _ := os.MkdirTemp(baseDir, "first")
	second, _ := os.MkdirTemp(baseDir, "second")

	f, _ := os.CreateTemp(first, "plugin")
	_ = f.Close()

	// create symlink
	dest := path.Join(second, "example")
	_ = os.Symlink(f.Name(), dest)

	plugins, err := loadPlugins(second)
	require.NoError(t, err)
	assert.Len(t, plugins, 0)
}

func TestLoadPlugins_EmptyArg(t *testing.T) {
	t.Parallel()

	mws, err := loadPlugins("")
	assert.Empty(t, mws)
	assert.Nil(t, err)
}

func TestLoadPlugins_EmptyDir(t *testing.T) {
	t.Parallel()

	baseDir, _ := os.MkdirTemp("", t.Name())
	t.Cleanup(func() {
		_ = os.RemoveAll(baseDir)
	})
	mws, err := loadPlugins(baseDir)
	assert.Empty(t, mws)
	assert.NoError(t, err)
}

func TestLoadPlugins_DirNotExist(t *testing.T) {
	t.Parallel()

	mws, err := loadPlugins("/does/not/exist")
	assert.Error(t, err)
	assert.Empty(t, mws)
}

func TestCreateServer_PluginModifiesHeadersBeforeBinding(t *testing.T) {
	pluginPath := path.Join(t.TempDir(), "plugin.sh")
	require.NoError(t, os.WriteFile(pluginPath, []byte("#!/bin/sh\nexec \"$PLUGIN_HELPER\" -test.run=TestPluginHelperProcess\n"), 0o700))
	t.Setenv("PLUGIN_HELPER", os.Args[0])
	p := plugin.NewFBPlugin(pluginPath)
	mw, err := plugin.NewMiddleware(p, make(chan error, 1))
	require.NoError(t, err)
	defer mw.Stop()

	var header string
	ssi := headerServer{getJobs: func(_ context.Context, request genAPI.GetJobsRequestObject) (genAPI.GetJobsResponseObject, error) {
		header = string(*request.Params.XClientID)
		return genAPI.GetJobs200JSONResponse{}, nil
	}}
	server, err := createServer(new(config.AppConfig), ssi, http.NewServeMux(), nil, []*plugin.Middleware{mw})
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/wfx/v1/jobs", nil)
	req.Header.Set("X-Client-Id", "setBeforePlugin")
	server.Handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "setByPlugin", header)
}

func TestPluginHelperProcess(t *testing.T) {
	if os.Getenv("PLUGIN_HELPER") == "" {
		return
	}
	req, err := pluginioutil.ReadRequest(os.Stdin)
	require.NoError(t, err)
	for _, header := range req.Request.Envelope {
		if header.Name == "X-Client-Id" {
			header.Values = []string{"setByPlugin"}
		}
	}
	require.NoError(t, pluginioutil.WriteResponse(os.Stdout, &genPlugin.PluginResponseT{
		Cookie: req.Cookie,
		Payload: &genPlugin.PayloadT{
			Type:  genPlugin.Payloadgenerated_plugin_client_Request,
			Value: req.Request,
		},
	}))
	os.Exit(0)
}

type headerServer struct {
	genAPI.StrictServerInterface
	getJobs func(context.Context, genAPI.GetJobsRequestObject) (genAPI.GetJobsResponseObject, error)
}

func (s headerServer) GetJobs(ctx context.Context, request genAPI.GetJobsRequestObject) (genAPI.GetJobsResponseObject, error) {
	return s.getJobs(ctx, request)
}
