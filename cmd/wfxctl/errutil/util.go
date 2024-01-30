package errutil

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
	"io"
	"reflect"

	"github.com/rs/zerolog/log"
	"github.com/siemens/wfx/generated/model"
)

type codeFn interface {
	Code() int
}

// ProcessErrorResponse extracts the Payload from an ErrorResponse.
func ProcessErrorResponse(w io.Writer, err error) {
	if errors := extractErrors(err); len(errors) > 0 {
		for _, msg := range errors {
			fmt.Fprintf(w, "ERROR: %s (code=%s, logref=%s)\n", msg.Message, msg.Code, msg.Logref)
		}
		return
	}
	var fn codeFn
	if errors.As(err, &fn) {
		fmt.Fprintf(w, "ERROR: HTTP status %d\n", fn.Code())
		return
	}
	fmt.Fprintf(w, "ERROR: %s\n", err)
}

func extractErrors(val any) []*model.Error {
	/* The below code is the generalization of the following snippet
	* and works for _all_ returned errors, since they all share the same structure thanks to the
	* ErrorType in the OpenAPI spec:

	 if resp, ok := err.(*jobs.PostJobsBadRequest); ok {
	     return resp.Payload.Errors
	 }
	*/
	var zeroValue reflect.Value
	if value := reflect.ValueOf(val); value.Kind() == reflect.Ptr {
		if field := value.Elem().FieldByName("Payload"); field != zeroValue {
			if resp, ok := field.Interface().(*model.ErrorResponse); ok {
				return resp.Errors
			}
		}
	}
	return nil
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
