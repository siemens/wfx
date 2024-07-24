package server

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net/http"
	"os"
	"path"
	"testing"
	"time"

	"github.com/siemens/wfx/cmd/wfx/cmd/config"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHTTPServer_CustomSettings(t *testing.T) {
	f := pflag.NewFlagSet("", pflag.ContinueOnError)
	f.Int(config.MaxHeaderSizeFlag, 1000000, "controls the maximum number of bytes the server will read parsing the request header's keys and values, including the request line. It does not limit the size of the request body")
	f.Duration(config.ReadTimeoutFlag, 30*time.Second, "maximum duration before timing out read of the request")
	f.Duration(config.WriteTimoutFlag, 10*time.Minute, "maximum duration before timing out write of the response")
	f.Bool(config.KeepAliveFlag, true, "sets the TCP keep-alive timeouts on accepted connections. It prunes dead TCP connections ( e.g. closing laptop mid-download)")
	f.Duration(config.CleanupTimeoutFlag, 10*time.Second, "grace period for which to wait before killing idle connections")
	_ = f.Parse([]string{"--max-header-size=1024", "--read-timeout=10s", "--write-timeout=5s", "--keep-alive", "--cleanup-timeout=5m"})

	cfg, err := config.NewAppConfig(f)
	require.NoError(t, err)
	defer cfg.Stop()

	handler := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})
	server, err := NewHTTPServer(cfg, handler)
	require.NoError(t, err)

	assert.Equal(t, 1024, server.MaxHeaderBytes)
	assert.Equal(t, 10*time.Second, server.ReadTimeout)
	assert.Equal(t, 5*time.Second, server.WriteTimeout)
	assert.Equal(t, 5*time.Minute, server.IdleTimeout)
}

func TestNewHTTPServer_DefaultSettings(t *testing.T) {
	_, err := NewHTTPServer(new(config.AppConfig), nil)
	require.NoError(t, err)
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

	f := pflag.NewFlagSet("", pflag.ContinueOnError)
	f.String(config.TLSCaFlag, "", "the certificate authority certificate file to be used with mutual TLS auth")
	_ = f.Parse([]string{"--tls-ca", caCrtFile})
	cfg, err := config.NewAppConfig(f)
	require.NoError(t, err)
	defer cfg.Stop()

	srv, err := NewHTTPServer(cfg, nil)
	require.NoError(t, err)
	assert.NotNil(t, srv.TLSConfig.ClientCAs)
}
