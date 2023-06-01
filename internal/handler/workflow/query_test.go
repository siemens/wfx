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

	"github.com/siemens/wfx/persistence"
	"github.com/siemens/wfx/workflow/dau"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueryWorkflow(t *testing.T) {
	db := newInMemoryDB(t)
	wf, err := CreateWorkflow(context.Background(), db, dau.DirectWorkflow())
	require.NoError(t, err)

	list, err := QueryWorkflows(context.Background(), db, persistence.PaginationParams{Limit: 10})
	assert.NoError(t, err)
	assert.Len(t, list.Content, 1)
	assert.Equal(t, wf.Name, list.Content[0].Name)
}
