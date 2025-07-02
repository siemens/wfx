package flags

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/rs/zerolog"
	"github.com/siemens/wfx/generated/api"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateHTTPClientTLS(t *testing.T) {
	cert := []byte(`
-----BEGIN CERTIFICATE-----
MIIDrzCCApegAwIBAgIQCDvgVpBCRrGhdWrJWZHHSjANBgkqhkiG9w0BAQUFADBh
MQswCQYDVQQGEwJVUzEVMBMGA1UEChMMRGlnaUNlcnQgSW5jMRkwFwYDVQQLExB3
d3cuZGlnaWNlcnQuY29tMSAwHgYDVQQDExdEaWdpQ2VydCBHbG9iYWwgUm9vdCBD
QTAeFw0wNjExMTAwMDAwMDBaFw0zMTExMTAwMDAwMDBaMGExCzAJBgNVBAYTAlVT
MRUwEwYDVQQKEwxEaWdpQ2VydCBJbmMxGTAXBgNVBAsTEHd3dy5kaWdpY2VydC5j
b20xIDAeBgNVBAMTF0RpZ2lDZXJ0IEdsb2JhbCBSb290IENBMIIBIjANBgkqhkiG
9w0BAQEFAAOCAQ8AMIIBCgKCAQEA4jvhEXLeqKTTo1eqUKKPC3eQyaKl7hLOllsB
CSDMAZOnTjC3U/dDxGkAV53ijSLdhwZAAIEJzs4bg7/fzTtxRuLWZscFs3YnFo97
nh6Vfe63SKMI2tavegw5BmV/Sl0fvBf4q77uKNd0f3p4mVmFaG5cIzJLv07A6Fpt
43C/dxC//AH2hdmoRBBYMql1GNXRor5H4idq9Joz+EkIYIvUX7Q6hL+hqkpMfT7P
T19sdl6gSzeRntwi5m3OFBqOasv+zbMUZBfHWymeMr/y7vrTC0LUq7dBMtoM1O/4
gdW7jVg/tRvoSSiicNoxBN33shbyTApOB6jtSj1etX+jkMOvJwIDAQABo2MwYTAO
BgNVHQ8BAf8EBAMCAYYwDwYDVR0TAQH/BAUwAwEB/zAdBgNVHQ4EFgQUA95QNVbR
TLtm8KPiGxvDl7I90VUwHwYDVR0jBBgwFoAUA95QNVbRTLtm8KPiGxvDl7I90VUw
DQYJKoZIhvcNAQEFBQADggEBAMucN6pIExIK+t1EnE9SsPTfrgT1eXkIoyQY/Esr
hMAtudXH/vTBH1jLuG2cenTnmCmrEbXjcKChzUyImZOMkXDiqw8cvpOp/2PV5Adg
06O/nVsJ8dWO41P0jmP6P6fbtGbfYmbW0W5BjfIttep3Sp+dWOIrWcBAI+0tKIJF
PnlUkiaY4IBIqDfv8NZ5YBberOgOzW6sRBc4L0na4UU+Krk2U886UAb3LujEV0ls
YSEY1QSteDwsOoBrp+uvFRTp2InBuThs4pFsiv9kuXclVzDAGySj4dzp30d8tbQk
CAUw7C29C79Fv1C5qfPrmAESrciIxpg0X40KPMbp1ZWVbd4=
-----END CERTIFICATE-----
	`)
	tmpFile, _ := os.CreateTemp("", "wfx-cert-")
	_, _ = tmpFile.Write(cert)
	t.Cleanup(func() { _ = os.Remove(tmpFile.Name()) })

	b := NewBaseCmd(pflag.NewFlagSet("wfx", pflag.ExitOnError))
	b.EnableTLS = true
	b.TLSCa = tmpFile.Name()

	client, _ := b.CreateHTTPClient()
	transport := client.Transport.(*http.Transport)
	assert.NotEmpty(t, transport.TLSClientConfig.RootCAs)
}

func TestCreateHTTPClient_AmbiguousSockets(t *testing.T) {
	b := NewBaseCmd(pflag.NewFlagSet("wfx", pflag.ExitOnError))
	b.Socket = "foo"
	b.MgmtSocket = "bar"
	_, err := b.CreateHTTPClient()
	require.Error(t, err)
}

func TestCreateHTTPClient_Socket(t *testing.T) {
	b := NewBaseCmd(pflag.NewFlagSet("wfx", pflag.ExitOnError))
	b.Socket = "/tmp/foo.sock"
	_, err := b.CreateHTTPClient()
	require.NoError(t, err)
}

func TestCreateClient_TLS(t *testing.T) {
	b := NewBaseCmd(pflag.NewFlagSet("wfx", pflag.ExitOnError))
	b.EnableTLS = true
	client, err := b.CreateClient()
	assert.NotNil(t, client)
	assert.NoError(t, err)
}

func TestCreateClient_NoTLS(t *testing.T) {
	b := NewBaseCmd(pflag.NewFlagSet("wfx", pflag.ExitOnError))
	b.EnableTLS = false
	client, err := b.CreateClient()
	assert.NotNil(t, client)
	assert.NoError(t, err)
}

func TestCreateMgmtClient(t *testing.T) {
	b := NewBaseCmd(pflag.NewFlagSet("wfx", pflag.ExitOnError))
	b.EnableTLS = true
	client, err := b.CreateMgmtClient()
	assert.NotNil(t, client)
	assert.NoError(t, err)
}

func TestCreateMgmtClient_NoTLS(t *testing.T) {
	b := NewBaseCmd(pflag.NewFlagSet("wfx", pflag.ExitOnError))
	b.EnableTLS = false
	client, err := b.CreateMgmtClient()
	assert.NotNil(t, client)
	assert.NoError(t, err)
}

func TestDumpPlain(t *testing.T) {
	payload := []byte("{\n  \"foo\": \"bar\",\n  \"id\": \"1\"\n}\n")
	var buf bytes.Buffer

	b := NewBaseCmd(pflag.NewFlagSet("wfx", pflag.ExitOnError))
	b.Filter = ""

	err := b.dumpResponse(&buf, payload)
	assert.NoError(t, err)
	assert.JSONEq(t, "{\n  \"foo\": \"bar\",\n  \"id\": \"1\"\n}\n", buf.String())
}

func TestDumpFilter(t *testing.T) {
	payload := []byte("{\n  \"foo\": \"bar\",\n  \"id\": \"1\"\n}\n")
	var buf bytes.Buffer

	b := NewBaseCmd(pflag.NewFlagSet("wfx", pflag.ExitOnError))
	b.Filter = ".id"
	b.RawOutput = false

	err := b.dumpResponse(&buf, payload)
	assert.NoError(t, err)
	assert.JSONEq(t, "\"1\"", buf.String())
}

func TestDumpFilterRaw(t *testing.T) {
	payload := []byte("{\n  \"foo\": \"bar\",\n  \"id\": \"1\"\n}\n")
	var buf bytes.Buffer
	err := dumpFiltered(payload, ".id", true, &buf)
	assert.NoError(t, err)
	assert.JSONEq(t, "1", buf.String())
}

func TestProcessResponse(t *testing.T) {
	recorder := httptest.NewRecorder()
	recorder.WriteHeader(http.StatusOK)
	_, _ = recorder.WriteString(`{"foo": "bar"}`)
	resp := recorder.Result()
	buf := new(bytes.Buffer)
	b := NewBaseCmd(pflag.NewFlagSet("wfx", pflag.ExitOnError))
	err := b.ProcessResponse(resp, buf)
	assert.NoError(t, err)
}

func TestProcessResponse_Empty(t *testing.T) {
	recorder := httptest.NewRecorder()
	recorder.WriteHeader(http.StatusNoContent)
	resp := recorder.Result()
	buf := new(bytes.Buffer)
	b := NewBaseCmd(pflag.NewFlagSet("wfx", pflag.ExitOnError))
	err := b.ProcessResponse(resp, buf)
	assert.NoError(t, err)
}

func TestProcessResponse_Error(t *testing.T) {
	recorder := httptest.NewRecorder()
	recorder.WriteHeader(http.StatusInternalServerError)
	resp := recorder.Result()
	buf := new(bytes.Buffer)
	b := NewBaseCmd(pflag.NewFlagSet("wfx", pflag.ExitOnError))
	err := b.ProcessResponse(resp, buf)
	assert.Error(t, err)
}

func TestProcessResponse_ErrorResponse(t *testing.T) {
	recorder := httptest.NewRecorder()
	recorder.WriteHeader(http.StatusInternalServerError)
	errResp := api.ErrorResponse{
		Errors: &[]api.Error{
			{Code: "foo", Message: "bar", Logref: "baz"},
		},
	}
	_ = json.NewEncoder(recorder).Encode(errResp)
	resp := recorder.Result()
	buf := new(bytes.Buffer)
	b := NewBaseCmd(pflag.NewFlagSet("wfx", pflag.ExitOnError))
	err := b.ProcessResponse(resp, buf)
	assert.Error(t, err)
}

func TestSortParam_Asc(t *testing.T) {
	b := NewBaseCmd(pflag.NewFlagSet("wfx", pflag.ExitOnError))
	b.Sort = "asc"
	val, err := b.SortParam()
	assert.NoError(t, err)
	assert.Equal(t, api.Asc, *val)
}

func TestSortParam_Desc(t *testing.T) {
	b := NewBaseCmd(pflag.NewFlagSet("wfx", pflag.ExitOnError))
	b.Sort = "desc"
	val, err := b.SortParam()
	assert.NoError(t, err)
	assert.Equal(t, api.Desc, *val)
}

func TestSortParam_Empty(t *testing.T) {
	b := NewBaseCmd(pflag.NewFlagSet("wfx", pflag.ExitOnError))
	b.Sort = ""
	val, err := b.SortParam()
	assert.NoError(t, err)
	assert.Equal(t, api.Asc, *val)
}

func TestSortParam_Invalid(t *testing.T) {
	b := NewBaseCmd(pflag.NewFlagSet("wfx", pflag.ExitOnError))
	b.Sort = "foo"
	val, err := b.SortParam()
	assert.Error(t, err)
	assert.Nil(t, val)
}

func TestNewBaseCmd_LogLevel(t *testing.T) {
	f := pflag.NewFlagSet("test", pflag.ContinueOnError)
	f.String(LogLevelFlag, "trace", "log level")
	_ = NewBaseCmd(f)
	assert.Equal(t, zerolog.TraceLevel, zerolog.GlobalLevel())
}

func TestNewBaseCmd_ConfigFile(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "test-log-level-*.yaml")
	require.NoError(t, err)
	t.Log(tmpfile.Name())
	t.Cleanup(func() { _ = os.Remove(tmpfile.Name()) })

	content := []byte("log-level: trace\n")
	_, err = tmpfile.Write(content)
	require.NoError(t, err)
	err = tmpfile.Close()
	require.NoError(t, err)

	f := pflag.NewFlagSet("test", pflag.ContinueOnError)
	f.StringSlice(ConfigFlag, []string{tmpfile.Name()}, "config files")
	_ = NewBaseCmd(f)
	assert.Equal(t, zerolog.TraceLevel, zerolog.GlobalLevel())
}

func TestNewBaseCmd_EnvVariables(t *testing.T) {
	t.Setenv("WFX_LOG_LEVEL", "trace")
	f := pflag.NewFlagSet("test", pflag.ContinueOnError)
	f.String(LogLevelFlag, "debug", "log level")
	NewBaseCmd(f)
	assert.Equal(t, zerolog.TraceLevel, zerolog.GlobalLevel())
}
