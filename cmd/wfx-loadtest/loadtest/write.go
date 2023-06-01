package loadtest

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
	"net/http"
	"sync/atomic"

	"github.com/Southclaws/fault"
	"github.com/rs/zerolog/log"
	"github.com/siemens/wfx/generated/model"
	vegeta "github.com/tsenart/vegeta/v12/lib"
)

func writeTargeter(tgt *vegeta.Target) error {
	if tgt == nil {
		log.Error().Msg("Target is nil")
		return vegeta.ErrNilTarget
	}
	// reset
	*tgt = vegeta.Target{}

	var status *model.JobStatus
	// pop
	queueMutex.RLock()
	n := len(queue)
	queueMutex.RUnlock()
	if n > 0 {
		queueMutex.Lock()
		// pop; we will add it back to the queue when we get the response from the server (not here)
		status, queue = queue[0], queue[1:]
		queueMutex.Unlock()
	}

	if status != nil {
		from := status.State
		var to string
		var eligible model.EligibleEnum
		// choose a suitable transition
		for _, t := range workflow.Transitions {
			if t.From == from {
				to = t.To
				eligible = t.Eligible
				break
			}
		}
		if to != "" {
			jobID := status.Context["id"].(string)

			var url string
			if eligible == model.EligibleEnumCLIENT {
				url = fmt.Sprintf("http://%s:%d/api/wfx/v1/jobs/%s/status", host, port, jobID)
			} else {
				url = fmt.Sprintf("http://%s:%d/api/wfx/v1/jobs/%s/status", mgmtHost, mgmtPort, jobID)
			}

			status.State = to
			// this field is read-only
			status.DefinitionHash = ""
			body, err := json.Marshal(status)
			if err != nil {
				log.Error().Err(err).Msg("Failed to marshal to JSON")
				return fault.Wrap(err)
			}
			log.Debug().
				Str("id", status.Context["id"].(string)).
				Str("from", from).
				Str("to", to).
				Str("url", url).
				RawJSON("body", body).
				Msg("Advancing job")
			*tgt = vegeta.Target{
				URL:    url,
				Method: http.MethodPut,
				Header: map[string][]string{
					"Accept":       {"application/json"},
					"Content-Type": {"application/json"},
				},
				Body: body,
			}
			return nil
		}
	}

	log.Debug().Msg("No job available, creating new one")
	jobReq := model.JobRequest{
		ClientID: "vegeta",
		Workflow: workflow.Name,
		Definition: map[string]any{
			"jobCounter": atomic.AddUint64(&jobCounter, 1),
		},
	}
	b, err := json.Marshal(&jobReq)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal job")
		return fault.Wrap(err)
	}
	*tgt = vegeta.Target{
		Method: http.MethodPost,
		URL:    fmt.Sprintf("http://%s:%d/api/wfx/v1/jobs", mgmtHost, mgmtPort),
		Body:   b,
		Header: map[string][]string{
			"Accept":       {"application/json"},
			"Content-Type": {"application/json"},
		},
	}
	return nil
}
