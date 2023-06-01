// SPDX-FileCopyrightText: 2023 Siemens AG
//
// SPDX-License-Identifier: Apache-2.0

package restapi

import (
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/rs/zerolog/log"

	"github.com/siemens/wfx/generated/southbound/restapi/operations"
	"github.com/siemens/wfx/internal/jsonutil"
)

func ConfigureAPI(api *operations.WorkflowExecutorAPI) http.Handler {
	// configure the api here
	api.ServeError = func(rw http.ResponseWriter, r *http.Request, err error) {
		log.Error().Msg(err.Error())
		errors.ServeError(rw, r, err)
	}

	api.Logger = log.Printf

	api.UseSwaggerUI()

	api.JSONConsumer = runtime.JSONConsumer()

	api.JSONProducer = jsonutil.JSONProducer()

	api.PreServerShutdown = func() {}

	return setupGlobalMiddleware(api.Serve(setupMiddlewares))
}

// The middleware configuration is for the handler executors. These do not apply to the swagger.json document.
// The middleware executes after routing but before authentication, binding and validation.
func setupMiddlewares(handler http.Handler) http.Handler {
	return handler
}

// The middleware configuration happens before anything, this middleware also applies to serving the swagger.json document.
// So this is a good place to plug in a panic handling middleware, logging and metrics
func setupGlobalMiddleware(handler http.Handler) http.Handler {
	return handler
}
