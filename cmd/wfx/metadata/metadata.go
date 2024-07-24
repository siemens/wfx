package metadata

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"time"

	"github.com/Southclaws/fault"
)

var (
	// values provided by linker
	Version    = "dev"
	Commit     = "unknown"
	Date       = "1970-01-01T00:00:00+00:00"
	APIVersion = "v1"
)

func BuildDate() (time.Time, error) {
	buildDate, err := time.Parse("2006-01-02T15:04:05-07:00", Date)
	return buildDate, fault.Wrap(err)
}
