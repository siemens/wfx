package errutil

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"fmt"
	"io"
	"net/http"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fmsg"
	"github.com/rs/zerolog/log"
	"github.com/siemens/wfx/generated/api"
)

// ProcessErrorResponse extracts the Payload from an ErrorResponse.
func ProcessErrorResponse(w io.Writer, resp api.ErrorResponse) {
	if resp.Errors != nil {
		for _, msg := range *resp.Errors {
			fmt.Fprintf(w, "ERROR: %s (code=%s, logref=%s)\n", msg.Message, msg.Code, msg.Logref)
		}
	}
}

func WithRequestID(err error, resp *http.Response) error {
	if err != nil && resp != nil {
		if id := resp.Header.Get("X-Request-ID"); id != "" {
			return fault.Wrap(err, fmsg.Withf("X-Request-ID: %q", id))
		}
	}
	return err
}

// Must is a utility function that takes a value and an error as parameters and returns the value if the error is nil.
// This function is useful for cases where the program cannot proceed if an error occurs.
// Note: This function will terminate the program if the error is not nil. Use with caution.
func Must[T any](value T, err error) T {
	if err != nil {
		log.Fatal().Err(err).Msg("A fatal error has occurred")
	}
	return value
}
