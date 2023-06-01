package jsonutil

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/Southclaws/fault"
	"github.com/go-openapi/runtime"
	"github.com/itchyny/gojq"
	"github.com/rs/zerolog/log"
)

// JSONProducer creates a new JSON producer. Nil slices are mapped to an empty array instead of "null".
// This is a workaround for https://github.com/golang/go/issues/27589
func JSONProducer() runtime.Producer {
	return runtime.ProducerFunc(func(writer io.Writer, data any) error {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return fault.Wrap(err)
		}

		if rw, ok := writer.(http.ResponseWriter); ok {
			filter := rw.Header().Get("X-Response-Filter")
			if filter != "" {
				var input any
				// need to unmarshal again, but to type 'any'
				if err := json.Unmarshal(jsonData, &input); err != nil {
					return fault.Wrap(err)
				}

				contextLogger := log.With().Str("filter", filter).Logger()
				query, err := gojq.Parse(filter)
				if err != nil {
					contextLogger.Error().Err(err).Msg("Failed to parse response filter")
					return fault.Wrap(err)
				}
				contextLogger.Debug().Msg("Applying response filter")
				iter := query.Run(input)
				ok := true
				for ok {
					var v any
					v, ok = iter.Next()
					if ok {
						jsonData, err = json.Marshal(v)
						if err != nil {
							return fault.Wrap(err)
						}
						_, err = writer.Write(jsonData)
						if err != nil {
							return fault.Wrap(err)
						}
					}
				}
				return nil
			}
		}

		_, err = writer.Write(jsonData)
		return fault.Wrap(err)
	})
}
