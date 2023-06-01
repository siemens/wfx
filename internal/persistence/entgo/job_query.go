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

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqljson"
	"github.com/Southclaws/fault"

	"github.com/siemens/wfx/generated/ent"
	"github.com/siemens/wfx/generated/ent/job"
	"github.com/siemens/wfx/generated/ent/tag"
	"github.com/siemens/wfx/generated/ent/workflow"
	"github.com/siemens/wfx/generated/model"
	"github.com/siemens/wfx/middleware/logging"
	"github.com/siemens/wfx/persistence"
)

// QueryJobs returns the jobs matching filterParams.
func (db Database) QueryJobs(ctx context.Context,
	filterParams persistence.FilterParams,
	sortParams persistence.SortParams,
	paginationParams persistence.PaginationParams,
) (*model.PaginatedJobList, error) {
	log := logging.LoggerFromCtx(ctx)
	builder := db.client.Job.Query().WithWorkflow().WithTags(func(q *ent.TagQuery) {
		q.Order(ent.Asc(tag.FieldName))
	})

	if filterParams.ClientID != nil && *filterParams.ClientID != "" {
		log.Debug().Str("clientID", *filterParams.ClientID).Msg("Adding filter")
		builder.Where(job.ClientID(*filterParams.ClientID))
	}
	if filterParams.State != nil && *filterParams.State != "" {
		log.Debug().Str("state", *filterParams.State).Msg("Adding filter")
		builder.Where(func(s *sql.Selector) {
			s.Where(sqljson.ValueEQ("status", filterParams.State, sqljson.Path("state")))
		})
	}
	if filterParams.Group != nil {
		log.Debug().Strs("groups", filterParams.Group).Msg("Adding filter")
		builder.Where(job.GroupIn(filterParams.Group...))
	}
	if filterParams.Workflow != nil && *filterParams.Workflow != "" {
		log.Debug().Str("workflow", *filterParams.Workflow).Msg("Adding filter")
		builder.Where(job.HasWorkflowWith(workflow.Name(*filterParams.Workflow)))
	}

	for _, t := range filterParams.Tags {
		builder.Where(job.HasTagsWith(tag.Name(t)))
	}

	// deterministic ordering
	if sortParams.Desc {
		builder.Order(ent.Desc(job.FieldStime))
	} else {
		builder.Order(ent.Asc(job.FieldStime))
	}

	// need to clone builder because it is unusable after we call `All`
	counter := builder.Clone()

	jobs, err := builder.
		Limit(int(paginationParams.Limit)).
		Offset(int(paginationParams.Offset)).
		All(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to fetch jobs")
		return nil, fault.Wrap(err)
	}

	total, err := counter.Count(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to count jobs")
		return nil, fault.Wrap(err)
	}

	result := model.PaginatedJobList{
		Pagination: &model.PaginatedJobListPagination{
			Total:  int64(total),
			Limit:  paginationParams.Limit,
			Offset: paginationParams.Offset,
		},
		Content: make([]*model.Job, 0, len(jobs)),
	}
	for _, entity := range jobs {
		result.Content = append(result.Content, convertJob(entity))
	}

	log.Debug().
		Int("total", total).
		Int32("limit", paginationParams.Limit).
		Int64("offset", paginationParams.Offset).
		Msg("Fetched jobs")

	return &result, nil
}
