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
	"errors"
	"time"

	"github.com/Southclaws/fault"
	"github.com/rs/zerolog/log"
	"github.com/siemens/wfx/generated/ent"
	"github.com/siemens/wfx/generated/ent/predicate"
	"github.com/siemens/wfx/generated/ent/tag"
	"github.com/siemens/wfx/generated/ent/workflow"
	"github.com/siemens/wfx/generated/model"
	wfutil "github.com/siemens/wfx/internal/workflow"
	"github.com/siemens/wfx/middleware/logging"
)

// CreateJob persists a new job and sets the job ID field.
func (db Database) CreateJob(ctx context.Context, job *model.Job) (*model.Job, error) {
	log := logging.LoggerFromCtx(ctx)

	tx, err := db.client.Tx(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to start transaction")
		return nil, errors.New("failed to start transaction")
	}
	createdJob, err := createJobHelper(ctx, tx, job)
	if err != nil {
		log.Error().Err(err).Msg("Rolling back transaction")
		if txErr := tx.Rollback(); txErr != nil {
			log.Error().Err(txErr).Msg("Rollback failed")
		}
		return nil, fault.Wrap(err)
	}

	if err = tx.Commit(); err != nil {
		log.Error().Err(err).Msg("Failed to commit transaction")
		return nil, fault.Wrap(err)
	}

	return createdJob, nil
}

func createJobHelper(ctx context.Context, tx *ent.Tx, job *model.Job) (*model.Job, error) {
	log.Debug().
		Str("workflow", job.Workflow.Name).
		Str("state", job.Status.State).
		Strs("tags", job.Tags).
		Msg("Creating new job")

	wfEntity, err := tx.Workflow.
		Query().
		Where(workflow.Name(job.Workflow.Name)).
		Only(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to fetch workflow from database")
		return nil, fault.Wrap(err)
	}
	wf := convertWorkflow(wfEntity)
	group := wfutil.FindStateGroup(wf, job.Status.State)

	// start tags
	n := len(job.Tags)
	allTagIDs := make([]int, 0, n)
	if n > 0 {
		tagPreds := make([]predicate.Tag, n)
		for i, name := range job.Tags {
			tagPreds[i] = tag.Name(name)
		}

		existingTags := make(map[string]bool)
		{ // query existing tags
			tags, err := tx.Tag.Query().Where(tag.Or(tagPreds...)).All(ctx)
			if err != nil {
				return nil, fault.Wrap(err)
			}
			for _, t := range tags {
				existingTags[t.Name] = true
				allTagIDs = append(allTagIDs, t.ID)
			}
		}

		{ // create missing tags
			delta := len(job.Tags) - len(existingTags)
			if delta > 0 {
				missingTags := make([]*ent.TagCreate, 0, delta)
				for _, name := range job.Tags {
					if _, found := existingTags[name]; !found {
						missingTags = append(missingTags, tx.Tag.Create().SetName(name))
					}
				}
				newTags, err := tx.Tag.CreateBulk(missingTags...).Save(ctx)
				if err != nil {
					log.Error().Err(err).Msg("Failed to persist new tags")
					return nil, fault.Wrap(err)
				}
				for _, t := range newTags {
					log.Debug().Str("name", t.Name).Msg("Persisted new tag")
					allTagIDs = append(allTagIDs, t.ID)
				}
			}
		}
		log.Debug().Ints("allTagIDs", allTagIDs).Msg("Found all job tags")
	}
	// end tags

	builder := tx.Job.
		Create().
		SetClientID(job.ClientID).
		SetStatus(*job.Status).
		SetWorkflowID(wfEntity.ID).
		SetDefinition(job.Definition).
		AddTagIDs(allTagIDs...).
		SetGroup(group)

	if job.Stime != nil {
		builder.SetStime(time.Time(*job.Stime))
	}
	if job.Mtime != nil {
		builder.SetMtime(time.Time(*job.Mtime))
	}

	entity, err := builder.Save(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to persist job")
		return nil, fault.Wrap(err)
	}

	result := convertJob(entity)
	// tags and workflow are not fetched by entgo, so we have to add them manually
	result.Tags = job.Tags
	result.Workflow = wf
	return result, nil
}
