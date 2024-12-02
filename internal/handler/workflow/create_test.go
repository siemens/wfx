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

	"github.com/siemens/wfx/internal/persistence/entgo"
	"github.com/siemens/wfx/persistence"
	"github.com/siemens/wfx/workflow/dau"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateWorkflow(t *testing.T) {
	db := newInMemoryDB(t)
	wf, err := CreateWorkflow(context.Background(), db, dau.DirectWorkflow())
	assert.NoError(t, err)
	assert.Equal(t, "wfx.workflow.dau.direct", wf.Name)
}

func newInMemoryDB(t *testing.T) persistence.Storage {
	var db entgo.SQLite
	err := db.Initialize("file:wfx?mode=memory&cache=shared&_fk=1")
	require.NoError(t, err)
	t.Cleanup(db.Shutdown)
	t.Cleanup(func() {
		_ = db.DeleteWorkflow(context.Background(), "wfx.workflow.dau.direct")
	})
	return &db
}
