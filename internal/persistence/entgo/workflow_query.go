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

	"github.com/Southclaws/fault"
	"github.com/siemens/wfx/generated/ent"
	"github.com/siemens/wfx/generated/ent/workflow"
	"github.com/siemens/wfx/generated/model"
	"github.com/siemens/wfx/middleware/logging"
	"github.com/siemens/wfx/persistence"
)

// QueryWorkflows returns multiple workflows (paginated).
func (db Database) QueryWorkflows(ctx context.Context, paginationParams persistence.PaginationParams) (*model.PaginatedWorkflowList, error) {
	log := logging.LoggerFromCtx(ctx)
	builder := db.client.Workflow.
		Query()

	// need to clone builder because it is unusable after we call `All`
	counter := builder.Clone()

	workflows, err := builder.
		Limit(int(paginationParams.Limit)).
		Offset(int(paginationParams.Offset)).
		Order(ent.Asc(workflow.FieldName)).
		All(ctx)
	if err != nil {
		return nil, fault.Wrap(err)
	}

	total, err := counter.Count(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to count workflows")
		return nil, fault.Wrap(err)
	}

	result := model.PaginatedWorkflowList{
		Pagination: &model.PaginatedWorkflowListPagination{
			Total:  int64(total),
			Offset: paginationParams.Offset,
			Limit:  paginationParams.Limit,
		},
		Content: make([]*model.Workflow, 0, len(workflows)),
	}
	for _, wf := range workflows {
		result.Content = append(result.Content, convertWorkflow(wf))
	}

	log.Debug().
		Int("total", total).
		Int32("limit", paginationParams.Limit).
		Int64("offset", paginationParams.Offset).
		Msg("Fetched workflows")

	return &result, nil
}
