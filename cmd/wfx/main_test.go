package main

/*
 * SPDX-FileCopyrightText: 2026 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fmsg"
	"github.com/Southclaws/fault/ftag"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	configureErrorStackMarshaler()
	goleak.VerifyTestMain(m)
}

func TestConfigureErrorStackMarshaler(t *testing.T) {
	originalMarshaler := zerolog.ErrorStackMarshaler
	t.Cleanup(func() {
		zerolog.ErrorStackMarshaler = originalMarshaler //nolint:reassign // restore global test state
	})
	configureErrorStackMarshaler()

	sentinel := errors.New("sentinel")
	err := fault.Wrap(fault.Wrap(sentinel, fmsg.With("operation failed"), ftag.With(ftag.NotFound)))

	stack := zerolog.ErrorStackMarshaler(err)
	encoded, marshalErr := json.Marshal(stack)
	require.NoError(t, marshalErr)
	assert.Contains(t, string(encoded), "main_test.go")
	assert.Contains(t, string(encoded), "operation failed")
	assert.Contains(t, string(encoded), "sentinel")
	assert.True(t, errors.Is(err, sentinel))
	assert.Equal(t, ftag.NotFound, ftag.Get(err))
}
