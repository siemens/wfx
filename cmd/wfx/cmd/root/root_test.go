//go:build linux

/*
 * SPDX-FileCopyrightText: 2024 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

package root

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/siemens/wfx/cmd/wfx/cmd/config"
	"github.com/siemens/wfx/cmd/wfxctl/cmd"
	"github.com/siemens/wfx/cmd/wfxctl/errutil"
	"github.com/siemens/wfx/generated/api"
	"github.com/siemens/wfx/workflow/dau"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
)

func TestUDS(t *testing.T) {
	clientSocket, mgmtSocket := generateTempFilenames(t)

	cmd := NewCommand()
	cmd.SetArgs([]string{
		"--log-level=debug",
		"--storage-opt=file:wfx?mode=memory&cache=shared&_fk=1",
		"--scheme=unix",
		"--client-unix-socket", clientSocket,
		"--mgmt-unix-socket", mgmtSocket,
	})
	ctx, cancel := context.WithCancel(context.Background())
	cmd.SetContext(ctx)

	var g errgroup.Group
	g.Go(cmd.Execute)

	for i := 0; i < 30; i++ {
		conn, err := net.Dial("unix", clientSocket)
		if err != nil {
			time.Sleep(time.Millisecond * 10)
			continue
		}
		defer conn.Close()

		req, err := http.NewRequest(http.MethodGet, "/version", nil)
		require.NoError(t, err)
		client := &http.Client{
			Transport: &http.Transport{
				DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
					return conn, nil
				},
			},
		}
		resp, err := client.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			break
		}
	}

	// tell Command to stop
	cancel()

	err := g.Wait()
	require.NoError(t, err)
}

func TestTLSOnly(t *testing.T) {
	privkey, pubkey := createKeypair(t)

	cmd := NewCommand()
	cmd.SetArgs([]string{
		"--log-level=debug",
		"--storage-opt=file:wfx?mode=memory&cache=shared&_fk=1",
		"--scheme=https",
		"--tls-certificate", pubkey,
		"--tls-key", privkey,
	})
	ctx, cancel := context.WithCancel(context.Background())
	cmd.SetContext(ctx)

	var g errgroup.Group
	g.Go(func() error {
		return cmd.Execute()
	})

	httpClient := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
			TLSHandshakeTimeout: time.Second * 10,
		},
		Timeout: time.Second * 10,
	}

	f := config.NewFlagset()

	var upCount int
	tlsEndpoints := []string{
		fmt.Sprintf("https://%s:%s", f.Lookup(config.ClientTLSHostFlag).DefValue, f.Lookup(config.ClientTLSPortFlag).DefValue),
		fmt.Sprintf("https://%s:%s", f.Lookup(config.MgmtTLSHostFlag).DefValue, f.Lookup(config.MgmtTLSPortFlag).DefValue),
	}
	for i := 0; i < 30; i++ {
		upCount = 0
		for _, endpoint := range tlsEndpoints {
			client := errutil.Must(api.NewClientWithResponses(endpoint, api.WithHTTPClient(&httpClient)))
			resp, _ := client.GetHealthWithResponse(ctx)
			if resp != nil && resp.JSON200 != nil && resp.JSON200.Status == api.Up {
				t.Log("Endpoint is up:", endpoint)
				upCount++
			} else {
				t.Log("Endpoint is down:", endpoint)
			}
		}
		t.Log("Upcount:", upCount)
		if upCount == len(tlsEndpoints) {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	insecureEndpoints := []string{
		fmt.Sprintf("http://%s:%s", f.Lookup(config.ClientHostFlag).DefValue, f.Lookup(config.ClientPortFlag).DefValue),
		fmt.Sprintf("http://%s:%s", f.Lookup(config.MgmtHostFlag).DefValue, f.Lookup(config.MgmtPortFlag).DefValue),
	}
	for _, endpoint := range insecureEndpoints {
		resp, err := httpClient.Get(fmt.Sprintf("%s/health", endpoint))
		assert.Error(t, err)
		assert.Nil(t, resp)
	}

	// tell Command to stop
	cancel()

	err := g.Wait()
	require.NoError(t, err)
	assert.Equal(t, len(tlsEndpoints), upCount)
}

func TestTLSMixedMode(t *testing.T) {
	privkey, pubkey := createKeypair(t)

	cmd := NewCommand()
	cmd.SetArgs([]string{
		"--log-level=debug",
		"--storage-opt=file:wfx?mode=memory&cache=shared&_fk=1",
		"--scheme=http",
		"--scheme=https",
		"--tls-certificate", pubkey,
		"--tls-key", privkey,
	})
	ctx, cancel := context.WithCancel(context.Background())
	cmd.SetContext(ctx)

	var g errgroup.Group
	g.Go(func() error {
		return cmd.Execute()
	})

	httpClient := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
			TLSHandshakeTimeout: time.Second * 10,
		},
		Timeout: time.Second * 10,
	}

	f := config.NewFlagset()

	var upCount int
	endpoints := []string{
		fmt.Sprintf("http://%s:%s", f.Lookup(config.ClientHostFlag).DefValue, f.Lookup(config.ClientPortFlag).DefValue),
		fmt.Sprintf("https://%s:%s", f.Lookup(config.ClientTLSHostFlag).DefValue, f.Lookup(config.ClientTLSPortFlag).DefValue),
		fmt.Sprintf("http://%s:%s", f.Lookup(config.MgmtHostFlag).DefValue, f.Lookup(config.MgmtPortFlag).DefValue),
		fmt.Sprintf("https://%s:%s", f.Lookup(config.MgmtTLSHostFlag).DefValue, f.Lookup(config.MgmtTLSPortFlag).DefValue),
	}
	t.Logf("Endpoints: %v", endpoints)
	for i := 0; i < 30; i++ {
		upCount = 0
		for _, endpoint := range endpoints {
			client := errutil.Must(api.NewClientWithResponses(endpoint, api.WithHTTPClient(&httpClient)))
			resp, _ := client.GetHealthWithResponse(ctx)
			if resp != nil && resp.JSON200 != nil && resp.JSON200.Status == api.Up {
				t.Log("Endpoint is up:", endpoint)
				upCount++
			} else {
				t.Log("Endpoint is down:", endpoint)
			}
		}
		t.Log("Upcount:", upCount)
		if upCount == len(endpoints) {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	// tell Command to stop
	cancel()

	err := g.Wait()
	require.NoError(t, err)
	assert.Equal(t, len(endpoints), upCount)
}

func TestAPI(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	var g errgroup.Group
	g.Go(func() error {
		cmd := NewCommand()
		cmd.SetArgs([]string{
			"--log-level=debug",
			"--storage-opt=file:wfx?mode=memory&cache=shared&_fk=1",
		})
		cmd.SetContext(ctx)
		return cmd.Execute()
	})

	t.Run("fetch openapiv3 spec", func(t *testing.T) {
		httpClient := new(http.Client)
		var spec openapi3.T
		for i := 0; i < 100; i++ {
			resp, err := httpClient.Get("http://localhost:8080/api/wfx/v1/openapi.json")
			if err != nil {
				time.Sleep(time.Millisecond * 10)
				continue
			}
			defer resp.Body.Close()
			// check valid json
			err = json.NewDecoder(resp.Body).Decode(&spec)
			require.NoError(t, err)
			break
		}
		assert.Equal(t, "3.0.0", spec.OpenAPI)
	})

	t.Run("/health endpoint", func(t *testing.T) {
		httpClient := new(http.Client)
		var result api.CheckerResult
		for i := 0; i < 100; i++ {
			resp, err := httpClient.Get("http://localhost:8080/health")
			if err != nil {
				time.Sleep(time.Millisecond * 10)
				continue
			}
			defer resp.Body.Close()
			// check valid json
			err = json.NewDecoder(resp.Body).Decode(&result)
			require.NoError(t, err)
			break
		}
		assert.Equal(t, api.Up, result.Status)
	})

	t.Run("/version endpoint", func(t *testing.T) {
		httpClient := new(http.Client)
		var result api.GetVersion200JSONResponse
		for i := 0; i < 100; i++ {
			resp, err := httpClient.Get("http://localhost:8080/version")
			if err != nil {
				time.Sleep(time.Millisecond * 10)
				continue
			}
			defer resp.Body.Close()
			// check valid json
			err = json.NewDecoder(resp.Body).Decode(&result)
			require.NoError(t, err)
			break
		}
		assert.Equal(t, "v1", result.ApiVersion)
	})

	t.Run("Response Filters", func(t *testing.T) {
		tmpFile, _ := os.CreateTemp("", "dau.yml.*")
		wf := dau.DirectWorkflow()
		_ = json.NewEncoder(tmpFile).Encode(wf)
		_ = tmpFile.Close()
		t.Cleanup(func() { _ = os.Remove(tmpFile.Name()) })

		t.Setenv("WFX_CLIENT_HOST", "localhost")
		t.Setenv("WFX_CLIENT_PORT", "8080")
		t.Setenv("WFX_MGMT_HOST", "localhost")
		t.Setenv("WFX_MGMT_PORT", "8081")

		wfxctl := cmd.NewCommand()
		wfxctl.SetArgs([]string{"workflow", "create", tmpFile.Name()})
		err := wfxctl.Execute()
		require.NoError(t, err)

		buf := new(bytes.Buffer)
		wfxctl = cmd.NewCommand()
		wfxctl.SetArgs([]string{
			"job", "create", "--workflow", wf.Name,
			"--client-id=Dana", "--filter=.id", "--raw",
		})
		wfxctl.SetOut(buf)
		err = wfxctl.Execute()
		require.NoError(t, err)
		jobID := strings.TrimRight(buf.String(), "\n")
		assert.NotEmpty(t, jobID)
		t.Log("jobID", jobID)

		httpClient := new(http.Client)
		var state string
		for i := 0; i < 100; i++ {
			url := fmt.Sprintf("http://localhost:8080/api/wfx/v1/jobs/%s/status", jobID)
			req, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)
			req.Header.Set("X-Response-Filter", ".state")
			resp, err := httpClient.Do(req)
			if err != nil {
				time.Sleep(time.Millisecond * 10)
				continue
			}
			defer resp.Body.Close()
			b, _ := io.ReadAll(resp.Body)
			state = strings.TrimSpace(string(b))
			break
		}
		assert.Equal(t, `"INSTALL"`, state)
	})

	t.Run("Fileserver Disabled", func(t *testing.T) {
		httpClient := new(http.Client)
		var statusCode int
		var body []byte
		for i := 0; i < 100; i++ {
			resp, err := httpClient.Get("http://localhost:8080/download/")
			if err != nil {
				time.Sleep(time.Millisecond * 10)
				continue
			}
			defer resp.Body.Close()
			body, _ = io.ReadAll(resp.Body)
			statusCode = resp.StatusCode

		}
		assert.Equal(t, http.StatusNotFound, statusCode)
		assert.Equal(t, `{"code":404,"message":"path /download is not served as simple file server is not enabled"}`, string(body))
	})

	cancel()
	err := g.Wait()
	require.NoError(t, err)
}

func TestSimpleFileServer(t *testing.T) {
	dir, err := os.MkdirTemp("", "wfx-fileserver*")
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(dir) })

	testFile := path.Join(dir, "hello.txt")
	err = os.WriteFile(testFile, []byte("Hello World!"), 0o644)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())

	cmd := NewCommand()
	cmd.SetContext(ctx)
	cmd.SetArgs([]string{
		"--log-level=debug",
		"--storage-opt=file:wfx?mode=memory&cache=shared&_fk=1",
		"--" + config.SimpleFileServerFlag, dir,
	})

	var g errgroup.Group
	g.Go(func() error {
		return cmd.Execute()
	})

	httpClient := new(http.Client)
	var statusCode int
	var body []byte
	for i := 0; i < 100; i++ {
		resp, err := httpClient.Get("http://localhost:8080/download/hello.txt")
		if err != nil {
			time.Sleep(time.Millisecond * 10)
			continue
		}
		defer resp.Body.Close()
		body, _ = io.ReadAll(resp.Body)
		statusCode = resp.StatusCode
	}

	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, "Hello World!", string(body))

	cancel()

	err = g.Wait()
	require.NoError(t, err)
}

func generateTempFilenames(t *testing.T) (string, string) {
	fnames := make([]string, 0, 2)
	for i := 0; i < 2; i++ {
		f, _ := os.CreateTemp(os.TempDir(), fmt.Sprintf("%s.*.sock", t.Name()))
		f.Close()
		fname := f.Name()
		_ = os.Remove(fname)
		fnames = append(fnames, fname)
	}
	return fnames[0], fnames[1]
}

func createKeypair(t *testing.T) (string, string) {
	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	template := x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "My CA"},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(0, 0, 1), // Valid for 1 day
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	derBytes, _ := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)

	dir, _ := os.MkdirTemp("", "wfx-testkeys.*")
	t.Cleanup(func() { os.RemoveAll(dir) })

	privkey, pubkey := path.Join(dir, "test.key"), path.Join(dir, "test.crt")
	certOut, _ := os.Create(pubkey)
	defer certOut.Close()

	_ = pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})

	keyOut, _ := os.OpenFile(privkey, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	defer keyOut.Close()

	_ = pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})

	return privkey, pubkey
}
