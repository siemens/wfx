//go:build ui

package root

/*
 * SPDX-FileCopyrightText: 2025 Siemens AG
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

func TestUIRedirect(t *testing.T) {
	dbMock := persistence.NewHealthyMockStorage(t)
	sc, err := NewServerCollection(new(config.AppConfig), dbMock)
	require.NotNil(t, sc)
	require.NoError(t, err)

	apitest.New().
		Handler(sc.north.Handler).
		Get("/ui").
		Expect(t).
		Status(http.StatusMovedPermanently).
		End()

	apitest.New().
		Handler(sc.south.Handler).
		Get("/ui").
		Expect(t).
		Status(http.StatusNotFound).
		End()
}

func TestUI(t *testing.T) {
	dbMock := persistence.NewHealthyMockStorage(t)
	sc, err := NewServerCollection(new(config.AppConfig), dbMock)
	require.NotNil(t, sc)
	require.NoError(t, err)

	apitest.New().
		Handler(sc.north.Handler).
		Get("/ui/").
		Expect(t).
		Status(http.StatusOK).
		End()

	apitest.New().
		Handler(sc.south.Handler).
		Get("/ui/").
		Expect(t).
		Status(http.StatusNotFound).
		End()
}
