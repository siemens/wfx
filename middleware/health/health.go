package health

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/Southclaws/fault"
	"github.com/alexliesenfeld/health"
	"github.com/rs/zerolog/log"

	"github.com/siemens/wfx/persistence"
)

func NewHealthMiddleware(storage persistence.Storage, next http.Handler) http.Handler {
	log.Debug().Msg("Adding health middleware")

	checker := health.NewChecker(
		health.WithTimeout(10*time.Second),
		health.WithPeriodicCheck(30*time.Second, 3*time.Second, health.Check{
			Name: "persistence",
			Check: func(ctx context.Context) error {
				if storage == nil {
					return nil
				}
				return fault.Wrap(storage.CheckHealth(ctx))
			},
		}),
		health.WithStatusListener(statusListener),
	)

	handler := health.NewHandler(checker)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/health") {
			handler(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func statusListener(_ context.Context, state health.CheckerState) {
	childLog := log.Warn()
	switch state.Status {
	case health.StatusDown:
		childLog = log.Error()
	case health.StatusUp:
		childLog = log.Info()
	case health.StatusUnknown:
		childLog = log.Warn()
	}

	for k, v := range state.CheckState {
		childLog.Str(k, string(v.Status))
	}

	childLog.Str("overall", string(state.Status)).Msg("Health status changed")
}
