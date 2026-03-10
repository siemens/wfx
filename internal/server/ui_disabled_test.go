//go:build !ui

package server

/*
 * SPDX-FileCopyrightText: 2026 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"net/http"
	"testing"

	"github.com/siemens/wfx/cmd/wfx/cmd/config"
	"github.com/siemens/wfx/persistence"
	"github.com/steinfletcher/apitest"
	"github.com/stretchr/testify/require"
)

func TestUINotFound(t *testing.T) {
	dbMock := persistence.NewHealthyMockStorage(t)
	sc, err := NewServerCollection(new(config.AppConfig), nil, dbMock)
	require.NotNil(t, sc)
	require.NoError(t, err)

	handlers := []http.Handler{
		sc.North.Handler,
		sc.South.Handler,
	}

	for _, handler := range handlers {
		apitest.New().
			Handler(handler).
			Get("/ui").
			Expect(t).
			Status(http.StatusNotFound).
			End()

		apitest.New().
			Handler(handler).
			Get("/ui/").
			Expect(t).
			Status(http.StatusNotFound).
			End()
	}
}
