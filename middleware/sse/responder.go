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
	"time"

	"github.com/Southclaws/fault"
	"github.com/siemens/wfx/middleware/logging"
)

// Responder streams server-sent events (SSE) to a client.
// It listens for events from the provided channel and dispatches them
// to the client as soon as they arrive.
type Responder[T any] struct {
	// ctx is the context used to manage the lifecycle of the SSE stream.
	// When the context is canceled, the Responder stops streaming events.
	ctx context.Context
	// idleDuration specifies the duration of inactivity before a keep-alive
	// event is sent to the client. This helps ensure the connection remains
	// open even when no events occur.
	idleDuration time.Duration
	// eventChan is a read-only channel from which the Responder receives events
	// to be sent to the client. Each event is transmitted as soon as it is received.
	eventChan <-chan T
}

func NewResponder[T any](ctx context.Context, idleDuration time.Duration, eventChan <-chan T) Responder[T] {
	return Responder[T]{ctx: ctx, idleDuration: idleDuration, eventChan: eventChan}
}

func (responder Responder[T]) VisitGetJobsEventsResponse(w http.ResponseWriter) error {
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
	_, _ = bufrw.WriteString("X-Accel-Buffering: no\r\n") // notify reverse proxy to disable buffering

	// finish header section
	_, _ = bufrw.WriteString("\r\n")

	if err := bufrw.Flush(); err != nil {
		log.Err(err).Msg("Failed to send event to client")
		return fault.Wrap(err)
	}

	idleTicker := time.NewTicker(responder.idleDuration)
	defer idleTicker.Stop()

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
			b, err := json.Marshal(ev)
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
		case <-idleTicker.C:
			log.Debug().Msg("Sending keep-alive message to client")
			_, err = bufrw.WriteString(": keepalive\n\n")
			if err != nil {
				log.Err(err).Msg("Error writing keepalive")
				return fault.Wrap(err)
			}
			if err := bufrw.Flush(); err != nil {
				log.Err(err).Msg("Failed to send event to client")
				return fault.Wrap(err)
			}
		case <-responder.ctx.Done():
			// this typically happens when the client closes the connection
			log.Debug().Msg("Client disconnected")
			break Loop
		}
	}
	log.Info().Msg("Event Subscriber finished")
	return nil
}
