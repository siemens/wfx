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
	"os"
	"path"
	"testing"

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
	assert.Equal(t, f.Name(), plugins[0].Name())
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
	assert.Equal(t, f.Name(), plugins[0].Name())
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
