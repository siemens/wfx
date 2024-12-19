package server

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"

	"github.com/Southclaws/fault"
	"github.com/rs/zerolog/log"
	"github.com/siemens/wfx/cmd/wfx/cmd/config"
)

func NewHTTPServer(cfg *config.AppConfig, handler http.Handler) (*http.Server, error) {
	server := new(http.Server)
	server.MaxHeaderBytes = int(cfg.MaxHeaderSize())
	server.ReadTimeout = cfg.ReadTimeout()
	server.WriteTimeout = cfg.WriteTimeout()
	server.SetKeepAlivesEnabled(cfg.KeepAlive())

	if cfg.CleanupTimeout() > 0 {
		server.IdleTimeout = cfg.CleanupTimeout()
	}
	if err := configureTLS(server, cfg.TLSCACertificate()); err != nil {
		return nil, fault.Wrap(err)
	}
	server.Handler = handler
	return server, nil
}

func configureTLS(server *http.Server, caCert string) error {
	// Inspired by https://blog.bracebin.com/achieving-perfect-ssl-labs-score-with-go
	server.TLSConfig = &tls.Config{
		// Causes servers to use Go's default ciphersuite preferences,
		// which are tuned to avoid attacks. Does nothing on clients.
		PreferServerCipherSuites: true,
		// Only use curves which have assembly implementations
		// https://github.com/golang/go/tree/master/src/crypto/elliptic
		CurvePreferences: []tls.CurveID{tls.CurveP256},
		// Use modern tls mode https://wiki.mozilla.org/Security/Server_Side_TLS#Modern_compatibility
		NextProtos: []string{"h2", "http/1.1"},
		// https://www.owasp.org/index.php/Transport_Layer_Protection_Cheat_Sheet#Rule_-_Only_Support_Strong_Protocols
		MinVersion: tls.VersionTLS12,
		// These ciphersuites support Forward Secrecy: https://en.wikipedia.org/wiki/Forward_secrecy
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
		},
	}

	if caCert != "" {
		log.Debug().
			Str("caCertificates", caCert).
			Msg("Loading CA certs")
		// include specified CA certificate
		caCert, caCertErr := os.ReadFile(caCert)
		if caCertErr != nil {
			return fault.Wrap(caCertErr)
		}
		caCertPool, err := x509.SystemCertPool()
		if err != nil {
			log.Warn().Err(err).Msg("Failed to load system cert pool, using empty pool")
			caCertPool = x509.NewCertPool()
		}
		caCertPool.AppendCertsFromPEM(caCert)
		ok := caCertPool.AppendCertsFromPEM(caCert)
		if !ok {
			return fmt.Errorf("cannot parse CA certificate")
		}
		server.TLSConfig.ClientCAs = caCertPool
		server.TLSConfig.ClientAuth = tls.RequireAndVerifyClientCert
	}
	return nil
}
