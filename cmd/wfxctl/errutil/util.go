package errutil

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"

	"github.com/rs/zerolog/log"
	"github.com/siemens/wfx/generated/model"
)

// ProcessErrorResponse extracts the Payload from an ErrorResponse.
func ProcessErrorResponse(w io.Writer, err error) {
	errors := extractErrors(err)
	if len(errors) > 0 {
		for _, msg := range errors {
			fmt.Fprintf(w, "ERROR: %s (code=%s, logref=%s)\n", msg.Message, msg.Code, msg.Logref)
		}
	} else {
		b, _ := json.Marshal(err)
		fmt.Fprintf(w, "ERROR: %s\n", b)
	}
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
	value := reflect.ValueOf(val)
	if value.Kind() == reflect.Ptr {
		value2 := value.Elem()
		field := value2.FieldByName("Payload")
		if field != zeroValue {
			if resp, ok := field.Interface().(*model.ErrorResponse); ok {
				return resp.Errors
			}
		}
	}
	return []*model.Error{}
}

// Must is a utility function that takes a value and an error as parameters and returns the value if the error is nil.
// This function is useful for cases where the program cannot proceed if an error occurs.
// Note: This function will terminate the program if the error is not nil. Use with caution.
func Must[T any](value T, err error) T {
	if err != nil {
		log.Fatal().Err(err).Msg("A fatal error has occured")
	}
	return value
}
