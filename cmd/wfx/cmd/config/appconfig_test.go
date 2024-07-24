package config

/*
 * SPDX-FileCopyrightText: 2024 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAppConfig(t *testing.T) {
	cfg, err := NewAppConfig(NewFlagset())
	defer cfg.Stop()
	require.NoError(t, err)

	// call all methods which do not accept arguments
	structValue := reflect.ValueOf(cfg)
	for i := 0; i < structValue.NumMethod(); i++ {
		method := structValue.Method(i)
		methodType := method.Type()
		if methodType.NumIn() == 0 {
			t.Run(methodType.Name(), func(*testing.T) {
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

	for i := 0; i < 500; i++ {
		if zerolog.GlobalLevel() == zerolog.ErrorLevel {
			break
		}
		time.Sleep(time.Millisecond * 10)
	}
	assert.Equal(t, zerolog.ErrorLevel.String(), zerolog.GlobalLevel().String())
}
