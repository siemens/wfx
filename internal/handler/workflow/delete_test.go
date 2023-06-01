package workflow

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
	"github.com/siemens/wfx/workflow/dau"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeleteWorkflow(t *testing.T) {
	db := newInMemoryDB(t)
	wf, err := CreateWorkflow(context.Background(), db, dau.DirectWorkflow())
	require.NoError(t, err)
	err = DeleteWorkflow(context.Background(), db, wf.Name)
	assert.NoError(t, err)
}

func TestDeleteWorkflow_NotFound(t *testing.T) {
	db := newInMemoryDB(t)
	err := DeleteWorkflow(context.Background(), db, "does not exist")
	require.NotNil(t, err)
	assert.Equal(t, ftag.NotFound, ftag.Get(err))
}
