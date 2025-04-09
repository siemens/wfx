package sse

/*
 * SPDX-FileCopyrightText: 2024 Siemens AG
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

	"github.com/Southclaws/fault"
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
type Responder struct {
	ctx       context.Context
	eventChan <-chan emitter.Event
}

func NewResponder(ctx context.Context, eventChan <-chan emitter.Event) Responder {
	return Responder{ctx: ctx, eventChan: eventChan}
}

func (responder Responder) VisitGetJobsEventsResponse(w http.ResponseWriter) error {
	log := logging.LoggerFromCtx(responder.ctx)

	// Check if the ResponseWriter supports hijacking
	hj, ok := w.(http.Hijacker)
	if !ok { // e.g. if HTTP/2 is being used
		return fmt.Errorf("http.Hijacker interface not supported")
	}

	// use raw connection to prevent it being closed by http.Server's idle cleanup routines
	conn, bufrw, err := hj.Hijack()
	if err != nil {
		return fmt.Errorf("failed to hijack connection: %w", err)
	}
	defer func() {
		_ = conn.Close()
	}()

	_, _ = fmt.Fprintf(bufrw, "HTTP/1.1 %d\r\n", http.StatusOK)
	_, _ = bufrw.WriteString("Content-Type: text/event-stream\r\n")
	_, _ = bufrw.WriteString("Cache-Control: no-cache\r\n")
	_, _ = bufrw.WriteString("Connection: keep-alive\r\n")
	_, _ = bufrw.WriteString("Access-Control-Allow-Origin: *\r\n")
	// finish header section
	_, _ = bufrw.WriteString("\r\n")

	if err := bufrw.Flush(); err != nil {
		log.Err(err).Msg("Failed to send event to client")
		return fault.Wrap(err)
	}

	var id uint64 = 1
Loop:
	for {
		log.Debug().Msg("Waiting for next event")
		select {
		case ev, ok := <-responder.eventChan:
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

			// must end with two newlines as required by the SSE spec:
			_, err = fmt.Fprintf(bufrw, "data: %s\nid: %d\n\n", b, id)
			if err != nil {
				log.Err(err).Msg("Cannot write to buffer")
				return fault.Wrap(err)
			}

			if err := bufrw.Flush(); err != nil {
				log.Err(err).Msg("Failed to send event to client")
				return fault.Wrap(err)
			}

			id++
		case <-responder.ctx.Done():
			// this typically happens when the client closes the connection
			log.Debug().Msg("Context is done")
			break Loop
		}
	}
	log.Info().Msg("Event Subscriber finished")
	return nil
}
