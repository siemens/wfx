package populate

import (
	"fmt"
	"runtime"
	"slices"
	"strings"

	"github.com/Southclaws/fault"
	"github.com/knadh/koanf/v2"
	"github.com/siemens/wfx/cmd/wfx/cmd/config"
	"github.com/siemens/wfx/persistence"
	"github.com/spf13/cobra"
)

var (
	flagCount   = "count"
	flagWorkers = "workers"
)

func NewCommand(k *koanf.Koanf) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "populate",
		Short: "Populate database with jobs",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fault.Wrap(run(cmd, k))
		},
	}

	f := cmd.Flags()
	f.Int(flagCount, 1000, "number of jobs to create")
	f.Int(flagWorkers, runtime.NumCPU(), "number of concurrent workers")

	supportedStorages := persistence.Storages()
	defaultStorage := supportedStorages[0]
	if slices.Index(supportedStorages, config.PreferedStorage) != -1 {
		defaultStorage = config.PreferedStorage
	}
	f.String(config.StorageFlag, defaultStorage, fmt.Sprintf("persistence storage. one of: [%s]", strings.Join(supportedStorages, ", ")))

	var storageOpts string
	if defaultStorage == config.PreferedStorage {
		storageOpts = config.SqliteDefaultOpts
	}
	f.String(config.StorageOptFlag, storageOpts, "storage options")

	return cmd
}
