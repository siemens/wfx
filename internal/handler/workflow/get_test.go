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
	"errors"
	"testing"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/ftag"
	"github.com/siemens/wfx/persistence"
	"github.com/siemens/wfx/workflow/dau"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetWorkflow(t *testing.T) {
	db := newInMemoryDB(t)
	wf, err := CreateWorkflow(context.Background(), db, dau.DirectWorkflow())
	require.NoError(t, err)

	wf2, err := GetWorkflow(context.Background(), db, wf.Name)
	assert.NoError(t, err)
	assert.Equal(t, wf.Name, wf2.Name)
}

func TestGetWorkflow_NotFound(t *testing.T) {
	ctx := context.Background()

	dbMock := persistence.NewHealthyMockStorage(t)
	dbMock.EXPECT().GetWorkflow(ctx, "foo").Return(nil, fault.Wrap(errors.New("Not found"), ftag.With(ftag.NotFound)))

	wf, err := GetWorkflow(ctx, dbMock, "foo")
	assert.Nil(t, wf)
	assert.NotNil(t, err)
}

func TestGetWorkflow_Internal(t *testing.T) {
	ctx := context.Background()

	dbMock := persistence.NewHealthyMockStorage(t)
	dbMock.EXPECT().GetWorkflow(ctx, "foo").Return(nil, fault.Wrap(errors.New("Not found"), ftag.With(ftag.Internal)))

	wf, err := GetWorkflow(ctx, dbMock, "foo")
	assert.Nil(t, wf)
	assert.NotNil(t, err)
}
