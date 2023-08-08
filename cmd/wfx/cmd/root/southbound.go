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
	"github.com/rs/cors"
	"github.com/siemens/wfx/api"
	"github.com/siemens/wfx/generated/southbound/restapi"
	"github.com/siemens/wfx/internal/server"
	"github.com/siemens/wfx/middleware"
	"github.com/siemens/wfx/middleware/fileserver"
	"github.com/siemens/wfx/middleware/health"
	"github.com/siemens/wfx/middleware/jq"
	"github.com/siemens/wfx/middleware/logging"
	"github.com/siemens/wfx/middleware/swagger"
	"github.com/siemens/wfx/middleware/version"
	"github.com/siemens/wfx/persistence"
)

func createSouthboundCollection(schemes []string, storage persistence.Storage) (*serverCollection, error) {
	var settings server.HTTPSettings
	k.Read(func(k *koanf.Koanf) {
		settings.Host = k.String(clientHostFlag)
		settings.TLSHost = k.String(clientTLSHostFlag)
		settings.Port = k.Int(clientPortFlag)
		settings.TLSPort = k.Int(clientTLSPortFlag)
		settings.UDSPath = k.String(clientUnixSocket)
	})
	api, err := api.NewSouthboundAPI(storage)
	if err != nil {
		return nil, fault.Wrap(err, fmsg.With("Failed to create southbound API"))
	}

	fsMW, err := fileserver.NewFileServerMiddleware(k)
	if err != nil {
		return nil, fault.Wrap(err)
	}

	swaggerJSON, _ := restapi.SwaggerJSON.MarshalJSON()
	mw := middleware.NewGlobalMiddleware(restapi.ConfigureAPI(api),
		[]middleware.IntermediateMW{
			// LIFO
			logging.MW{},
			jq.MW{},
			fsMW,
			swagger.NewSpecMiddleware(api.Context().BasePath(), swaggerJSON),
			health.NewHealthMiddleware(storage),
			version.MW{},
			middleware.PromoteWrapper(cors.AllowAll().Handler),
		})

	servers, err := createServers(schemes, mw, settings)
	if err != nil {
		return nil, fault.Wrap(err)
	}
	return &serverCollection{
		name:       "southbound",
		servers:    servers,
		middleware: mw,
	}, nil
}
