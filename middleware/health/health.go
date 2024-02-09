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

type MW struct {
	checker health.Checker
	health  http.Handler
}

func NewHealthMiddleware(storage persistence.Storage) MW {
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
		health.WithStatusListener(statusListener))

	return MW{
		checker: checker,
		health:  health.NewHandler(checker),
	}
}

func (mw MW) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/health") {
			mw.health.ServeHTTP(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (mw MW) Shutdown() {
	mw.checker.Stop()
}

func statusListener(_ context.Context, state health.CheckerState) {
	logFn := log.Warn
	switch state.Status {
	case health.StatusDown:
		logFn = log.Error
	case health.StatusUp:
		logFn = log.Info
	case health.StatusUnknown:
		logFn = log.Warn
	}

	childLog := logFn()
	for k, v := range state.CheckState {
		childLog.Str(k, string(v.Status))
	}

	childLog.Str("overall", string(state.Status)).Msg("Health status changed")
}
