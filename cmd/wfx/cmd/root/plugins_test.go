//go:build plugin

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
	"io"
	"os"
	"path"
	"testing"

	"github.com/knadh/koanf/v2"
	"github.com/siemens/wfx/persistence"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadPluginsEmpty(t *testing.T) {
	t.Parallel()

	dir, _ := os.MkdirTemp("", "TestLoadPluginsEmpty")
	t.Cleanup(func() {
		_ = os.Remove(dir)
	})
	plugins, err := loadPlugins(dir)
	require.NoError(t, err)
	assert.Empty(t, plugins)
}

func TestLoadPlugins(t *testing.T) {
	t.Parallel()

	dir, _ := os.MkdirTemp("", "TestLoadPlugins")
	t.Cleanup(func() {
		_ = os.RemoveAll(dir)
	})

	f, _ := os.CreateTemp(dir, "plugin")
	_ = f.Close()
	_ = os.Chmod(f.Name(), os.FileMode(0o700))

	plugins, err := loadPlugins(dir)
	require.NoError(t, err)
	assert.Len(t, plugins, 1)
	assert.Equal(t, f.Name(), plugins[0].Name())
}

func TestLoadPluginsIgnoreNonExecutable(t *testing.T) {
	t.Parallel()

	dir, _ := os.MkdirTemp("", "TestLoadPluginsIgnoreNonExecutable")
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

	baseDir, _ := os.MkdirTemp("", "TestLoadPluginsSymlink")
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
	assert.Equal(t, f.Name(), plugins[0].Name())
}

func TestLoadPluginsSymlinkIgnoreNonExecutable(t *testing.T) {
	t.Parallel()

	baseDir, _ := os.MkdirTemp("", "TestLoadPluginsSymlinkIgnoreNonExecutable")
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

func TestCreatePluginMiddlewares_InvalidDir(t *testing.T) {
	chQuit := make(chan error)
	mws, err := createPluginMiddlewares("", chQuit)
	assert.Nil(t, mws)
	assert.NotNil(t, err)
}

func TestCreatePluginMiddlewares_EmptyDir(t *testing.T) {
	baseDir, _ := os.MkdirTemp("", "TestCreatePluginMiddlewares_EmptydDir")
	t.Cleanup(func() {
		_ = os.RemoveAll(baseDir)
	})
	chQuit := make(chan error)
	mws, err := createPluginMiddlewares(baseDir, chQuit)
	assert.Empty(t, mws)
	assert.NoError(t, err)
}

func TestCreatePluginMiddlewares_PluginFailure(t *testing.T) {
	baseDir, _ := os.MkdirTemp("", "TestCreatePluginMiddlewares_PluginFailure")
	t.Cleanup(func() {
		_ = os.RemoveAll(baseDir)
	})

	f, err := os.CreateTemp(baseDir, "plugin*.sh")
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = os.Remove(f.Name())
	})
	_, _ = io.WriteString(f, "no shebang")
	fname := f.Name()
	_ = f.Close()
	_ = os.Chmod(fname, os.FileMode(0o700))

	chQuit := make(chan error)
	mws, err := createPluginMiddlewares(baseDir, chQuit)
	assert.Nil(t, mws)
	assert.Error(t, err)
}

func TestCreatePluginMiddlewares(t *testing.T) {
	baseDir, _ := os.MkdirTemp("", "TestCreatePluginMiddlewares")
	t.Cleanup(func() {
		_ = os.RemoveAll(baseDir)
	})

	f, err := os.CreateTemp(baseDir, "plugin*.sh")
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = os.Remove(f.Name())
	})
	_, _ = io.WriteString(f, `#!/bin/sh
while true; do
	sleep 1
done
`)
	fname := f.Name()
	_ = f.Close()
	_ = os.Chmod(fname, os.FileMode(0o700))

	chQuit := make(chan error)
	mws, err := createPluginMiddlewares(baseDir, chQuit)
	assert.Len(t, mws, 1)
	assert.NoError(t, err)
	mws[0].Shutdown()
}

func TestCreateNorthboundCollection_PluginsDir(t *testing.T) {
	dir, _ := os.MkdirTemp("", "TestCreateNorthboundCollection_PluginsDir.*")
	t.Cleanup(func() { os.RemoveAll(dir) })
	k.Write(func(k *koanf.Koanf) {
		_ = k.Set(mgmtPluginsDirFlag, dir)
	})
	dbMock := persistence.NewMockStorage(t)
	chQuit := make(chan error)
	sc, err := createNorthboundCollection([]string{"http"}, dbMock, chQuit)
	t.Cleanup(func() { sc.Shutdown(context.Background()) })
	assert.NoError(t, err)
	assert.NotNil(t, sc)
}

func TestCreateNorthboundCollection_PluginsDirError(t *testing.T) {
	k.Write(func(k *koanf.Koanf) {
		_ = k.Set(mgmtPluginsDirFlag, "/does/not/exist")
	})
	dbMock := persistence.NewMockStorage(t)
	chQuit := make(chan error)
	sc, err := createNorthboundCollection([]string{"http"}, dbMock, chQuit)
	assert.Error(t, err)
	assert.Nil(t, sc)
}

func TestCreateSouthboundCollection_PluginsDir(t *testing.T) {
	dir, _ := os.MkdirTemp("", "TestCreateSouthboundCollection_PluginsDir.*")
	t.Cleanup(func() { os.RemoveAll(dir) })
	k.Write(func(k *koanf.Koanf) {
		_ = k.Set(clientPluginsDirFlag, dir)
	})
	dbMock := persistence.NewMockStorage(t)
	chQuit := make(chan error)
	sc, err := createSouthboundCollection([]string{"http"}, dbMock, chQuit)
	t.Cleanup(func() { sc.Shutdown(context.Background()) })
	assert.NoError(t, err)
	assert.NotNil(t, sc)
}

func TestCreateSouthboundCollection_PluginsDirError(t *testing.T) {
	k.Write(func(k *koanf.Koanf) {
		_ = k.Set(clientPluginsDirFlag, "/does/not/exist")
	})

	dbMock := persistence.NewMockStorage(t)
	chQuit := make(chan error)
	sc, err := createSouthboundCollection([]string{"http"}, dbMock, chQuit)
	assert.Error(t, err)
	assert.Nil(t, sc)
}
