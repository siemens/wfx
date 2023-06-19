package util

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"net/http"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	"github.com/rs/zerolog/log"
	"github.com/siemens/wfx/internal/producer"
)

// ForceJSONResponse generates a JSON response using the provided payload.
func ForceJSONResponse(statusCode int, payload any) middleware.ResponderFunc {
	return func(rw http.ResponseWriter, _ runtime.Producer) {
		rw.Header().Add("Content-Type", "application/json")
		rw.WriteHeader(statusCode)
		if err := producer.JSONProducer().Produce(rw, payload); err != nil {
			log.Error().Err(err).Msg("Failed to generate JSON response")
		}
	}
}
