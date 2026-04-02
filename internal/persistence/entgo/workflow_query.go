package entgo

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"context"
	"time"

	"github.com/Southclaws/fault"
	"github.com/siemens/wfx/generated/api"
	"github.com/siemens/wfx/generated/ent"
	"github.com/siemens/wfx/generated/ent/workflow"
	"github.com/siemens/wfx/middleware/logging"
	"github.com/siemens/wfx/persistence"
)

// QueryWorkflows returns multiple workflows (paginated).
func (db Database) QueryWorkflows(ctx context.Context, sortParams persistence.SortParams, paginationParams persistence.PaginationParams) (*api.PaginatedWorkflowList, error) {
	log := logging.LoggerFromCtx(ctx)
	builder := db.client.Workflow.
		Query()

	// need to clone builder because it is unusable after we call `All`
	counter := builder.Clone()

	builder.
		Limit(int(paginationParams.Limit)).
		Offset(int(paginationParams.Offset))

	// deterministic ordering
	if sortParams.Desc {
		log.Debug().Msg("Sorting workflows in descending order")
		builder.Order(ent.Desc(workflow.FieldName))
	} else {
		log.Debug().Msg("Sorting workflows in ascending order")
		builder.Order(ent.Asc(workflow.FieldName))
	}

	workflows, err := builder.All(ctx)
	if err != nil {
		return nil, fault.Wrap(err)
	}

	var result api.PaginatedWorkflowList

	if paginationParams.ComputeTotal {
		start := time.Now()
		count, err := counter.Count(ctx)
		if err != nil {
			log.Error().Err(err).Msg("Failed to count workflows")
			return nil, fault.Wrap(err)
		}
		duration := time.Since(start)
		log.Debug().Dur("duration", duration).Int("count", count).Msg("Computed total number of workflows")

		result.Pagination = &api.Pagination{
			Total:  int64(count),
			Offset: paginationParams.Offset,
			Limit:  paginationParams.Limit,
		}
	}

	result.Content = make([]api.Workflow, 0, len(workflows))
	for _, wf := range workflows {
		result.Content = append(result.Content, convertWorkflow(wf))
	}
	return &result, nil
}
