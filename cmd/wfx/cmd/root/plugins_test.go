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
	"errors"
	"os"
	"path"
	"testing"

	"github.com/siemens/wfx/cmd/wfx/cmd/config"
	"github.com/siemens/wfx/middleware/plugin"
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
	plugins, err := loadPlugins(dir, make(chan error))
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

	plugins, err := loadPlugins(dir, make(chan error))
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

	plugins, err := loadPlugins(dir, make(chan error))
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

	plugins, err := loadPlugins(second, make(chan error))
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

	plugins, err := loadPlugins(second, make(chan error))
	require.NoError(t, err)
	assert.Len(t, plugins, 0)
}

func TestLoadPlugins_EmptyArg(t *testing.T) {
	chErr := make(chan error)
	mws, err := loadPlugins("", chErr)
	assert.Empty(t, mws)
	assert.Nil(t, err)
}

func TestLoadPlugins_EmptyDir(t *testing.T) {
	baseDir, _ := os.MkdirTemp("", "TestCreatePluginMiddlewares_EmptydDir")
	t.Cleanup(func() {
		_ = os.RemoveAll(baseDir)
	})
	chErr := make(chan error)
	mws, err := loadPlugins(baseDir, chErr)
	assert.Empty(t, mws)
	assert.NoError(t, err)
}

func TestLoadPlugins_DirNotExist(t *testing.T) {
	chErr := make(chan error)
	mws, err := loadPlugins("/does/not/exist", chErr)
	assert.Error(t, err)
	assert.Empty(t, mws)
}

type FailingPlugin struct{}

func (FailingPlugin) Name() string {
	return "FailingPlugin"
}

func (FailingPlugin) Start() (chan plugin.Message, error) {
	return nil, errors.New("FailingPlugin cannot be started")
}

func (FailingPlugin) Stop() error {
	return nil
}

type TrivialPlugin struct{}

func (TrivialPlugin) Name() string {
	return "DummyPlugin"
}

func (TrivialPlugin) Start() (chan plugin.Message, error) {
	return make(chan plugin.Message), nil
}

func (TrivialPlugin) Stop() error {
	return nil
}

func TestCreatePluginMiddlewares_PluginFailure(t *testing.T) {
	mws, err := createPluginMiddlewares(context.Background(), []plugin.Plugin{FailingPlugin{}})
	assert.Nil(t, mws)
	assert.Error(t, err)
}

func TestCreatePluginMiddlewares(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	mws, err := createPluginMiddlewares(ctx, []plugin.Plugin{TrivialPlugin{}})
	assert.Len(t, mws, 1)
	assert.NoError(t, err)
}

func TestNewServerCollection(t *testing.T) {
	dbMock := persistence.NewHealthyMockStorage(t)
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	sc, err := NewServerCollection(ctx, new(config.AppConfig), dbMock, make(chan error))
	assert.NoError(t, err)
	assert.NotNil(t, sc)
}
