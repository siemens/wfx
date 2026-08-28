package entgo

/*
 * SPDX-FileCopyrightText: 2026 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import (
	"context"

	"entgo.io/ent"
	generatedEnt "github.com/siemens/wfx/generated/ent"
	"github.com/siemens/wfx/generated/ent/job"
	"github.com/siemens/wfx/persistence"
)

func addClientIDInterceptor(client *generatedEnt.Client) {
	client.Job.Intercept(ent.TraverseFunc(func(ctx context.Context, query ent.Query) error {
		clientID, ok := persistence.ClientIDFromContext(ctx)
		if !ok {
			return nil
		}
		query.(*generatedEnt.JobQuery).Where(job.ClientID(clientID))
		return nil
	}))
	client.Job.Use(func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, mutation ent.Mutation) (ent.Value, error) {
			clientID, ok := persistence.ClientIDFromContext(ctx)
			if ok && mutation.Op().Is(ent.OpUpdate|ent.OpUpdateOne|ent.OpDelete|ent.OpDeleteOne) {
				mutation.(*generatedEnt.JobMutation).Where(job.ClientID(clientID))
			}
			return next.Mutate(ctx, mutation)
		})
	})
}
