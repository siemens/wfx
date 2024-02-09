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
	"fmt"
	"net/http"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	"github.com/olebedev/emitter"
	"github.com/siemens/wfx/middleware/logging"
)

// Responder streams server-sent events (SSE) to a client.
// It listens for events from the provided channel and dispatches them
// to the client as soon as they arrive.
//
// Parameters:
// - ctx: The context for managing the lifecycle of the stream. If canceled, streaming stops.
// - source: A read-only channel of events to be transmitted.
func Responder(ctx context.Context, source <-chan emitter.Event) middleware.ResponderFunc {
	return func(w http.ResponseWriter, _ runtime.Producer) {
		log := logging.LoggerFromCtx(ctx)

		flusher, _ := w.(http.Flusher)

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		flusher.Flush()

		var id uint64 = 1
	Loop:
		for {
			log.Debug().Msg("Waiting for next event")
			select {
			case ev, ok := <-source:
				if !ok {
					log.Debug().Msg("SSE channel is closed")
					break Loop
				}
				b, err := json.Marshal(ev.Args[0])
				if err != nil {
					log.Error().Err(err).Msg("Failed to marshal status event")
					continue Loop
				}
				log.Debug().RawJSON("event", b).Msg("Sending event to client")

				_, _ = w.Write([]byte("data: "))
				_, _ = w.Write(b)
				// must end with two newlines as required by the SSE spec:
				_, _ = w.Write([]byte(fmt.Sprintf("\nid: %d\n\n", id)))

				flusher.Flush()

				id++
			case <-ctx.Done():
				// this typically happens when the client closes the connection
				log.Debug().Msg("Context is done")
				break Loop
			}
		}
		log.Info().Msg("Event Subscriber finished")
	}
}
