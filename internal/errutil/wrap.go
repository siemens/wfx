package errutil

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import "github.com/Southclaws/fault"

// Wrap2 wraps the provided error using fault.Wrap and returns
// the original value along with the wrapped error.
func Wrap2[T any](value T, err error) (T, error) {
	return value, fault.Wrap(err)
}
