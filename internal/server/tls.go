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
)

type TLSSettings struct {
	TLSCertificate    string
	TLSCertificateKey string
	TLSCACertificate  string
}

func ConfigureTLS(server *http.Server, settings *TLSSettings) error {
	var err error

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

	server.TLSConfig.Certificates = make([]tls.Certificate, 1)
	log.Debug().
		Str("certificate", settings.TLSCertificate).
		Str("certificateKey", settings.TLSCertificateKey).
		Msg("Loading X509 keypairs")
	server.TLSConfig.Certificates[0], err = tls.LoadX509KeyPair(settings.TLSCertificate, settings.TLSCertificateKey)
	if err != nil {
		return fault.Wrap(err)
	}

	if settings.TLSCACertificate != "" {
		log.Debug().
			Str("caCertificates", settings.TLSCACertificate).
			Msg("Loading CA certs")
		// include specified CA certificate
		caCert, caCertErr := os.ReadFile(settings.TLSCACertificate)
		if caCertErr != nil {
			return fault.Wrap(caCertErr)
		}
		caCertPool := x509.NewCertPool()
		ok := caCertPool.AppendCertsFromPEM(caCert)
		if !ok {
			return fmt.Errorf("cannot parse CA certificate")
		}
		server.TLSConfig.ClientCAs = caCertPool
		server.TLSConfig.ClientAuth = tls.RequireAndVerifyClientCert
	}
	return nil
}
