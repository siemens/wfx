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
	"path/filepath"
	"sort"

	"github.com/Southclaws/fault"
	"github.com/knadh/koanf/v2"
	"github.com/rs/zerolog/log"
	"github.com/siemens/wfx/internal/errutil"
	"github.com/siemens/wfx/middleware"
	"github.com/siemens/wfx/middleware/plugin"
)

const (
	clientPluginsDirFlag = "client-plugins-dir"
	mgmtPluginsDirFlag   = "mgmt-plugins-dir"
)

func init() {
	f := Command.PersistentFlags()

	_ = Command.MarkPersistentFlagDirname(clientPluginsDirFlag)
	f.String(clientPluginsDirFlag, "", "directory containing client plugins")

	_ = Command.MarkPersistentFlagDirname(mgmtPluginsDirFlag)
	f.String(mgmtPluginsDirFlag, "", "directory containing management plugins")
}

func LoadNorthboundPlugins(chQuit chan error) ([]middleware.IntermediateMW, error) {
	return loadPluginSet(mgmtPluginsDirFlag, chQuit)
}

func LoadSouthboundPlugins(chQuit chan error) ([]middleware.IntermediateMW, error) {
	return loadPluginSet(clientPluginsDirFlag, chQuit)
}

func loadPluginSet(flag string, chQuit chan error) ([]middleware.IntermediateMW, error) {
	var pluginsDir string
	k.Read(func(k *koanf.Koanf) {
		pluginsDir = k.String(flag)
	})
	if pluginsDir == "" {
		return []middleware.IntermediateMW{}, nil
	}
	return errutil.Wrap2(createPluginMiddlewares(pluginsDir, chQuit))
}

func createPluginMiddlewares(pluginsDir string, chQuit chan error) ([]middleware.IntermediateMW, error) {
	pluginMws, err := loadPlugins(pluginsDir)
	if err != nil {
		return nil, fault.Wrap(err)
	}
	result := make([]middleware.IntermediateMW, 0, len(pluginMws))
	for _, p := range pluginMws {
		mw, err := plugin.NewMiddleware(p, chQuit)
		if err != nil {
			return nil, fault.Wrap(err)
		}
		result = append(result, mw)
	}
	return result, nil
}

func loadPlugins(dir string) ([]plugin.Plugin, error) {
	log.Debug().Msg("Loading plugins")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fault.Wrap(err)
	}

	result := make([]plugin.Plugin, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			dest, err := filepath.EvalSymlinks(path.Join(dir, entry.Name()))
			if err != nil {
				return nil, fault.Wrap(err)
			}
			info, err := os.Stat(dest)
			if err != nil {
				return nil, fault.Wrap(err)
			}
			// check if file is executable
			if (info.Mode() & 0o111) != 0 {
				result = append(result, plugin.NewFBPlugin(dest))
			} else {
				log.Warn().Str("dest", dest).Msg("Ignoring non-executable file")
			}
		}
	}
	sort.Slice(result, func(i int, j int) bool { return result[i].Name() < result[j].Name() })
	log.Debug().Int("count", len(result)).Msg("Loaded plugins")
	return result, nil
}
