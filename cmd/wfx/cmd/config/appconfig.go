package config

/*
 * SPDX-FileCopyrightText: 2024 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/Southclaws/fault"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env/v2"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/posflag"
	"github.com/knadh/koanf/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/siemens/wfx/persistence"
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
	simpleFileServer string

	ssePingInterval  time.Duration
	sseGraceInterval time.Duration

	maxHeaderSize int
	readTimeout   time.Duration
	writeTimeout  time.Duration
	jqOpts        JQOpts

	keepAlive      bool
	cleanupTimeout time.Duration

	corsOpts  CORSOpts
	oauthOpts OAuthOpts

	tlsCACertificate string
	tlsCertificate   string
	tlsKey           string

	clientHosts      []ListenAddr
	clientPluginsDir string

	mgmtHosts      []ListenAddr
	mgmtPluginsDir string
}

type ListenAddr struct {
	Network string
	Addr    string
	TLS     bool
}

type JQOpts struct {
	FilterTimeout         time.Duration
	FilterMaxResponseSize int
}

type CORSOpts struct {
	Enabled          bool
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	AllowCredentials bool
	MaxAge           int
}

type OAuthOpts struct {
	Issuer   string
	ClientID string
	Scope    string
}

func NewAppConfig(flags *pflag.FlagSet) (*AppConfig, error) {
	k := koanf.New(".")
	knownOptions := make(map[string]bool, 64)
	flags.VisitAll(func(flag *pflag.Flag) {
		knownOptions[flag.Name] = true
	})

	envNames := make(map[string]string)
	mergeFn := func(source string) koanf.Option {
		return koanf.WithMergeFunc(func(src, dest map[string]any) error {
			// merge src into dest
			for k, v := range src {
				if _, exists := knownOptions[k]; !exists {
					opt := k
					if source == "env" {
						if orig, ok := envNames[k]; ok {
							opt = orig
						}
					}
					fmt.Fprintf(os.Stderr, "WARN: Ignoring unknown config option '%s' from %s\n", opt, source)
					continue
				}
				dest[k] = v
			}
			return nil
		})
	}

	// Load the config files provided in the commandline and set up file watches
	cFiles, _ := flags.GetStringSlice(ConfigFlag)
	fileProviders := make([]*file.File, 0, len(cFiles))
	for _, fname := range cFiles {
		if _, err := os.Stat(fname); err == nil {
			fp := file.Provider(fname)
			if err := k.Load(fp, yaml.Parser(), mergeFn("file")); err != nil {
				return nil, fault.Wrap(err)
			}
			fileProviders = append(fileProviders, fp)
		}
	}

	envProvider := env.Provider(".", env.Opt{
		Prefix: "WFX_",
		TransformFunc: func(k string, v string) (string, any) {
			// WFX_LOG_LEVEL becomes log-level
			key := strings.ReplaceAll(strings.ToLower(strings.TrimPrefix(k, "WFX_")), "_", "-")
			envNames[key] = k
			if key == ClientHostFlag || key == MgmtHostFlag {
				return key, splitListenURLs(v)
			}
			return key, v
		},
	})
	if err := k.Load(envProvider, nil, mergeFn("env")); err != nil {
		fmt.Fprintln(os.Stderr, "ERROR: Could not load env variables")
	}
	if err := k.Load(posflag.Provider(flags, ".", k), nil); err != nil {
		fmt.Fprintln(os.Stderr, "ERROR: Could not load CLI flags")
	}

	cfg := new(AppConfig)
	cfg.flags = flags
	cfg.k = k
	if ok := cfg.Reload(); !ok {
		return nil, fault.New("configuration contains errors")
	}

	// start watching config
	for _, fp := range fileProviders {
		if err := fp.Watch(func(_ any, err error) {
			if err != nil {
				return
			}
			if err := k.Load(fp, yaml.Parser(), mergeFn("file")); err == nil {
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
	if name != PreferedStorage && (!changed || cfg.storageOpts == SqliteDefaultOpts) {
		return ""
	}
	return cfg.storageOpts
}

func (cfg *AppConfig) GracefulTimeout() time.Duration {
	cfg.mutex.RLock()
	defer cfg.mutex.RUnlock()
	return cfg.gracefulTimeout
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

	cfg.corsOpts.Enabled = cfg.k.Bool(CORSEnabledFlag)
	cfg.corsOpts.AllowedOrigins = cfg.k.Strings(CORSAllowedOriginsFlag)
	cfg.corsOpts.AllowedMethods = cfg.k.Strings(CORSAllowedMethodsFlag)
	cfg.corsOpts.AllowedHeaders = cfg.k.Strings(CORSAllowedHeadersFlag)
	cfg.corsOpts.AllowCredentials = cfg.k.Bool(CORSAllowCredentialsFlag)
	cfg.corsOpts.MaxAge = int(cfg.k.Duration(CORSMaxAgeFlag) / time.Second)
	if cfg.corsOpts.AllowCredentials && slices.Contains(cfg.corsOpts.AllowedOrigins, "*") {
		fmt.Fprintln(os.Stderr, "cors-allow-credentials requires explicit cors-allowed-origins")
		ok = false
	}

	cfg.oauthOpts.Issuer = cfg.k.String(OAuthIssuerFlag)
	cfg.oauthOpts.ClientID = cfg.k.String(OAuthClientIDFlag)
	cfg.oauthOpts.Scope = cfg.k.String(OAuthScopeFlag)

	cfg.tlsCACertificate = cfg.k.String(TLSCaFlag)
	cfg.tlsCertificate = cfg.k.String(TLSCertificateFlag)
	cfg.tlsKey = cfg.k.String(TLSKeyFlag)

	cfg.maxHeaderSize = cfg.k.Int(MaxHeaderSizeFlag)
	cfg.readTimeout = cfg.k.Duration(ReadTimeoutFlag)
	cfg.writeTimeout = cfg.k.Duration(WriteTimoutFlag)
	cfg.jqOpts.FilterTimeout = cfg.k.Duration(JQFilterTimeoutFlag)
	cfg.jqOpts.FilterMaxResponseSize = cfg.k.Int(JQFilterMaxResponseSizeFlag)
	if cfg.jqOpts.FilterTimeout < 0 {
		log.Error().Msgf("%s must not be negative", JQFilterTimeoutFlag)
		ok = false
	}
	if cfg.jqOpts.FilterMaxResponseSize < 0 {
		log.Error().Msgf("%s must not be negative", JQFilterMaxResponseSizeFlag)
		ok = false
	}
	cfg.cleanupTimeout = cfg.k.Duration(CleanupTimeoutFlag)
	cfg.keepAlive = cfg.k.Bool(KeepAliveFlag)

	cfg.mgmtPluginsDir = cfg.k.String(MgmtPluginsDirFlag)
	cfg.clientPluginsDir = cfg.k.String(ClientPluginsDirFlag)

	if addrs, err := parseListenURLs(hostValues(cfg.k, MgmtHostFlag)); err != nil {
		log.Error().Err(err).Msgf("Invalid %s", MgmtHostFlag)
		ok = false
	} else {
		cfg.mgmtHosts = addrs
	}
	if addrs, err := parseListenURLs(hostValues(cfg.k, ClientHostFlag)); err != nil {
		log.Error().Err(err).Msgf("Invalid %s", ClientHostFlag)
		ok = false
	} else {
		cfg.clientHosts = addrs
	}

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

func (cfg *AppConfig) JQOpts() JQOpts {
	cfg.mutex.RLock()
	defer cfg.mutex.RUnlock()
	return cfg.jqOpts
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

func (cfg *AppConfig) ClientHosts() []ListenAddr {
	cfg.mutex.RLock()
	defer cfg.mutex.RUnlock()
	return slices.Clone(cfg.clientHosts)
}

func (cfg *AppConfig) ClientPluginsDir() string {
	cfg.mutex.RLock()
	defer cfg.mutex.RUnlock()
	return cfg.clientPluginsDir
}

func (cfg *AppConfig) MgmtHosts() []ListenAddr {
	cfg.mutex.RLock()
	defer cfg.mutex.RUnlock()
	return slices.Clone(cfg.mgmtHosts)
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

func (cfg *AppConfig) CORSOpts() CORSOpts {
	cfg.mutex.RLock()
	defer cfg.mutex.RUnlock()

	opts := cfg.corsOpts
	opts.AllowedOrigins = slices.Clone(opts.AllowedOrigins)
	opts.AllowedMethods = slices.Clone(opts.AllowedMethods)
	opts.AllowedHeaders = slices.Clone(opts.AllowedHeaders)
	return opts
}

func (cfg *AppConfig) OAuthOpts() OAuthOpts {
	cfg.mutex.RLock()
	defer cfg.mutex.RUnlock()
	return cfg.oauthOpts
}

func (cfg *AppConfig) InitStorage() (persistence.Storage, error) {
	name, options := cfg.Storage(), cfg.StorageOptions()
	log.Debug().Str("name", name).Str("options", options).Msgf("Setting up persistent storage %q", name)

	// note: storage is shared between north- and southbound API
	storage := persistence.GetStorage(name)
	if storage == nil {
		return nil, fault.Newf("unknown storage %s", name)
	}
	log.Debug().Str("name", name).Msgf("Initializing storage %q", name)
	if err := storage.Initialize(options); err != nil {
		return nil, fault.Wrap(err)
	}
	log.Info().Str("name", name).Msgf("Initialized storage %q", name)
	return storage, nil
}

func hostValues(k *koanf.Koanf, key string) []string {
	if strs := k.Strings(key); len(strs) > 0 {
		return strs
	}
	if s := k.String(key); s != "" {
		return splitListenURLs(s)
	}
	return nil
}

func splitListenURLs(v string) []string {
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func parseListenURLs(raw []string) ([]ListenAddr, error) {
	addrs := make([]ListenAddr, 0, len(raw))
	for _, s := range raw {
		addr, err := parseListenURL(s)
		if err != nil {
			return nil, fault.Wrap(err)
		}
		addrs = append(addrs, addr)
	}
	return addrs, nil
}

func parseListenURL(raw string) (ListenAddr, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return ListenAddr{}, fault.Newf("invalid listen URL %q", raw)
	}
	switch u.Scheme {
	case "http", "https":
		if u.Hostname() == "" {
			return ListenAddr{}, fault.Newf("host missing from %q", raw)
		}
		port := u.Port()
		if port == "" {
			return ListenAddr{}, fault.Newf("port missing from %q", raw)
		}
		return ListenAddr{
			Network: "tcp",
			Addr:    net.JoinHostPort(u.Hostname(), port),
			TLS:     u.Scheme == "https",
		}, nil
	case "unix":
		if u.Host != "" || u.Path == "" {
			return ListenAddr{}, fault.New("unix host must have form unix:///path/to/socket")
		}
		return ListenAddr{Network: "unix", Addr: u.Path}, nil
	default:
		return ListenAddr{}, fault.Newf("unsupported host scheme %q (want http, https, or unix)", u.Scheme)
	}
}
