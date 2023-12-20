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

type MW struct{}

func (mw MW) Wrap(next http.Handler) http.Handler {
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

		if contextLogger.GetLevel() <= zerolog.TraceLevel {
			request, err := PeekBody(r)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			myResponseWriter := newMyResponseWriter(w)
			next.ServeHTTP(myResponseWriter, r)

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
