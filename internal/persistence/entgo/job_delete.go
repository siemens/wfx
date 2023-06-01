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
	"github.com/siemens/wfx/generated/ent"
)

func (db Database) DeleteJob(ctx context.Context, jobID string) error {
	err := db.client.Job.DeleteOneID(jobID).Exec(ctx)
	if ent.IsNotFound(err) {
		return fault.Wrap(fmt.Errorf("job with id %s was not found", jobID), ftag.With(ftag.NotFound))
	}
	return fault.Wrap(err)
}
