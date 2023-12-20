package plugin

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"errors"
	"net/http"
	"net/url"
	"sync/atomic"
	"time"

	"github.com/Southclaws/fault"
	"github.com/rs/zerolog/log"
	genPlugin "github.com/siemens/wfx/generated/plugin"
	"github.com/siemens/wfx/generated/plugin/client"
	"github.com/siemens/wfx/middleware/logging"
)

type MW struct {
	plugin    Plugin
	chMessage chan Message
	chQuit    chan error
}

func NewMiddleware(plugin Plugin, chQuit chan error) (*MW, error) {
	contextLogger := log.With().Str("plugin", plugin.Name()).Logger()
	contextLogger.Debug().Msg("Creating new plugin middleware")
	chMessage, err := plugin.Start(chQuit)
	if err != nil {
		_ = plugin.Stop()
		return nil, fault.Wrap(err)
	}
	return &MW{
		plugin:    plugin,
		chMessage: chMessage,
		chQuit:    chQuit,
	}, nil
}

func (mw *MW) Wrap(next http.Handler) http.Handler {
	var cookieCounter atomic.Uint64

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := logging.LoggerFromCtx(r.Context()).With().Str("plugin", mw.plugin.Name()).Logger()

		req := convertRequest(r, cookieCounter.Add(1))
		msg := Message{
			request:  req,
			response: make(chan genPlugin.PluginResponseT, 1),
		}

		log.Debug().Msg("Sending request to plugin")
		start := time.Now()
		mw.chMessage <- msg
		log.Debug().Msg("Waiting for plugin response")
		resp := <-msg.response
		duration := time.Since(start)
		log.Info().Dur("duration", duration).Msg("Received plugin response")

		if resp.Payload != nil {
			switch resp.Payload.Type {
			case genPlugin.Payloadgenerated_plugin_client_Response:
				log.Info().Msg("Sending client response provided by plugin")
				val := resp.Payload.Value.(*client.ResponseT)
				for _, h := range val.Envelope {
					for _, value := range h.Values {
						w.Header().Add(h.Name, value)
					}
				}
				switch val.Status {
				case client.ResponseStatusAccept:
					w.WriteHeader(http.StatusOK)
				case client.ResponseStatusModified:
					w.WriteHeader(http.StatusOK)
				case client.ResponseStatusDeny:
					w.WriteHeader(http.StatusForbidden)
				case client.ResponseStatusUnavailable:
					w.WriteHeader(http.StatusServiceUnavailable)
				}
				_, _ = w.Write(val.Content)
				return
			case genPlugin.Payloadgenerated_plugin_client_Request:
				log.Info().Msg("Request was modified by plugin")
				// override http.Request with the response
				val := resp.Payload.Value.(*client.RequestT)

				if parsedURL, err := url.Parse(val.Destination); err != nil {
					log.Err(err).Str("destination", val.Destination).Msg("Failed to parse destination")
				} else {
					r.URL = parsedURL
				}

				// delete existing headers
				for k := range r.Header {
					delete(r.Header, k)
				}
				if len(val.Envelope) > 0 {
					if r.Header == nil {
						r.Header = make(http.Header)
					}
					for _, h := range val.Envelope {
						for _, value := range h.Values {
							r.Header.Add(h.Name, value)
						}
					}
				}
			default:
				// shouldn't happen, but it's possible; maybe a plugin version
				// mismatch or just a poorly written plugin, in any case we
				// don't want to continue.
				log.Error().Int("type", int(resp.Payload.Type)).Msg("Received unsupported payload type from plugin")
				mw.chQuit <- errors.New("unsupported payload type")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
		log.Debug().Msg("Request may continue")
		next.ServeHTTP(w, r)
	})
}

func (mw *MW) Shutdown() {
	close(mw.chMessage)
	if err := mw.plugin.Stop(); err != nil {
		log.Err(err).Str("path", mw.plugin.Name()).Msg("There was an error while stopping the plugin")
	}
}

func convertRequest(r *http.Request, cookie uint64) *genPlugin.PluginRequestT {
	envelope := make([]*client.EnvelopeT, 0, len(r.Header))
	for name, values := range r.Header {
		header := client.EnvelopeT{Name: name, Values: values}
		envelope = append(envelope, &header)
	}

	req := genPlugin.PluginRequestT{
		Cookie: cookie,
		Request: &client.RequestT{
			Action:      httpMethodToAction(r.Method),
			Destination: r.URL.String(),
			Envelope:    envelope,
		},
	}

	body, _ := logging.PeekBody(r)
	if body != nil {
		req.Request.Content = body
	}
	return &req
}

func httpMethodToAction(method string) client.Action {
	switch method {
	case http.MethodPost:
		return client.ActionCreate
	case http.MethodPut:
		return client.ActionUpdate
	case http.MethodPatch:
		return client.ActionUpdate
	case http.MethodDelete:
		return client.ActionDelete
	default:
		return client.ActionRead
	}
}
