package root

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

	"github.com/knadh/koanf/v2"
	"github.com/siemens/wfx/internal/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServerKind(t *testing.T) {
	assert.Equal(t, "http", kindHTTP.String())
	assert.Equal(t, "https", kindHTTPS.String())
	assert.Equal(t, "unix", kindUnix.String())
}

func TestCreateServers_InvalidKind(t *testing.T) {
	_, err := createServers([]string{"foo"}, nil, server.HTTPSettings{})
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "unknown scheme: foo")
}

func TestCreateServers_TLS(t *testing.T) {
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

	k.Write(func(k *koanf.Koanf) {
		_ = k.Set(tlsCertificateFlag, caCrtFile)
		_ = k.Set(tlsKeyFlag, caKeyFile)
		_ = k.Set(tlsCaFlag, caCrtFile)
	})

	handler := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
	})

	servers, err := createServers([]string{kindHTTPS.String()}, handler, server.HTTPSettings{})
	require.Nil(t, err)
	require.Len(t, servers, 1)
	assert.Equal(t, kindHTTPS, servers[0].Kind)
	assert.NotEmpty(t, servers[0].Srv.TLSConfig.Certificates)
}
