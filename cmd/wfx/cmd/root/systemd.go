package root

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"errors"
	"fmt"
	"net"
	"net/http"

	"github.com/rs/zerolog/log"
	"github.com/siemens/wfx/persistence"
)

func adoptSystemdSockets(listeners []net.Listener, storage persistence.Storage, errChan chan error) error {
	count := len(listeners)
	if count == 0 {
		return nil
	}
	if count != 2 {
		return fmt.Errorf("invalid fd count: %d", count)
	}

	log.Debug().
		Int("count", count).
		Msg("Received socket fds from systemd")

	go func() {
		srv, err := createSouthboundServer(storage, kindHTTP)
		if err != nil {
			errChan <- err
			return
		}

		log.Info().Msg("Starting southbound UDS listener (activated by systemd)")
		l := listeners[0]
		err = srv.Srv.Serve(l)
		if err == nil || errors.Is(err, http.ErrServerClosed) {
			return
		}
		errChan <- err
	}()

	go func() {
		srv, err := createNorthboundServer(storage, kindHTTP)
		if err != nil {
			errChan <- err
			return
		}

		log.Info().Msg("Starting northbound UDS listener (activated by systemd)")
		l := listeners[1]
		err = srv.Srv.Serve(l)
		if err == nil || errors.Is(err, http.ErrServerClosed) {
			return
		}
		errChan <- err
	}()

	return nil
}
