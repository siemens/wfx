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
