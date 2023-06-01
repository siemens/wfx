package root

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"fmt"
	"strings"
	"time"

	mcobra "github.com/muesli/mango-cobra"
	"github.com/muesli/roff"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/siemens/wfx/cmd/wfx/metadata/config"
	"github.com/siemens/wfx/middleware/fileserver"
	"github.com/siemens/wfx/persistence"
	"github.com/spf13/cobra"

	// import storages (must be here because we include them into the --help output
	_ "github.com/siemens/wfx/internal/persistence/entgo"
)

const (
	configFlag     = "config"
	storageFlag    = "storage"
	storageOptFlag = "storage-opt"
	logFormatFlag  = "log-format"
	logLevelFlag   = "log-level"

	clientHostFlag    = "client-host"
	clientPortFlag    = "client-port"
	clientTLSHostFlag = "client-tls-host"
	clientTLSPortFlag = "client-tls-port"
	clientUnixSocket  = "client-unix-socket"

	mgmtHostFlag       = "mgmt-host"
	mgmtPortFlag       = "mgmt-port"
	mgmtTLSHostFlag    = "mgmt-tls-host"
	mgmtTLSPortFlag    = "mgmt-tls-port"
	mgmtUnixSocketFlag = "mgmt-unix-socket"

	schemeFlag        = "scheme"
	keepAliveFlag     = "keep-alive"
	maxHeaderSizeFlag = "max-header-size"

	cleanupTimeoutFlag  = "cleanup-timeout"
	gracefulTimeoutFlag = "graceful-timeout"
	readTimeoutFlag     = "read-timeout"
	writeTimoutFlag     = "write-timeout"

	tlsCaFlag          = "tls-ca"
	tlsCertificateFlag = "tls-certificate"
	tlsKeyFlag         = "tls-key"

	preferedStorage    = "sqlite"
	defaultStorageOpts = "file:wfx.db?_fk=1&_journal=WAL"
)

func init() {
	Command.AddCommand(&cobra.Command{
		Use:   "man",
		Short: "Generate man page and exit",
		Run: func(cmd *cobra.Command, args []string) {
			manPage, err := mcobra.NewManPage(1, Command)
			if err != nil {
				log.Fatal().Err(err).Msg("Failed to generate man page")
			}
			manPage = manPage.WithSection("Copyright", "(C) 2023 Siemens AG.\n"+
				"Licensed under the Apache License, Version 2.0")
			fmt.Println(manPage.Build(roff.NewDocument()))
		},
	})

	f := Command.PersistentFlags()

	f.StringSlice(schemeFlag, []string{"http"}, "the listeners to enable, this can be repeated and defaults to the schemes in the swagger spec")
	f.Duration(cleanupTimeoutFlag, 10*time.Second, "grace period for which to wait before killing idle connections")
	f.Duration(gracefulTimeoutFlag, 15*time.Second, "grace period for which to wait before shutting down the server")
	f.Int(maxHeaderSizeFlag, 1000000, "controls the maximum number of bytes the server will read parsing the request header's keys and values, including the request line. It does not limit the size of the request body")
	f.Duration(keepAliveFlag, 3*time.Minute, "sets the TCP keep-alive timeouts on accepted connections. It prunes dead TCP connections ( e.g. closing laptop mid-download)")
	f.Duration(readTimeoutFlag, 30*time.Second, "maximum duration before timing out read of the request")
	f.Duration(writeTimoutFlag, 10*time.Minute, "maximum duration before timing out write of the response")

	f.String(tlsCertificateFlag, "", "the certificate file to use for secure connections")
	f.String(tlsKeyFlag, "", "the private key file to use for secure connections (without passphrase)")
	f.String(tlsCaFlag, "", "the certificate authority certificate file to be used with mutual TLS auth")

	f.String(clientHostFlag, "0.0.0.0", "the IP to listen on")
	f.Int(clientPortFlag, 8080, "the port to listen on for insecure connections")
	f.String(clientTLSHostFlag, "0.0.0.0", "the IP to listen on")
	f.Int(clientTLSPortFlag, 8443, "the port to listen on for secure connections, defaults to a random value")
	f.String(clientUnixSocket, "/tmp/wfx-client.sock", "the unix domain socket to use")

	f.String(mgmtHostFlag, "0.0.0.0", "management host")
	f.Int(mgmtPortFlag, 8081, "management port")
	f.String(mgmtTLSHostFlag, "0.0.0.0", "management TLS host")
	f.Int(mgmtTLSPortFlag, 8444, "TLS management port")
	f.String(mgmtUnixSocketFlag, "/tmp/wfx-mgmt.sock", "the unix domain socket to use")

	_ = Command.MarkPersistentFlagDirname(fileserver.SimpleFileServerFlag)
	f.String(fileserver.SimpleFileServerFlag, "", "root directory for built-in fileserver (available under /download)")

	f.StringSlice(configFlag, config.DefaultConfigFiles(), "path to one or more .yaml config files")
	_ = Command.MarkPersistentFlagFilename(configFlag, "yml", "yaml")

	{
		var defaultStorage string
		supportedStorages := persistence.Storages()
		for _, storage := range supportedStorages {
			if storage == preferedStorage {
				defaultStorage = preferedStorage
				break
			}
			defaultStorage = storage
		}
		f.String(storageFlag, defaultStorage, fmt.Sprintf("persistence storage. one of: [%s]", strings.Join(supportedStorages, ", ")))

		if defaultStorage == preferedStorage {
			f.String(storageOptFlag, defaultStorageOpts, "custom storage options")
		} else {
			f.String(storageOptFlag, "", "custom storage options")
		}
	}

	f.String(logLevelFlag, "info", fmt.Sprintf("set log level. one of: [%s,%s,%s,%s,%s,%s,%s]",
		zerolog.TraceLevel.String(),
		zerolog.DebugLevel.String(),
		zerolog.InfoLevel.String(),
		zerolog.WarnLevel.String(),
		zerolog.ErrorLevel.String(),
		zerolog.FatalLevel.String(),
		zerolog.PanicLevel.String()))
	f.String(logFormatFlag, "auto", "log format; possible values: json, pretty, auto")
}
