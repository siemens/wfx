package sse

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	"github.com/siemens/wfx/internal/pubsub/subscriber"
	"github.com/siemens/wfx/middleware/logging"
)

// Responder sends a stream of events to the client.
// This function runs (i.e. blocks the caller) until the channel is closed.
func Responder[V any](ctx context.Context, subscriber *subscriber.Subscriber[V]) middleware.ResponderFunc {
	return func(w http.ResponseWriter, p runtime.Producer) {
		log := logging.LoggerFromCtx(ctx)

		flusher, _ := w.(http.Flusher)

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		flusher.Flush()

		log.Debug().Msg("Waiting for next event")
		for ev := range subscriber.Events() {
			b, err := json.Marshal(ev)
			if err != nil {
				log.Error().Err(err).Msg("Failed to marshal status event")
				continue
			}
			log.Debug().RawJSON("event", b).Msg("Received status event. Notifying client.")

			_, _ = w.Write([]byte("data: "))
			_, _ = w.Write(b)
			// text/event-stream responses are "chunked" with double newline breaks
			_, _ = w.Write([]byte("\n\n"))

			flusher.Flush()

			log.Debug().Msg("Waiting for next status event")
		}
	}
}
