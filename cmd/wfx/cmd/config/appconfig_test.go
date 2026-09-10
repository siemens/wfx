package config

/*
 * SPDX-FileCopyrightText: 2024 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"io"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAppConfig(t *testing.T) {
	cfg, err := NewAppConfig(NewFlagset())
	require.NoError(t, err)
	t.Cleanup(cfg.Stop)

	// call all methods which do not accept arguments
	structValue := reflect.ValueOf(cfg)
	for i := 0; i < structValue.NumMethod(); i++ {
		method := structValue.Method(i)
		methodType := method.Type()
		methodName := structValue.Type().Method(i).Name
		if methodType.NumIn() == 0 {
			t.Run(methodName, func(*testing.T) {
				if methodName == "InitStorage" {
					return
				}
				_ = method.Call([]reflect.Value{})
			})
		}
	}
}

func TestNewAppConfig_Invalid(t *testing.T) {
	flags := pflag.NewFlagSet("TestNewAppConfig_Invalid", pflag.ContinueOnError)
	_ = flags.String(LogLevelFlag, "info", "Log level")
	_ = flags.Parse([]string{"--log-level", "foo"})

	cfg, err := NewAppConfig(flags)
	assert.Nil(t, cfg)
	assert.Error(t, err)
}

func TestNewAppConfigNegativeJQLimit(t *testing.T) {
	for _, args := range [][]string{
		{"--" + JQFilterTimeoutFlag, "-1s"},
		{"--" + JQFilterMaxResponseSizeFlag, "-1"},
	} {
		flags := NewFlagset()
		require.NoError(t, flags.Parse(args))

		cfg, err := NewAppConfig(flags)

		assert.Nil(t, cfg)
		assert.Error(t, err)
	}
}

func TestNewAppConfig_ListenURLs(t *testing.T) {
	flags := NewFlagset()
	require.NoError(t, flags.Parse([]string{
		"--" + ClientHostFlag, "http://127.0.0.1:18080",
		"--" + ClientHostFlag, "https://0.0.0.0:18443",
		"--" + MgmtHostFlag, "unix:///tmp/wfx-mgmt.sock",
	}))

	cfg, err := NewAppConfig(flags)
	require.NoError(t, err)
	t.Cleanup(cfg.Stop)

	assert.Equal(t, []ListenAddr{
		{Network: "tcp", Addr: "127.0.0.1:18080"},
		{Network: "tcp", Addr: "0.0.0.0:18443", TLS: true},
	}, cfg.ClientHosts())
	assert.Equal(t, []ListenAddr{{Network: "unix", Addr: "/tmp/wfx-mgmt.sock"}}, cfg.MgmtHosts())
}

func TestNewAppConfig_InvalidListenURL(t *testing.T) {
	flags := NewFlagset()
	require.NoError(t, flags.Parse([]string{"--" + ClientHostFlag, "ftp://localhost:21"}))

	cfg, err := NewAppConfig(flags)
	assert.Nil(t, cfg)
	assert.Error(t, err)
}

func TestNewAppConfig_ListenURLsFromYAML(t *testing.T) {
	cfgFile, err := os.CreateTemp("", "config.yaml")
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.Remove(cfgFile.Name()) })
	_, err = cfgFile.WriteString("client-host:\n  - http://127.0.0.1:18080\n  - https://127.0.0.1:18443\nmgmt-host: unix:///tmp/wfx-mgmt.sock\n")
	require.NoError(t, err)
	require.NoError(t, cfgFile.Close())

	flags := NewFlagset()
	require.NoError(t, flags.Parse([]string{"--" + ConfigFlag, cfgFile.Name()}))
	cfg, err := NewAppConfig(flags)
	require.NoError(t, err)
	t.Cleanup(cfg.Stop)

	assert.Equal(t, []ListenAddr{
		{Network: "tcp", Addr: "127.0.0.1:18080"},
		{Network: "tcp", Addr: "127.0.0.1:18443", TLS: true},
	}, cfg.ClientHosts())
	assert.Equal(t, []ListenAddr{{Network: "unix", Addr: "/tmp/wfx-mgmt.sock"}}, cfg.MgmtHosts())
}

func TestNewAppConfig_ListenURLsFromEnv(t *testing.T) {
	t.Setenv("WFX_CLIENT_HOST", "http://127.0.0.1:18080,https://127.0.0.1:18443")
	t.Setenv("WFX_MGMT_HOST", "unix:///tmp/wfx-mgmt.sock")

	flags := NewFlagset()
	require.NoError(t, flags.Parse(nil))
	cfg, err := NewAppConfig(flags)
	require.NoError(t, err)
	t.Cleanup(cfg.Stop)

	assert.Equal(t, []ListenAddr{
		{Network: "tcp", Addr: "127.0.0.1:18080"},
		{Network: "tcp", Addr: "127.0.0.1:18443", TLS: true},
	}, cfg.ClientHosts())
	assert.Equal(t, []ListenAddr{{Network: "unix", Addr: "/tmp/wfx-mgmt.sock"}}, cfg.MgmtHosts())
}

func TestNewAppConfig_UnknownOptionWarn(t *testing.T) {
	t.Setenv("WFX_FOO", "bar")
	t.Setenv("WFX_Mixed_Case", "x")

	cfgFile, err := os.CreateTemp("", "config.yaml")
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.Remove(cfgFile.Name()) })
	_, err = cfgFile.WriteString("unknown-file-opt: 42\n")
	require.NoError(t, err)
	require.NoError(t, cfgFile.Close())

	flags := NewFlagset()
	require.NoError(t, flags.Parse([]string{"--" + ConfigFlag, cfgFile.Name()}))

	oldStderr := os.Stderr
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stderr = w

	cfg, err := NewAppConfig(flags)

	_ = w.Close()
	os.Stderr = oldStderr

	require.NoError(t, err)
	t.Cleanup(cfg.Stop)

	var buf strings.Builder
	_, _ = io.Copy(&buf, r)
	_ = r.Close()

	out := buf.String()
	assert.Contains(t, out, "WARN: Ignoring unknown config option 'WFX_FOO' from env\n")
	assert.Contains(t, out, "WARN: Ignoring unknown config option 'WFX_Mixed_Case' from env\n")
	assert.Contains(t, out, "WARN: Ignoring unknown config option 'unknown-file-opt' from file\n")
}

func TestNewAppConfig_CORSWildcardOriginWithCredentials(t *testing.T) {
	flags := NewFlagset()
	_ = flags.Parse([]string{"--" + CORSAllowCredentialsFlag})

	cfg, err := NewAppConfig(flags)
	assert.Nil(t, cfg)
	assert.Error(t, err)
}

func TestOAuthOpts(t *testing.T) {
	flags := NewFlagset()
	require.NoError(t, flags.Parse([]string{
		"--oauth-issuer", "https://id.example",
		"--" + OAuthClientIDFlag, "wfx-ui",
	}))
	cfg, err := NewAppConfig(flags)
	require.NoError(t, err)
	t.Cleanup(cfg.Stop)

	assert.Equal(t, OAuthOpts{
		Issuer:   "https://id.example",
		ClientID: "wfx-ui",
		Scope:    "openid email profile",
	}, cfg.OAuthOpts())
}

func TestReload(t *testing.T) {
	dir, _ := os.MkdirTemp("", "TestReload")
	cfgFile, _ := os.CreateTemp("", "config.yaml")
	t.Cleanup(func() {
		_ = cfgFile.Close()
		_ = os.RemoveAll(dir)
	})
	_, _ = cfgFile.Write([]byte("log-level: trace"))

	f := NewFlagset()
	_ = f.Parse([]string{"--config", cfgFile.Name()})
	cfg, err := NewAppConfig(f)
	defer cfg.Stop()
	require.NoError(t, err)

	assert.Equal(t, zerolog.TraceLevel.String(), cfg.LogLevel().String())

	{ // modify config file
		_, _ = cfgFile.Seek(0, 0)
		_, _ = cfgFile.Write([]byte("log-level: error"))
	}

	for range 500 {
		if zerolog.GlobalLevel() == zerolog.ErrorLevel {
			break
		}
		time.Sleep(time.Millisecond * 10)
	}
	assert.Equal(t, zerolog.ErrorLevel.String(), zerolog.GlobalLevel().String())
}
