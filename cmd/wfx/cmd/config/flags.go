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
	"path"
	"slices"
	"strings"
	"time"

	"github.com/OpenPeeDeeP/xdg"
	"github.com/rs/zerolog"
	"github.com/siemens/wfx/persistence"
	"github.com/spf13/pflag"

	// import storages (must be here because we include them into the --help output)
	_ "github.com/siemens/wfx/internal/persistence/entgo"
)

// CLI flags
const (
	VersionFlag          = "version"
	ConfigFlag           = "config"
	LogFormatFlag        = "log-format"
	LogLevelFlag         = "log-level"
	StorageFlag          = "storage"
	StorageOptFlag       = "storage-opt"
	SimpleFileServerFlag = "simple-fileserver"

	ClientHostFlag       = "client-host"
	ClientPortFlag       = "client-port"
	ClientTLSHostFlag    = "client-tls-host"
	ClientTLSPortFlag    = "client-tls-port"
	ClientUnixSocketFlag = "client-unix-socket"
	ClientPluginsDirFlag = "client-plugins-dir"

	MgmtHostFlag       = "mgmt-host"
	MgmtPortFlag       = "mgmt-port"
	MgmtTLSHostFlag    = "mgmt-tls-host"
	MgmtTLSPortFlag    = "mgmt-tls-port"
	MgmtUnixSocketFlag = "mgmt-unix-socket"
	MgmtPluginsDirFlag = "mgmt-plugins-dir"

	SchemeFlag          = "scheme"
	KeepAliveFlag       = "keep-alive"
	MaxHeaderSizeFlag   = "max-header-size"
	CleanupTimeoutFlag  = "cleanup-timeout"
	GracefulTimeoutFlag = "graceful-timeout"
	ReadTimeoutFlag     = "read-timeout"
	WriteTimoutFlag     = "write-timeout"

	TLSCaFlag          = "tls-ca"
	TLSCertificateFlag = "tls-certificate"
	TLSKeyFlag         = "tls-key"
)

const (
	preferedStorage   = "sqlite"
	sqliteDefaultOpts = "file:wfx.db?_fk=1&_journal=WAL"
)

func NewFlagset() *pflag.FlagSet {
	f := pflag.NewFlagSet("wfx", pflag.ExitOnError)

	f.BoolP(VersionFlag, "v", false, "version for wfx")
	f.StringSlice(ConfigFlag, DefaultConfigFiles(), "path to one or more .yaml config files")
	f.StringSlice(SchemeFlag, []string{"http"}, "the listeners to enable, this can be repeated and defaults to the schemes in the swagger spec")
	f.Duration(CleanupTimeoutFlag, 10*time.Second, "grace period for which to wait before killing idle connections")
	f.Duration(GracefulTimeoutFlag, 15*time.Second, "grace period for which to wait before shutting down the server")
	f.Int(MaxHeaderSizeFlag, 1000000, "controls the maximum number of bytes the server will read parsing the request header's keys and values, including the request line. It does not limit the size of the request body")
	f.Bool(KeepAliveFlag, true, "sets the TCP keep-alive timeouts on accepted connections. It prunes dead TCP connections ( e.g. closing laptop mid-download)")
	f.Duration(ReadTimeoutFlag, 30*time.Second, "maximum duration before timing out read of the request")
	f.Duration(WriteTimoutFlag, 10*time.Minute, "maximum duration before timing out write of the response")
	f.String(SimpleFileServerFlag, "", "root directory for built-in fileserver (available under /download)")

	f.String(TLSCertificateFlag, "", "the certificate file to use for secure connections")
	f.String(TLSKeyFlag, "", "the private key file to use for secure connections (without passphrase)")
	f.String(TLSCaFlag, "", "the certificate authority certificate file to be used with mutual TLS auth")

	f.String(ClientHostFlag, "0.0.0.0", "the IP to listen on")
	f.Int(ClientPortFlag, 8080, "the port to listen on for insecure connections")
	f.String(ClientTLSHostFlag, "0.0.0.0", "the IP to listen on")
	f.Int(ClientTLSPortFlag, 8443, "the port to listen on for secure connections, defaults to a random value")
	f.String(ClientUnixSocketFlag, "/tmp/wfx-client.sock", "the unix domain socket to use")
	f.String(ClientPluginsDirFlag, "", "directory containing client plugins")

	f.String(MgmtHostFlag, "127.0.0.1", "management host")
	f.Int(MgmtPortFlag, 8081, "management port")
	f.String(MgmtTLSHostFlag, "127.0.0.1", "management TLS host")
	f.Int(MgmtTLSPortFlag, 8444, "TLS management port")
	f.String(MgmtUnixSocketFlag, "/tmp/wfx-mgmt.sock", "the unix domain socket to use")
	f.String(MgmtPluginsDirFlag, "", "directory containing management plugins")

	{

		supportedStorages := persistence.Storages()
		defaultStorage := supportedStorages[0]
		if slices.Index(supportedStorages, preferedStorage) != -1 {
			defaultStorage = preferedStorage
		}
		f.String(StorageFlag, defaultStorage, fmt.Sprintf("persistence storage. one of: [%s]", strings.Join(supportedStorages, ", ")))

		var storageOpts string
		if defaultStorage == preferedStorage {
			storageOpts = sqliteDefaultOpts
		}
		f.String(StorageOptFlag, storageOpts, "storage options")
	}

	allLevels := []string{zerolog.TraceLevel.String(), zerolog.DebugLevel.String(), zerolog.InfoLevel.String(), zerolog.WarnLevel.String(), zerolog.ErrorLevel.String(), zerolog.FatalLevel.String(), zerolog.PanicLevel.String()}
	f.String(LogLevelFlag, "info", fmt.Sprintf("set log level. one of: [%s]", strings.Join(allLevels, ", ")))
	f.String(LogFormatFlag, "auto", "log format; possible values: json, pretty, auto")
	return f
}

func DefaultConfigFiles() []string {
	configFiles := []string{
		// current directory
		"wfx.yml",
		// user home
		path.Join(xdg.ConfigHome(), "wfx", "config.yml"),
		path.Join("/etc/wfx/wfx.yml"),
	}
	return configFiles
}
