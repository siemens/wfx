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
	p := NewFBPlugin("true")
	assert.NotNil(t, p)
}

func TestNewPluginEmpty(t *testing.T) {
	p := NewFBPlugin("")
	assert.NotNil(t, p)
}

func TestStart_NotFound(t *testing.T) {
	p := NewFBPlugin("foobar")
	assert.NotNil(t, p)
}

func TestStopWithoutStart(t *testing.T) {
	p := NewFBPlugin("true")
	err := p.Stop()
	assert.NoError(t, err)
}

func TestStop(t *testing.T) {
	p := NewFBPlugin("cat")

	chErr := make(chan error)
	chMessages, err := p.Start(chErr)
	require.NoError(t, err)

	err = p.Stop()
	assert.NoError(t, err)

	err = p.Stop()
	assert.NoError(t, err)

	select {
	case err := <-chErr:
		assert.NoError(t, err)
	default:
		//
	}

	close(chMessages)
}

func TestSendAndReceive(t *testing.T) {
	p := NewFBPlugin("cat")
	chErr := make(chan error)
	chMessages, err := p.Start(chErr)
	require.NoError(t, err)

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
	chMessages <- msg
	// wait for response
	resp := <-msg.response
	assert.Equal(t, req.Cookie, resp.Cookie)

	close(chMessages)
	_ = p.Stop()

	select {
	case err := <-chErr:
		assert.NoError(t, err)
	default:
		//
	}
}

func TestName(t *testing.T) {
	assert.Equal(t, "true", NewFBPlugin("true").Name())
}
