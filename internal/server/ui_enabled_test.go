//go:build ui

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

	wfxAPI "github.com/siemens/wfx/api"
	"github.com/siemens/wfx/cmd/wfx/cmd/config"
	"github.com/siemens/wfx/persistence"
	"github.com/steinfletcher/apitest"
	"github.com/stretchr/testify/require"
)

func TestUIRedirect(t *testing.T) {
	dbMock := persistence.NewHealthyMockStorage(t)
	wfx := wfxAPI.NewWfxServer(dbMock)
	sc, err := NewServerCollection(new(config.AppConfig), wfx, dbMock)
	require.NotNil(t, sc)
	require.NoError(t, err)

	apitest.New().
		Handler(sc.North.Handler).
		Get("/ui").
		Expect(t).
		Status(http.StatusMovedPermanently).
		End()

	apitest.New().
		Handler(sc.South.Handler).
		Get("/ui").
		Expect(t).
		Status(http.StatusNotFound).
		End()
}

func TestUI(t *testing.T) {
	dbMock := persistence.NewHealthyMockStorage(t)
	wfx := wfxAPI.NewWfxServer(dbMock)
	sc, err := NewServerCollection(new(config.AppConfig), wfx, dbMock)
	require.NotNil(t, sc)
	require.NoError(t, err)

	apitest.New().
		Handler(sc.North.Handler).
		Get("/ui/").
		Expect(t).
		Status(http.StatusOK).
		End()

	apitest.New().
		Handler(sc.South.Handler).
		Get("/ui/").
		Expect(t).
		Status(http.StatusNotFound).
		End()
}
