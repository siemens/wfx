package job

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"context"
	"testing"

	"github.com/Southclaws/fault/ftag"
	"github.com/stretchr/testify/assert"
)

func TestDeleteJob(t *testing.T) {
	db := newInMemoryDB(t)
	createDirectWorkflow(t, db)

	tmpJob := newValidJob("abc", "INSTALLING")
	job, err := db.CreateJob(context.Background(), &tmpJob)
	assert.NoError(t, err)

	err = DeleteJob(context.Background(), db, job.ID)
	assert.NoError(t, err)
}

func TestDeleteJob_NotFound(t *testing.T) {
	db := newInMemoryDB(t)
	err := DeleteJob(context.Background(), db, "42")
	assert.Equal(t, ftag.NotFound, ftag.Get(err))
}
