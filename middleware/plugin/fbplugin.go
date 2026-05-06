package plugin

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"sync/atomic"

	"github.com/Southclaws/fault"
	"github.com/rs/zerolog/log"
	genPlugin "github.com/siemens/wfx/generated/plugin"
	"github.com/siemens/wfx/middleware/plugin/ioutil"
)

// compile-time check to ensure we fulfill the interface
var _ Plugin = (*FBPlugin)(nil)

// FBPlugin is a plugin which communicates using FlatBuffer messages.
type FBPlugin struct {
	path string

	responses      map[uint64]chan genPlugin.PluginResponseT
	responsesMutex sync.Mutex

	cmd        *exec.Cmd
	waited     atomic.Bool
	stopCalled atomic.Bool
	chErr      chan error
}

// NewFBPlugin creates a new plugin instance. In order to start the plugin, call
// the Start() function.
func NewFBPlugin(path string) *FBPlugin {
	return &FBPlugin{path: path}
}

func (p *FBPlugin) Name() string {
	return p.path
}

func (p *FBPlugin) Start(chErr chan error) (chan Message, error) {
	log.Info().Str("path", p.path).Msgf("Starting plugin %q", p.path)
	cmd := createCmd(p.path)

	// this ensures that a process group is created (needed to kill all child processes)
	p.responses = make(map[uint64]chan genPlugin.PluginResponseT)
	p.chErr = chErr

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fault.Wrap(err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fault.Wrap(err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fault.Wrap(err)
	}

	chMessage := make(chan Message)

	if err := cmd.Start(); err != nil {
		return nil, fault.Wrap(err)
	}
	log.Debug().Str("path", cmd.Path).Msgf("Plugin %q started", cmd.Path)

	go func() { // our reaper
		defer close(chErr)
		_ = cmd.Wait()
		log.Debug().Msg("Plugin subprocess has exited")
		p.waited.Store(true)
		if !p.stopCalled.Load() {
			chErr <- fmt.Errorf("plugin '%s' stopped unexpectedly", p.Name())
		}
	}()

	go p.sender(stdin, chMessage)
	go p.receiver(stdout)
	go p.forwardLogs(stderr)

	p.cmd = cmd

	return chMessage, nil
}

func (p *FBPlugin) Stop() error {
	log.Info().Str("path", p.path).Msgf("Stopping plugin %q", p.path)
	alreadyStopped := p.stopCalled.Swap(true)
	alreadyWaited := p.waited.Load()
	if alreadyStopped || alreadyWaited || p.cmd == nil {
		log.Debug().Str("path", p.path).Msgf("Plugin %q already stopped", p.path)
		return nil
	}

	return fault.Wrap(p.terminateProcess())
}

func (p *FBPlugin) sender(w io.Writer, chMessage <-chan Message) {
	for msg := range chMessage {
		p.responsesMutex.Lock()
		p.responses[msg.request.Cookie] = msg.response
		p.responsesMutex.Unlock()

		if err := ioutil.WriteRequest(w, msg.request); err != nil {
			log.Error().Err(err).Msg("Failed to write message")
			break
		}
		log.Debug().Uint64("cookie", msg.request.Cookie).Msgf("Request with cookie %d sent to plugin", msg.request.Cookie)
	}
	log.Info().Str("name", p.Name()).Msg("Plugin writer stopped")
}

func (p *FBPlugin) receiver(r io.Reader) {
	for !p.waited.Load() {
		resp, err := ioutil.ReadResponse(r)
		if err != nil {
			if errors.Is(err, os.ErrClosed) || errors.Is(err, io.EOF) {
				break
			}
			log.Error().Err(err).Msg("Failed to read message")
			continue
		}

		cookie := resp.Cookie
		log.Debug().Uint64("cookie", cookie).Msgf("Received plugin response for cookie %d", cookie)
		p.responsesMutex.Lock()
		chResp, ok := p.responses[cookie]
		delete(p.responses, cookie)
		p.responsesMutex.Unlock()
		if !ok {
			log.Error().Uint64("cookie", cookie).Msgf("Received unexpected response from plugin for cookie %d", cookie)
			_ = p.terminateProcess() // this results in wfx stopping gracefully because the plugin stops without Stop() being called
			break
		}
		chResp <- *resp
		close(chResp) // there can only be one response
	}
	log.Info().Str("name", p.Name()).Msg("Plugin receiver stopped")
}

func (p *FBPlugin) forwardLogs(r io.Reader) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		msg := scanner.Text()
		log.Debug().Str("path", p.path).Str("msg", msg).Msg("Plugin log message")
	}
	log.Info().Str("name", p.Name()).Msg("Log forwarder stopped")
}
