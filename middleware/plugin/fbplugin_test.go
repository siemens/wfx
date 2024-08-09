package plugin

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/siemens/wfx/generated/plugin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPlugin(t *testing.T) {
	p := NewFBPlugin("true", nil)
	assert.NotNil(t, p)
}

func TestNewPluginEmpty(t *testing.T) {
	p := NewFBPlugin("", nil)
	assert.NotNil(t, p)
}

func TestStart_NotFound(t *testing.T) {
	p := NewFBPlugin("foobar", nil)
	assert.NotNil(t, p)
}

func TestStopWithoutStart(t *testing.T) {
	p := NewFBPlugin("true", nil)
	err := p.Stop()
	assert.NoError(t, err)
}

func TestStop(t *testing.T) {
	chQuit := make(chan error)
	p := NewFBPlugin("cat", chQuit)

	ch, err := p.Start()
	t.Cleanup(func() { close(ch) })
	require.NoError(t, err)

	err = p.Stop()
	assert.NoError(t, err)

	err = p.Stop()
	assert.NoError(t, err)
}

func TestSendAndReceive(t *testing.T) {
	chQuit := make(chan error)
	p := NewFBPlugin("cat", chQuit)
	ch, err := p.Start()
	t.Cleanup(func() { close(ch) })
	require.NoError(t, err)
	t.Cleanup(func() { _ = p.Stop() })

	headers := make(map[string][]string)
	headers["Content-Type"] = []string{"application/json"}

	httpReq := http.Request{
		Method: http.MethodGet,
		URL: &url.URL{
			Host: "localhost",
			Path: "/foo",
		},
		Header: headers,
	}
	req := convertRequest(&httpReq, 1)
	msg := Message{
		request:  req,
		response: make(chan plugin.PluginResponseT, 1),
	}
	// send message to plugin
	ch <- msg
	// wait for response
	resp := <-msg.response
	assert.Equal(t, req.Cookie, resp.Cookie)
}

func TestName(t *testing.T) {
	assert.Equal(t, "true", NewFBPlugin("true", nil).Name())
}

func TestStart_Reaper(t *testing.T) {
	chQuit := make(chan error)
	p := NewFBPlugin("cat", chQuit)
	ch, err := p.Start()
	assert.NoError(t, err)

	_ = p.terminateProcess()

	err = <-chQuit
	assert.NotNil(t, err)

	t.Cleanup(func() { close(ch) })
}
