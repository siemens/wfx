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
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/siemens/wfx/generated/plugin"
	"github.com/siemens/wfx/generated/plugin/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type StartFailPlugin struct{}

func (p StartFailPlugin) Name() string { return "StartFailPlugin" }
func (p StartFailPlugin) Start(chan error) (chan Message, error) {
	return nil, errors.New("failed to start plugin")
}
func (p StartFailPlugin) Stop() error { return nil }

func TestNewMiddleware_StartFails(t *testing.T) {
	p := StartFailPlugin{}
	mw, err := NewMiddleware(p, make(chan error))
	assert.Error(t, err)
	assert.Nil(t, mw)
}

type TestPlugin struct{ chMessage chan Message }

func NewTestPlugin() *TestPlugin { return &TestPlugin{chMessage: make(chan Message)} }

func (p *TestPlugin) Name() string { return "TestPlugin" }

func (p *TestPlugin) Start(chan error) (chan Message, error) { return p.chMessage, nil }

func (p *TestPlugin) Stop() error { return nil }

func TestNewMiddleware_ModifyRequest(t *testing.T) {
	p := NewTestPlugin()
	go func() {
		for msg := range p.chMessage {
			t.Log("Sending response")
			msg.response <- plugin.PluginResponseT{
				Cookie: msg.request.Cookie,
				Payload: &plugin.PayloadT{
					Type: plugin.Payloadgenerated_plugin_client_Request,
					Value: &client.RequestT{
						Action: client.ActionRead,
						Envelope: []*client.EnvelopeT{
							{Name: "User-Agent", Values: []string{"gotest"}},
						},
						Destination: "localhost/foo/bar",
					},
				},
			}
		}
	}()

	mw, err := NewMiddleware(p, make(chan error))
	require.Nil(t, err)
	defer mw.Stop()

	handler := mw.Middleware()(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	recorder := httptest.NewRecorder()
	httpReq := &http.Request{
		Method: http.MethodGet,
		URL: &url.URL{
			Host: "localhost",
			Path: "/foo",
		},
	}
	handler.ServeHTTP(recorder, httpReq)
	assert.Equal(t, http.StatusOK, recorder.Code)
}

func TestNewMiddleware_SendResponse(t *testing.T) {
	p := NewTestPlugin()
	go func() {
		for msg := range p.chMessage {
			t.Log("Sending response")
			msg.response <- plugin.PluginResponseT{
				Cookie: msg.request.Cookie,
				Payload: &plugin.PayloadT{
					Type: plugin.Payloadgenerated_plugin_client_Response,
					Value: &client.ResponseT{
						Status: client.ResponseStatusAccept,
						Envelope: []*client.EnvelopeT{
							{Name: "User-Agent", Values: []string{"gotest"}},
						},
						Content: []byte{},
					},
				},
			}
		}
	}()

	mw, err := NewMiddleware(p, make(chan error))
	require.Nil(t, err)
	defer mw.Stop()

	handler := mw.Middleware()(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	recorder := httptest.NewRecorder()
	httpReq := &http.Request{
		Method: http.MethodGet,
		URL: &url.URL{
			Host: "localhost",
			Path: "/foo",
		},
	}
	handler.ServeHTTP(recorder, httpReq)
	assert.Equal(t, http.StatusOK, recorder.Code)
}

func TestNewMiddleware_Unavailable(t *testing.T) {
	p := NewTestPlugin()
	go func() {
		for msg := range p.chMessage {
			msg.response <- plugin.PluginResponseT{
				Cookie: msg.request.Cookie,
				Payload: &plugin.PayloadT{
					Type: plugin.Payloadgenerated_plugin_client_Response,
					Value: &client.ResponseT{
						Status: client.ResponseStatusUnavailable,
						Envelope: []*client.EnvelopeT{
							{Name: "User-Agent", Values: []string{"gotest"}},
						},
						Content: []byte{},
					},
				},
			}
		}
	}()

	mw, err := NewMiddleware(p, make(chan error))
	require.Nil(t, err)
	defer mw.Stop()

	handler := mw.Middleware()(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	recorder := httptest.NewRecorder()
	httpReq := &http.Request{
		Method: http.MethodGet,
		URL: &url.URL{
			Host: "localhost",
			Path: "/foo",
		},
	}
	handler.ServeHTTP(recorder, httpReq)
	assert.Equal(t, http.StatusServiceUnavailable, recorder.Code)
}

func TestHttpMethodToAction(t *testing.T) {
	assert.Equal(t, client.ActionCreate, httpMethodToAction(http.MethodPost))
	assert.Equal(t, client.ActionUpdate, httpMethodToAction(http.MethodPut))
	assert.Equal(t, client.ActionUpdate, httpMethodToAction(http.MethodPatch))
	assert.Equal(t, client.ActionDelete, httpMethodToAction(http.MethodDelete))
	otherMethods := []string{
		http.MethodGet,
		http.MethodConnect,
		http.MethodOptions,
		http.MethodTrace,
	}
	for _, action := range otherMethods {
		assert.Equal(t, client.ActionRead, httpMethodToAction(action))
	}
}

func TestNewMiddleware_InvalidResponse(t *testing.T) {
	p := NewTestPlugin()
	go func() {
		// receive message
		msg := <-p.chMessage
		// send response
		msg.response <- plugin.PluginResponseT{
			Cookie: msg.request.Cookie,
			Payload: &plugin.PayloadT{
				Type: plugin.Payload(42),
			},
		}
	}()

	mw, err := NewMiddleware(p, make(chan error))
	require.Nil(t, err)
	defer mw.Stop()

	handler := mw.Middleware()(nil)
	recorder := httptest.NewRecorder()
	httpReq := &http.Request{
		Method: http.MethodGet,
		URL: &url.URL{
			Host: "localhost",
			Path: "/foo",
		},
	}
	handler.ServeHTTP(recorder, httpReq)
	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
}
