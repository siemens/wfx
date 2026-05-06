//go:build !no_plugin

package server

/*
 * SPDX-FileCopyrightText: 2024 Siemens AG
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
	"github.com/rs/zerolog/log"
	"github.com/siemens/wfx/middleware/plugin"
)

func loadPlugins(dir string) ([]plugin.Plugin, error) {
	if dir == "" {
		return []plugin.Plugin{}, nil
	}
	log.Debug().Str("dir", dir).Msgf("Loading plugins from %q", dir)
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
				log.Info().Str("dest", dest).Msgf("Loading plugin %q", dest)
				result = append(result, plugin.NewFBPlugin(dest))
			} else {
				log.Debug().Str("dest", dest).Msgf("Ignoring non-executable file %q", dest)
			}
		}
	}
	sort.Slice(result, func(i int, j int) bool { return result[i].Name() < result[j].Name() })
	log.Debug().Int("count", len(result)).Msg("Loaded plugins")
	return result, nil
}
