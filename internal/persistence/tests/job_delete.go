//go:build testing

package tests

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
	"github.com/siemens/wfx/persistence"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeleteJob(t *testing.T, db persistence.Storage) {
	tmpJob := newValidJob(defaultClientID)
	_, err := db.CreateWorkflow(context.Background(), tmpJob.Workflow)
	require.NoError(t, err)

	job, err := db.CreateJob(context.Background(), tmpJob)
	assert.NoError(t, err)

	err = db.DeleteJob(context.Background(), job.ID)
	assert.NoError(t, err)
}

func TestDeleteJobNotFound(t *testing.T, db persistence.Storage) {
	err := db.DeleteJob(context.Background(), "9999")
	assert.Equal(t, ftag.NotFound, ftag.Get(err))
}
