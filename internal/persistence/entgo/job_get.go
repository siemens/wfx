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
	"fmt"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/ftag"
	"github.com/go-openapi/strfmt"
	"github.com/siemens/wfx/generated/ent"
	"github.com/siemens/wfx/generated/ent/history"
	"github.com/siemens/wfx/generated/ent/job"
	"github.com/siemens/wfx/generated/ent/tag"
	"github.com/siemens/wfx/generated/model"
	"github.com/siemens/wfx/middleware/logging"
	"github.com/siemens/wfx/persistence"
)

func (db Database) GetJob(ctx context.Context, jobID string, fetchParams persistence.FetchParams) (*model.Job, error) {
	log := logging.LoggerFromCtx(ctx)
	contextLogger := log.With().Str("id", jobID).Logger()
	contextLogger.Debug().Msg("Fetching job")

	builder := db.client.Job.
		Query().Where(job.ID(jobID)).
		WithWorkflow().
		WithTags(func(q *ent.TagQuery) {
			q.Order(ent.Asc(tag.FieldName))
		})

	if fetchParams.History {
		contextLogger.Debug().Msg("Fetching history")
		builder.WithHistory(func(query *ent.HistoryQuery) {
			query.
				Limit(8192).
				Order(ent.Desc(history.FieldID))
		})
	}

	job, err := builder.Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			contextLogger.Debug().Msg("Job not found")
			return nil, fault.Wrap(fmt.Errorf("job with id %s does not exist", jobID), ftag.With(ftag.NotFound))
		}
		contextLogger.Error().Err(err).Msg("Failed to fetch job")
		return nil, fault.Wrap(err)
	}
	contextLogger.Debug().Msg("Fetched job")
	return convertJob(job), nil
}

func convertJob(entity *ent.Job) *model.Job {
	var wf *model.Workflow
	if entity.Edges.Workflow != nil {
		wf = convertWorkflow(entity.Edges.Workflow)
	}

	job := model.Job{
		ID:         entity.ID,
		ClientID:   entity.ClientID,
		Definition: entity.Definition,
		Stime:      strfmt.DateTime(entity.Stime),
		Mtime:      strfmt.DateTime(entity.Mtime),
		Status:     &entity.Status,
		Tags:       convertTags(entity.Edges.Tags),
		Workflow:   wf,
	}

	n := len(entity.Edges.History)
	if n > 0 {
		job.History = make([]*model.History, n)
		for i, entity := range entity.Edges.History {
			hist := convertHistory(entity)
			job.History[i] = &hist
		}
	}
	return &job
}

func convertHistory(entity *ent.History) model.History {
	return model.History{
		Mtime:  strfmt.DateTime(entity.Mtime),
		Status: &entity.Status,
	}
}

func convertTags(tags []*ent.Tag) []string {
	result := make([]string, len(tags))
	for i, t := range tags {
		result[i] = t.Name
	}
	return result
}
