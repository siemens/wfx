package api

/*
 * SPDX-FileCopyrightText: 2024 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"context"
	"testing"

	"github.com/alexliesenfeld/health"
)

func TestStatusListener(*testing.T) {
	healthStatusListener(context.Background(), health.CheckerState{Status: health.StatusUp})
	healthStatusListener(context.Background(), health.CheckerState{Status: health.StatusDown})
	healthStatusListener(context.Background(), health.CheckerState{Status: health.StatusUnknown})
	healthStatusListener(context.Background(), health.CheckerState{
		Status: health.StatusUp,
		CheckState: map[string]health.CheckState{
			"db": {Status: health.StatusUp},
		},
	})
}
