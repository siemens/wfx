package middleware

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"net/http"

	"github.com/Southclaws/fault"
	"github.com/go-openapi/runtime/middleware"
	"github.com/rs/cors"
	"github.com/siemens/wfx/internal/config"
	"github.com/siemens/wfx/middleware/fileserver"
	"github.com/siemens/wfx/middleware/health"
	"github.com/siemens/wfx/middleware/jq"
	"github.com/siemens/wfx/middleware/logging"
	"github.com/siemens/wfx/middleware/swagger"
	"github.com/siemens/wfx/middleware/version"
	"github.com/siemens/wfx/persistence"
)

type Config struct {
	Config      *config.ThreadSafeKoanf
	Storage     persistence.Storage
	BasePath    string
	SwaggerJSON []byte
}

func SetupGlobalMiddleware(config Config, handler http.Handler) (http.Handler, error) {
	handler = logging.NewLoggingMiddleware(handler)

	handler = jq.NewJqMiddleware(handler)

	var err error
	handler, err = fileserver.NewFileServerMiddleware(config.Config, handler)
	if err != nil {
		return nil, fault.Wrap(err)
	}

	// expose swagger.json under basePath (useful if you're using a reverse proxy in front of wfx)
	handler = middleware.Spec(config.BasePath, config.SwaggerJSON, handler)
	// be friendly and tell user about swagger.json
	handler = swagger.NewSpecMiddleware(handler)

	handler = health.NewHealthMiddleware(config.Storage, handler)
	handler = version.NewVersionMiddleware(handler)

	// this is the first handler which is executed in the chain (LIFO):
	handler = cors.AllowAll().Handler(handler)
	return handler, nil
}
