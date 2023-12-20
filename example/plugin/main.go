package main

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/siemens/wfx/generated/plugin"
	"github.com/siemens/wfx/generated/plugin/client"
	"github.com/siemens/wfx/middleware/plugin/ioutil"
)

const queueSize = 64

type QueueEntry struct {
	start   time.Time
	request *plugin.PluginRequestT
}

func main() {
	log.SetOutput(os.Stderr)
	log.Println("[INFO] Plugin starting...")

	queue := make(chan QueueEntry, queueSize)

	// reader goroutine: it's only purpose is to read from stdin as soon as
	// something comes in, to prevent the other side from blocking
	go func() {
		for {
			req, err := ioutil.ReadRequest(os.Stdin)
			if err != nil {
				log.Println("[ERROR] Failed to receive message", err)
				continue
			}

			entry := newQueueEntry(req)
			// this is important: we rather discard the message instead of
			// blocking, since we want to read the next message from stdin
			select {
			case queue <- entry:
				log.Println("[DEBUG] Message enqueued")
				// success
			default:
				log.Println("[ERROR] Queue full. Message discarded.")
				// be nice and inform wfx about it
				resp := plugin.PluginResponseT{
					Cookie: req.Cookie,
					Payload: &plugin.PayloadT{
						Type: plugin.Payloadgenerated_plugin_client_Response,
						Value: &client.ResponseT{
							Status:  client.ResponseStatusUnavailable,
							Content: []byte("Plugin overloaded. Please try again later.\n"),
						},
					},
				}
				if err := ioutil.WriteResponse(os.Stdout, &resp); err != nil {
					log.Println("[ERROR] Failed to send message", err)
				}
			}
		}
	}()

	// here we decide what to do with the request and send the response
	for entry := range queue {
		req := entry.request

		destination := req.Request.Destination
		body := "\n"
		if len(req.Request.Content) > 0 {
			body = string(req.Request.Content)
		}
		log.Printf("[DEBUG] Processing request: cookie=%d, destination=%s, body=%s", req.Cookie, destination, body)

		// prepare the response using the cookie from the request so that wfx
		// can associate the response with the request
		resp := plugin.PluginResponseT{Cookie: req.Cookie}

		// this just an example; prevent access to /workflows
		if strings.Contains(destination, "/api/wfx/v1/workflows") {
			log.Println("[DEBUG] Denying request")
			resp.Payload = &plugin.PayloadT{
				Type: plugin.Payloadgenerated_plugin_client_Response,
				Value: &client.ResponseT{
					Status:  client.ResponseStatusDeny,
					Content: []byte("You are not allowed to access the workflows resource.\n"),
				},
			}
		} else {
			log.Println("[DEBUG] Allowing request")
		}

		if err := ioutil.WriteResponse(os.Stdout, &resp); err != nil {
			log.Println("[ERROR] Failed to send message", err)
		}
		delta := time.Since(entry.start)
		log.Printf("[INFO] Processed request in %0.02f us\n", float64(delta.Nanoseconds())/1_000.)
	}
}

func newQueueEntry(request *plugin.PluginRequestT) QueueEntry {
	return QueueEntry{
		start:   time.Now(),
		request: request,
	}
}
