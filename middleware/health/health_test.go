package health

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"context"
	"encoding/json"
	"io"
	"testing"

	"github.com/alexliesenfeld/health"
	"github.com/steinfletcher/apitest"
	"github.com/stretchr/testify/require"
)

func TestNewHealthMiddleware(t *testing.T) {
	mw := NewHealthMiddleware(nil)
	defer mw.Shutdown()

	result := apitest.New().
		Handler(mw.Wrap(nil)).
		Get("/health").
		Expect(t).
		End()

	raw, err := io.ReadAll(result.Response.Body)
	require.NoError(t, err)
	var checkResult health.CheckerResult
	err = json.Unmarshal(raw, &checkResult)
	require.NoError(t, err)
}

func TestStatusListener(*testing.T) {
	statusListener(context.Background(), health.CheckerState{Status: health.StatusUp})
	statusListener(context.Background(), health.CheckerState{Status: health.StatusDown})
	statusListener(context.Background(), health.CheckerState{Status: health.StatusUnknown})
	statusListener(context.Background(), health.CheckerState{
		Status: health.StatusUp,
		CheckState: map[string]health.CheckState{
			"db": {Status: health.StatusUp},
		},
	})
}
