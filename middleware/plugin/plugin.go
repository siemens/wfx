package plugin

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	generated "github.com/siemens/wfx/generated/plugin"
)

type Plugin interface {
	// Name returns a name identifying the plugin.
	Name() string

	// Start starts the plugin but does not wait for it to complete.
	//
	// It returns a channel which must be used to deliver messages to the plugin.
	// This channel must be closed when there are no more messages to be delivered.
	//
	// After a successful call to Start the Stop method must be called in
	// order to release associated system resources.
	Start(chan error) (chan Message, error)

	// Stop stops the plugin and release system resources.
	Stop() error
}

type Message struct {
	// Channel for the messages to be sent to the plugin
	request *generated.PluginRequestT
	// Channel to receive the plugin responses
	response chan generated.PluginResponseT
}
