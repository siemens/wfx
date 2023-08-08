package logging

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type key int

const (
	KeyRequestLogger key = iota
)

type MW struct{}

func (mw MW) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		reqID := uuid.New().String()
		contextLogger := log.With().
			Str("reqID", reqID).
			Str("remoteAddr", r.RemoteAddr).
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Str("host", r.Host).
			Bool("tls", r.TLS != nil).
			Logger()
		contextLogger.Debug().Msg("Processing incoming request")

		l := log.With().Str("reqID", reqID).Logger()
		ctx := context.WithValue(r.Context(), KeyRequestLogger, l)
		r = r.WithContext(ctx)

		if contextLogger.GetLevel() <= zerolog.TraceLevel {

			myResponseWriter := newMyResponseWriter(w)
			myRequestReader := newMyRequestReader(r)
			next.ServeHTTP(myResponseWriter, r)

			request := myRequestReader.requestBody.Bytes()

			contextLogger.Trace().
				Bytes("request", request).
				Msg("Request")
			if len(myResponseWriter.responseBody.Bytes()) > 0 {
				contextLogger.Trace().
					RawJSON("response", myResponseWriter.responseBody.Bytes()).
					Msg("Response")
			}
		} else {
			next.ServeHTTP(w, r)
		}
		contextLogger.Debug().
			TimeDiff("duration", time.Now(), start).
			Msg("Finished request")
	})
}

func (mw MW) Shutdown() {}

func LoggerFromCtx(ctx context.Context) zerolog.Logger {
	if log, ok := ctx.Value(KeyRequestLogger).(zerolog.Logger); ok {
		return log
	}
	return log.Logger
}
