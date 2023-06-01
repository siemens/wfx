package server

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

func TestConfigureTLS_DefaultSettings(t *testing.T) {
	srv := NewHTTPServer(&HTTPSettings{}, nil)
	err := ConfigureTLS(srv, &TLSSettings{})
	require.Error(t, err)
}

func TestConfigureTLS(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	template := x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "My CA"},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(0, 0, 1), // Valid for 1 day
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	require.NoError(t, err)

	dir, err := os.MkdirTemp(os.TempDir(), "wfx-tlsca")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	caCrtFile := path.Join(dir, "ca.crt")
	caKeyFile := path.Join(dir, "ca.key")
	{
		certOut, err := os.Create(caCrtFile)
		require.NoError(t, err)
		defer certOut.Close()

		err = pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
		require.NoError(t, err)

		keyOut, err := os.OpenFile(caKeyFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
		require.NoError(t, err)
		defer keyOut.Close()

		err = pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})
		require.NoError(t, err)
	}

	srv := NewHTTPServer(&HTTPSettings{}, nil)
	err = ConfigureTLS(srv, &TLSSettings{
		TLSCACertificate:  caCrtFile,
		TLSCertificate:    caCrtFile,
		TLSCertificateKey: caKeyFile,
	})
	require.NoError(t, err)
}
