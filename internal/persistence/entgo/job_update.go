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

	// Resolve (and if necessary create) any new tag rows BEFORE starting the job-update transaction. Tag rows are
	// global, idempotent, and protected by a UNIQUE(name) index. Two concurrent UpdateJob calls (on different jobs)
	// that both add the same brand-new tag would otherwise observe the row as missing, both attempt to insert, and one
	// would lose with a constraint violation that aborts the entire enclosing transaction (Postgres aborts the
	// transaction on any error; MySQL/SQLite just fail the statement). Doing it outside the main transaction lets us
	// swallow the benign duplicate-key error and re-resolve, without poisoning the job update.
	tagsByName, err := db.ensureTagsExist(ctx, request.AddTags)
	if err != nil {
		log.Error().Err(err).Msg("Failed to ensure tags exist")
		return nil, fault.Wrap(err)
	}

	tx, err := db.client.Tx(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to start transaction")
		return nil, fault.Wrap(err)
	}

	updatedJob, err := doUpdateJob(ctx, tx, job, request, tagsByName)
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

func doUpdateJob(ctx context.Context, tx *ent.Tx, job *api.Job, request persistence.JobUpdate, tagsByName map[string]*ent.Tag) (*api.Job, error) {
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
			// Tag rows were already resolved/created outside this tx (see
			// UpdateJob). Attach the ones the job does not have yet.
			for _, name := range *request.AddTags {
				if _, found := allTags[name]; found {
					continue
				}
				t, ok := tagsByName[name]
				if !ok {
					// Should not happen: ensureTagsExist guarantees coverage.
					return nil, fault.Wrap(fmt.Errorf("tag %q not pre-resolved", name))
				}
				updater.AddTags(t)
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

// ensureTagsExist resolves the given tag names to ent.Tag rows, creating any
// missing rows. It tolerates concurrent creators: if another transaction
// inserts the same name between our query and our insert, the resulting
// unique-constraint violation is swallowed and the row is re-fetched. The
// returned map is keyed by tag name and contains exactly the requested tags
// (empty if names was empty / nil).
func (db Database) ensureTagsExist(ctx context.Context, names *[]string) (map[string]*ent.Tag, error) {
	if names == nil || len(*names) == 0 {
		return map[string]*ent.Tag{}, nil
	}

	// Deduplicate.
	wanted := make(map[string]struct{}, len(*names))
	for _, n := range *names {
		wanted[n] = struct{}{}
	}
	wantedNames := make([]string, 0, len(wanted))
	for n := range wanted {
		wantedNames = append(wantedNames, n)
	}

	const maxAttempts = 3
	result := make(map[string]*ent.Tag, len(wanted))

	for range maxAttempts {
		existing, err := db.client.Tag.Query().Where(tag.NameIn(wantedNames...)).All(ctx)
		if err != nil {
			return nil, fault.Wrap(err)
		}
		for _, t := range existing {
			result[t.Name] = t
		}

		missing := make([]string, 0)
		for n := range wanted {
			if _, ok := result[n]; !ok {
				missing = append(missing, n)
			}
		}
		if len(missing) == 0 {
			return result, nil
		}

		// create each missing tag in its own statement so a duplicate from a concurrent writer doesn't poison the
		// others
		for _, n := range missing {
			t, err := db.client.Tag.Create().SetName(n).Save(ctx)
			if err == nil {
				result[n] = t
				continue
			}
			if ent.IsConstraintError(err) {
				// Lost the race - someone else just inserted this name. Fall
				// through to the next loop iteration which will re-query.
				continue
			}
			return nil, fault.Wrap(err)
		}
	}

	// re-query so we pick up any rows that were inserted by concurrent writers in the last iteration
	final, err := db.client.Tag.Query().Where(tag.NameIn(wantedNames...)).All(ctx)
	if err != nil {
		return nil, fault.Wrap(err)
	}
	for _, t := range final {
		result[t.Name] = t
	}
	for n := range wanted {
		if _, ok := result[n]; !ok {
			return nil, fault.Wrap(fmt.Errorf("failed to ensure tag %q exists after %d attempts", n, maxAttempts))
		}
	}
	return result, nil
}
