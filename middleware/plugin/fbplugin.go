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
	"syscall"

	"github.com/Southclaws/fault"
	"github.com/rs/zerolog/log"
	genPlugin "github.com/siemens/wfx/generated/plugin"
	"github.com/siemens/wfx/middleware/plugin/ioutil"
)

// FBPlugin is a plugin which communicates using FlatBuffer messages.
type FBPlugin struct {
	path string

	responses      map[uint64]chan genPlugin.PluginResponseT
	responsesMutex sync.Mutex

	cmd     *exec.Cmd
	waited  atomic.Bool
	stopped atomic.Bool

	chQuit chan error
}

// NewFBPlugin creates a new plugin instance. In order to start the plugin, call
// the Start() function.
func NewFBPlugin(path string) *FBPlugin {
	return &FBPlugin{path: path}
}

func (p *FBPlugin) Name() string {
	return p.path
}

func (p *FBPlugin) Start(chQuit chan error) (chan Message, error) {
	log.Info().Str("path", p.path).Msg("Starting plugin")
	cmd := exec.Command(p.path)
	p.chQuit = chQuit

	// this ensures that a process group is created (needed to kill all child processes)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	p.responses = make(map[uint64]chan genPlugin.PluginResponseT)

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
	log.Debug().Str("path", cmd.Path).Msg("Plugin started")

	go func() { // our reaper
		_ = cmd.Wait()
		p.waited.Store(true)
		if !p.stopped.Load() {
			chQuit <- fmt.Errorf("plugin '%s' is down, shutdown necessary", p.Name())
		}
	}()

	go p.sender(stdin, chMessage)
	go p.receiver(stdout)
	go p.forwardLogs(stderr)

	p.cmd = cmd

	return chMessage, nil
}

func (p *FBPlugin) Stop() error {
	log.Info().Str("path", p.path).Msg("Stopping plugin")
	stopped := p.stopped.Swap(true)
	if stopped || p.cmd == nil {
		return nil
	}
	if !p.waited.Load() {
		return fault.Wrap(p.terminateProcess())
	}
	return nil
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
		log.Debug().Uint64("cookie", msg.request.Cookie).Msg("Request sent to plugin")
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
		log.Debug().Uint64("cookie", cookie).Msg("Received plugin response")
		p.responsesMutex.Lock()
		chResp, ok := p.responses[cookie]
		delete(p.responses, cookie)
		p.responsesMutex.Unlock()
		if !ok {
			log.Warn().Uint64("cookie", cookie).Msg("Received unexpected response from plugin")
			p.chQuit <- errors.New("received unexpected response from plugin")
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
