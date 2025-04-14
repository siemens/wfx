package config

/*
 * SPDX-FileCopyrightText: 2024 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/Southclaws/fault"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/posflag"
	"github.com/knadh/koanf/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/pflag"
)

type AppConfig struct {
	mutex         sync.RWMutex
	k             *koanf.Koanf
	flags         *pflag.FlagSet
	fileProviders []*file.File

	// flags
	logLevel         zerolog.Level
	logFormat        string
	storage          string
	storageOpts      string
	gracefulTimeout  time.Duration
	schemes          []Scheme
	simpleFileServer string

	ssePingInterval  time.Duration
	sseGraceInterval time.Duration

	maxHeaderSize  int
	readTimeout    time.Duration
	writeTimeout   time.Duration
	keepAlive      bool
	cleanupTimeout time.Duration

	tlsCACertificate string
	tlsCertificate   string
	tlsKey           string

	clientHost       string
	clientPort       int
	clientTLSHost    string
	clientTLSPort    int
	clientUnixSocket string
	clientPluginsDir string

	mgmtHost       string
	mgmtPort       int
	mgmtTLSHost    string
	mgmtTLSPort    int
	mgmtUnixSocket string
	mgmtPluginsDir string
}

type Scheme int

const (
	SchemeHTTP Scheme = iota
	SchemeHTTPS
	SchemeUnix
)

func (scheme Scheme) String() string {
	return []string{"http", "https", "unix"}[scheme]
}

func NewAppConfig(flags *pflag.FlagSet) (*AppConfig, error) {
	k := koanf.New(".")
	knownOptions := make(map[string]bool, 64)
	flags.VisitAll(func(flag *pflag.Flag) {
		knownOptions[flag.Name] = true
	})

	mergeFn := koanf.WithMergeFunc(func(src, dest map[string]any) error {
		// merge src into dest
		for k, v := range src {
			if _, exists := knownOptions[k]; !exists {
				fmt.Fprintf(os.Stderr, "WARN: Ignoring unknown config option '%s'", k)
				continue
			}
			dest[k] = v
		}
		return nil
	})

	// Load the config files provided in the commandline and set up file watches
	cFiles, _ := flags.GetStringSlice(ConfigFlag)
	fileProviders := make([]*file.File, 0, len(cFiles))
	for _, fname := range cFiles {
		if _, err := os.Stat(fname); err == nil {
			fp := file.Provider(fname)
			if err := k.Load(fp, yaml.Parser(), mergeFn); err != nil {
				return nil, fault.Wrap(err)
			}
			fileProviders = append(fileProviders, fp)
		}
	}

	envProvider := env.Provider("WFX_", ".", func(s string) string {
		// WFX_LOG_LEVEL becomes log-level
		return strings.ReplaceAll(strings.ToLower(strings.TrimPrefix(s, "WFX_")), "_", "-")
	})
	if err := k.Load(envProvider, nil, mergeFn); err != nil {
		fmt.Fprintln(os.Stderr, "ERROR: Could not load env variables")
	}
	if err := k.Load(posflag.Provider(flags, ".", k), nil); err != nil {
		fmt.Fprintln(os.Stderr, "ERROR: Could not load CLI flags")
	}

	cfg := new(AppConfig)
	cfg.flags = flags
	cfg.k = k
	if ok := cfg.Reload(); !ok {
		return nil, errors.New("configuration contains errors")
	}

	// start watching config
	for _, fp := range fileProviders {
		if err := fp.Watch(func(_ interface{}, err error) {
			if err != nil {
				return
			}
			if err := k.Load(fp, yaml.Parser(), mergeFn); err == nil {
				if ok := cfg.Reload(); !ok {
					log.Error().Err(err).Msg("Failed to reload config")
				}
			}
		}); err != nil {
			log.Error().Err(err).Msg("Failed to set up config file watcher")
		}
	}
	cfg.fileProviders = fileProviders
	return cfg, nil
}

func (cfg *AppConfig) Stop() {
	for _, fp := range cfg.fileProviders {
		_ = fp.Unwatch()
	}
}

func (cfg *AppConfig) LogLevel() zerolog.Level {
	cfg.mutex.RLock()
	defer cfg.mutex.RUnlock()
	return cfg.logLevel
}

func (cfg *AppConfig) LogFormat() string {
	cfg.mutex.RLock()
	defer cfg.mutex.RUnlock()
	return cfg.logFormat
}

func (cfg *AppConfig) Storage() string {
	cfg.mutex.RLock()
	defer cfg.mutex.RUnlock()
	return cfg.storage
}

func (cfg *AppConfig) StorageOptions() string {
	name := cfg.Storage()

	cfg.mutex.RLock()
	defer cfg.mutex.RUnlock()

	storageOpt := cfg.flags.Lookup(StorageOptFlag)
	changed := storageOpt != nil && storageOpt.Changed
	// do not return SQLite options for non-SQLite backends
	if name != preferedStorage && (!changed || cfg.storageOpts == sqliteDefaultOpts) {
		return ""
	}
	return cfg.storageOpts
}

func (cfg *AppConfig) GracefulTimeout() time.Duration {
	cfg.mutex.RLock()
	defer cfg.mutex.RUnlock()
	return cfg.gracefulTimeout
}

func (cfg *AppConfig) Schemes() []Scheme {
	cfg.mutex.RLock()
	defer cfg.mutex.RUnlock()
	return cfg.schemes
}

func (cfg *AppConfig) SimpleFileserver() string {
	cfg.mutex.RLock()
	defer cfg.mutex.RUnlock()
	return cfg.simpleFileServer
}

func (cfg *AppConfig) Reload() bool {
	ok := true
	fmt.Fprintln(os.Stderr, "Reloading config")

	cfg.mutex.Lock()
	defer cfg.mutex.Unlock()

	cfg.logFormat = cfg.k.String(LogFormatFlag)
	cfg.storage = cfg.k.String(StorageFlag)
	cfg.storageOpts = cfg.k.String(StorageOptFlag)
	cfg.gracefulTimeout = cfg.k.Duration(GracefulTimeoutFlag)
	cfg.ssePingInterval = cfg.k.Duration(SSEPingIntervalFlag)
	cfg.sseGraceInterval = cfg.k.Duration(SSEGraceIntervalFlag)

	if schemes := cfg.k.Strings(SchemeFlag); len(schemes) > 0 {
		cfg.schemes = make([]Scheme, 0, len(schemes))
		for _, s := range schemes {
			switch s {
			case "http":
				cfg.schemes = append(cfg.schemes, SchemeHTTP)
			case "https":
				cfg.schemes = append(cfg.schemes, SchemeHTTPS)
			case "unix":
				cfg.schemes = append(cfg.schemes, SchemeUnix)
			default:
				log.Error().Str("scheme", s).Msg("Unknown scheme")
				ok = false
			}
		}
	}

	cfg.tlsCACertificate = cfg.k.String(TLSCaFlag)
	cfg.tlsCertificate = cfg.k.String(TLSCertificateFlag)
	cfg.tlsKey = cfg.k.String(TLSKeyFlag)

	cfg.maxHeaderSize = cfg.k.Int(MaxHeaderSizeFlag)
	cfg.readTimeout = cfg.k.Duration(ReadTimeoutFlag)
	cfg.writeTimeout = cfg.k.Duration(WriteTimoutFlag)
	cfg.cleanupTimeout = cfg.k.Duration(CleanupTimeoutFlag)
	cfg.keepAlive = cfg.k.Bool(KeepAliveFlag)

	cfg.mgmtHost = cfg.k.String(MgmtHostFlag)
	cfg.mgmtPort = cfg.k.Int(MgmtPortFlag)
	cfg.mgmtTLSHost = cfg.k.String(MgmtTLSHostFlag)
	cfg.mgmtTLSPort = cfg.k.Int(MgmtTLSPortFlag)
	cfg.mgmtUnixSocket = cfg.k.String(MgmtUnixSocketFlag)
	cfg.mgmtPluginsDir = cfg.k.String(MgmtPluginsDirFlag)

	cfg.clientHost = cfg.k.String(ClientHostFlag)
	cfg.clientPort = cfg.k.Int(ClientPortFlag)
	cfg.clientTLSHost = cfg.k.String(ClientTLSHostFlag)
	cfg.clientTLSPort = cfg.k.Int(ClientTLSPortFlag)
	cfg.clientUnixSocket = cfg.k.String(ClientUnixSocketFlag)
	cfg.clientPluginsDir = cfg.k.String(ClientPluginsDirFlag)

	lvlString := cfg.k.String(LogLevelFlag)
	if lvl, err := zerolog.ParseLevel(lvlString); err != nil {
		fmt.Fprintln(os.Stderr, "Failed to parse log level:", lvlString)
		ok = false
	} else {
		fmt.Fprintln(os.Stderr, "Setting global log level:", lvl)
		cfg.logLevel = lvl
		zerolog.SetGlobalLevel(lvl)
	}

	cfg.simpleFileServer = cfg.k.String(SimpleFileServerFlag)
	if cfg.simpleFileServer != "" {
		info, err := os.Stat(cfg.simpleFileServer)
		if err != nil || !info.IsDir() {
			ok = false
			fmt.Fprintf(os.Stderr, "%s is not a valid directory", cfg.simpleFileServer)
		}
	}
	return ok
}

func (cfg *AppConfig) MaxHeaderSize() int {
	cfg.mutex.RLock()
	defer cfg.mutex.RUnlock()
	return cfg.maxHeaderSize
}

func (cfg *AppConfig) ReadTimeout() time.Duration {
	cfg.mutex.RLock()
	defer cfg.mutex.RUnlock()
	return cfg.readTimeout
}

func (cfg *AppConfig) WriteTimeout() time.Duration {
	cfg.mutex.RLock()
	defer cfg.mutex.RUnlock()
	return cfg.writeTimeout
}

func (cfg *AppConfig) KeepAlive() bool {
	cfg.mutex.RLock()
	defer cfg.mutex.RUnlock()
	return cfg.keepAlive
}

func (cfg *AppConfig) CleanupTimeout() time.Duration {
	cfg.mutex.RLock()
	defer cfg.mutex.RUnlock()
	return cfg.cleanupTimeout
}

func (cfg *AppConfig) TLSCACertificate() string {
	cfg.mutex.RLock()
	defer cfg.mutex.RUnlock()
	return cfg.tlsCACertificate
}

func (cfg *AppConfig) TLSCertificate() string {
	cfg.mutex.RLock()
	defer cfg.mutex.RUnlock()
	return cfg.tlsCertificate
}

func (cfg *AppConfig) TLSKey() string {
	cfg.mutex.RLock()
	defer cfg.mutex.RUnlock()
	return cfg.tlsKey
}

func (cfg *AppConfig) ClientHost() string {
	cfg.mutex.RLock()
	defer cfg.mutex.RUnlock()
	return cfg.clientHost
}

func (cfg *AppConfig) ClientPort() int {
	cfg.mutex.RLock()
	defer cfg.mutex.RUnlock()
	return cfg.clientPort
}

func (cfg *AppConfig) ClientTLSHost() string {
	cfg.mutex.RLock()
	defer cfg.mutex.RUnlock()
	return cfg.clientTLSHost
}

func (cfg *AppConfig) ClientTLSPort() int {
	cfg.mutex.RLock()
	defer cfg.mutex.RUnlock()
	return cfg.clientTLSPort
}

func (cfg *AppConfig) ClientUnixSocket() string {
	cfg.mutex.RLock()
	defer cfg.mutex.RUnlock()
	return cfg.clientUnixSocket
}

func (cfg *AppConfig) ClientPluginsDir() string {
	cfg.mutex.RLock()
	defer cfg.mutex.RUnlock()
	return cfg.clientPluginsDir
}

func (cfg *AppConfig) MgmtHost() string {
	cfg.mutex.RLock()
	defer cfg.mutex.RUnlock()
	return cfg.mgmtHost
}

func (cfg *AppConfig) MgmtPort() int {
	cfg.mutex.RLock()
	defer cfg.mutex.RUnlock()
	return cfg.mgmtPort
}

func (cfg *AppConfig) MgmtTLSHost() string {
	cfg.mutex.RLock()
	defer cfg.mutex.RUnlock()
	return cfg.mgmtTLSHost
}

func (cfg *AppConfig) MgmtTLSPort() int {
	cfg.mutex.RLock()
	defer cfg.mutex.RUnlock()
	return cfg.mgmtTLSPort
}

func (cfg *AppConfig) MgmtUnixSocket() string {
	cfg.mutex.RLock()
	defer cfg.mutex.RUnlock()
	return cfg.mgmtUnixSocket
}

func (cfg *AppConfig) MgmtPluginsDir() string {
	cfg.mutex.RLock()
	defer cfg.mutex.RUnlock()
	return cfg.mgmtPluginsDir
}

func (cfg *AppConfig) SSEPingInterval() time.Duration {
	cfg.mutex.RLock()
	defer cfg.mutex.RUnlock()
	return cfg.ssePingInterval
}

func (cfg *AppConfig) SSEGraceInterval() time.Duration {
	cfg.mutex.RLock()
	defer cfg.mutex.RUnlock()
	return cfg.sseGraceInterval
}
