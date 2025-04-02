package logging

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/Southclaws/fault"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type key int

const (
	KeyRequestLogger key = iota
)

func NewLoggingMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			reqID := uuid.New().String()
			var path string
			if r.URL != nil {
				path = r.URL.Path
			}
			contextLogger := log.With().
				Str("reqID", reqID).
				Str("remoteAddr", r.RemoteAddr).
				Str("method", r.Method).
				Str("path", path).
				Str("host", r.Host).
				Bool("tls", r.TLS != nil).
				Logger()
			contextLogger.Debug().Msg("Processing incoming request")

			l := log.With().Str("reqID", reqID).Logger()
			ctx := context.WithValue(r.Context(), KeyRequestLogger, l)
			r = r.WithContext(ctx)

			tracing := contextLogger.GetLevel() <= zerolog.TraceLevel
			writer := newMyResponseWriter(w, tracing)
			if tracing {
				request, err := PeekBody(r)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				contextLogger.Trace().
					Bytes("request", request).
					Msg("Request")
			}

			next.ServeHTTP(writer, r)
			if response := writer.responseBody.Bytes(); len(response) > 0 {
				if json.Valid(response) {
					contextLogger.Trace().RawJSON("response", response).Msg("Response")
				} else { // body is not in JSON format
					contextLogger.Trace().Bytes("response", response).Msg("Response")
				}
			}

			contextLogger.Debug().
				TimeDiff("duration", time.Now(), start).
				Int("code", writer.statusCode).
				Msg("Finished request")
		})
	}
}

func LoggerFromCtx(ctx context.Context) zerolog.Logger {
	if log, ok := ctx.Value(KeyRequestLogger).(zerolog.Logger); ok {
		return log
	}
	return log.Logger
}

func PeekBody(r *http.Request) ([]byte, error) {
	// consume request body
	var request []byte
	if r.Body != nil {
		var err error
		request, err = io.ReadAll(r.Body)
		if err != nil {
			return nil, fault.Wrap(err)
		}
		_ = r.Body.Close()
		// restore original body for other middlewares
		r.Body = io.NopCloser(bytes.NewBuffer(request))
	}
	return request, nil
}
