package root

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fmsg"
	"github.com/knadh/koanf/v2"
	"github.com/siemens/wfx/api"
	"github.com/siemens/wfx/generated/southbound/restapi"
	"github.com/siemens/wfx/internal/server"
	"github.com/siemens/wfx/middleware"
	"github.com/siemens/wfx/persistence"
)

func createSouthboundServers(schemes []string, storage persistence.Storage) ([]myServer, error) {
	var settings server.HTTPSettings
	k.Read(func(k *koanf.Koanf) {
		settings.Host = k.String(clientHostFlag)
		settings.TLSHost = k.String(clientTLSHostFlag)
		settings.Port = k.Int(clientPortFlag)
		settings.TLSPort = k.Int(clientTLSPortFlag)
		settings.UDSPath = k.String(clientUnixSocket)
	})
	swaggerJSON, _ := restapi.SwaggerJSON.MarshalJSON()
	api, err := api.NewSouthboundAPI(storage)
	if err != nil {
		return nil, fault.Wrap(err, fmsg.With("Failed to create southbound API"))
	}
	cfg := middleware.Config{
		Config:      k,
		Storage:     storage,
		BasePath:    api.Context().BasePath(),
		SwaggerJSON: swaggerJSON,
	}
	// add our global middlewares
	handler, err := middleware.SetupGlobalMiddleware(cfg, restapi.ConfigureAPI(api))
	if err != nil {
		return nil, fault.Wrap(err)
	}

	servers, err := createServers(schemes, handler, settings)
	return servers, fault.Wrap(err)
}
