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
	"sort"
	"time"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/ftag"
	"github.com/siemens/wfx/generated/api"
	"github.com/siemens/wfx/generated/ent"
	entjob "github.com/siemens/wfx/generated/ent/job"
	"github.com/siemens/wfx/generated/ent/tag"
	"github.com/siemens/wfx/internal/errkind"
	"github.com/siemens/wfx/internal/workflow"
	"github.com/siemens/wfx/middleware/logging"
	"github.com/siemens/wfx/persistence"
)

// UpdateJob updates an existing job and its history.
func (db Database) UpdateJob(ctx context.Context, job *api.Job, request persistence.JobUpdate) (*api.Job, error) {
	log := logging.LoggerFromCtx(ctx).With().Str("id", job.ID).Logger()

	tx, err := db.client.Tx(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to start transaction")
		return nil, fault.Wrap(err)
	}

	updatedJob, err := doUpdateJob(ctx, tx, job, request)
	if err != nil {
		log.Error().Err(err).Msg("Rolling back transaction")
		if txErr := tx.Rollback(); txErr != nil {
			log.Error().Err(txErr).Msg("Rollback failed")
		}
		return nil, fault.Wrap(err)
	}

	err = tx.Commit()
	if err != nil {
		log.Error().Err(err).Msg("Failed to commit transaction")
		return nil, fault.Wrap(err)
	}

	log.Debug().
		Str("state", updatedJob.Status.State).
		Msg("Updated job")

	return updatedJob, nil
}

func doUpdateJob(ctx context.Context, tx *ent.Tx, job *api.Job, request persistence.JobUpdate) (*api.Job, error) {
	log := logging.LoggerFromCtx(ctx).With().Str("id", job.ID).Logger()

	oldMtime := time.Time(*job.Mtime)

	// Optimistic concurrency control: only update if mtime in the database still matches the mtime of the job we read
	// earlier. This prevents time-of-check-to-time-of-use races where two callers concurrently transition the same job
	// (for example, both moving from state A, but to different target states); without this guard, both would succeed
	// and the second would silently overwrite the first.
	//
	// Note: simply wrapping the read + validate + write in a single transaction is NOT sufficient to fix this race. At
	// the default isolation level (READ COMMITTED on PostgreSQL, REPEATABLE READ on MySQL InnoDB), two transactions can
	// both SELECT the same row and then both UPDATE it - the second waits for the first to commit and then clobbers its
	// result. SQLite serializes writers globally, but likewise just runs the second writer after the first. Preventing
	// the lost update therefore requires either pessimistic row locking (SELECT ... FOR UPDATE, not portable to SQLite
	// and forcing a transactional Storage API) or this optimistic mtime check, which is portable across all supported
	// backends and adds no contention.
	updater := tx.Job.UpdateOneID(job.ID).Where(entjob.MtimeEQ(oldMtime))

	if request.Status != nil {
		updater.SetStatus(*request.Status)
		if job.Workflow != nil {
			g := workflow.FindStateGroup(job.Workflow, request.Status.State)
			updater.SetGroup(g)
		}
	}
	if request.Definition != nil {
		updater.SetDefinition(*request.Definition)
	}

	allTags := make(map[string]any)
	if job.Tags != nil {
		for _, t := range *job.Tags {
			allTags[t] = nil
		}
	}
	{ // deal with tags
		if request.AddTags != nil && len(*request.AddTags) > 0 {
			// tags which we have to add to the job
			tagsToAdd := make([]string, 0, len(*request.AddTags))
			for _, name := range *request.AddTags {
				if _, found := allTags[name]; !found {
					tagsToAdd = append(tagsToAdd, name)
				}
			}

			// query already existing tags from the database
			var existingTags map[string]*ent.Tag
			{
				existing, err := tx.Tag.Query().Where(tag.NameIn(tagsToAdd...)).All(ctx)
				if err != nil {
					return nil, fault.Wrap(err)
				}
				existingTags = make(map[string]*ent.Tag, len(existing))
				for _, t := range existing {
					existingTags[t.Name] = t
				}
			}

			// tags which do not exist yet
			tagsToCreate := make([]*ent.TagCreate, 0)
			for _, name := range tagsToAdd {
				if t, found := existingTags[name]; found {
					// otherwise we lose them
					updater.AddTags(t)
				} else {
					// add it to our bulk creation query
					tagsToCreate = append(tagsToCreate, tx.Tag.Create().SetName(name))
				}
			}

			if len(tagsToCreate) > 0 {
				// create tags in db
				newTags, err := tx.Tag.CreateBulk(tagsToCreate...).Save(ctx)
				if err != nil {
					log.Error().Err(err).Msg("Failed to create new tags")
					return nil, fault.Wrap(err)
				}
				// add them to our update query
				updater.AddTags(newTags...)
			}

			// union set
			for _, t := range *request.AddTags {
				allTags[t] = nil
			}
		}

		if request.DelTags != nil && len(*request.DelTags) > 0 {
			tags, err := tx.Tag.Query().Where(tag.NameIn(*request.DelTags...)).All(ctx)
			if err != nil {
				return nil, fault.Wrap(err)
			}
			updater.RemoveTags(tags...)
			for _, t := range tags {
				delete(allTags, t.Name)
			}
		}
	}

	entity, err := updater.Save(ctx)
	if err != nil {
		// If no row matched, distinguish between "job no longer exists" and
		// "job was concurrently modified" so callers (and humans reading logs)
		// can tell why their update was rejected.
		if ent.IsNotFound(err) {
			exists, existsErr := tx.Job.Query().Where(entjob.IDEQ(job.ID)).Exist(ctx)
			if existsErr == nil && exists {
				log.Warn().Time("expectedMtime", oldMtime).Msg("Concurrent update detected; aborting")
				return nil, fault.Wrap(fmt.Errorf("status of job %s was concurrently modified", job.ID), ftag.With(errkind.TOCTOU))
			}
		}
		return nil, fault.Wrap(err)
	}

	{ // update history table
		history := tx.History.Create().
			SetJobID(job.ID).
			SetMtime(oldMtime)
		// if status changed, save old status
		if request.Status != nil {
			history.SetStatus(*job.Status)
		}
		if request.Definition != nil && job.Definition != nil {
			history.SetDefinition(job.Definition)
		}
		if _, err = history.Save(ctx); err != nil {
			return nil, fault.Wrap(err)
		}
	}

	updatedJob := convertJob(entity)

	// XXX: this feels like a bug in entgo, shouldn't be necessary to set Tags
	tags := make([]string, 0, len(allTags))
	for t := range allTags {
		tags = append(tags, t)
	}
	sort.Strings(tags)

	updatedJob.Tags = &tags

	return &updatedJob, nil
}
