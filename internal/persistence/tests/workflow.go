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
	"encoding/json"
	"testing"

	"github.com/Southclaws/fault/ftag"
	"github.com/siemens/wfx/persistence"
	"github.com/siemens/wfx/workflow/dau"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCRDWorkflow(t *testing.T, db persistence.Storage) {
	// create, read, delete; no update
	workflow := dau.DirectWorkflow()
	name := "TestCRDWorkflow"
	workflow.Name = name

	{
		t.Log("Creating new workflow")
		_, err := db.CreateWorkflow(context.Background(), workflow)
		require.NoError(t, err)
	}

	b, _ := json.Marshal(workflow)
	expectedJSON := string(b)

	{
		t.Log("Getting existing workflow")
		actual, err := db.GetWorkflow(context.Background(), workflow.Name)
		assert.NoError(t, err)
		b, _ := json.Marshal(actual)
		assert.JSONEq(t, expectedJSON, string(b))
	}

	{
		t.Log("Deleting existing workflow")
		err := db.DeleteWorkflow(context.Background(), workflow.Name)
		assert.NoError(t, err)
	}

	{
		t.Log("Deleting non-existing workflow")
		err := db.DeleteWorkflow(context.Background(), defaultClientID)
		assert.Equal(t, ftag.NotFound, ftag.Get(err))
	}

	{
		t.Log("Getting non-existing workflow")
		actual, err := db.GetWorkflow(context.Background(), workflow.Name)
		assert.Nil(t, actual)
		assert.Equal(t, ftag.NotFound, ftag.Get(err))
	}
}

func TestQueryWorkflows(t *testing.T, db persistence.Storage) {
	_, err := db.CreateWorkflow(context.Background(), dau.DirectWorkflow())
	require.NoError(t, err)

	allCount := 1
	result, err := db.QueryWorkflows(context.Background(), persistence.SortParams{}, persistence.PaginationParams{Offset: 0, Limit: int32(allCount)})
	require.NoError(t, err)
	assert.Len(t, result.Content, allCount)
}

func TestWorkflowsPagination(t *testing.T, db persistence.Storage) {
	_, err := db.CreateWorkflow(context.Background(), dau.DirectWorkflow())
	require.NoError(t, err)
	_, err = db.CreateWorkflow(context.Background(), dau.PhasedWorkflow())
	require.NoError(t, err)

	result, err := db.QueryWorkflows(context.Background(), persistence.SortParams{}, persistence.PaginationParams{Offset: 0, Limit: 1})
	assert.NoError(t, err)
	assert.Len(t, result.Content, 1)

	result2, err := db.QueryWorkflows(context.Background(), persistence.SortParams{}, persistence.PaginationParams{Offset: 1, Limit: 1})
	assert.NoError(t, err)
	assert.Len(t, result2.Content, 1)
	assert.NotEqual(t, result.Content[0].Name, result2.Content[0].Name)

	result3, err := db.QueryWorkflows(context.Background(), persistence.SortParams{}, persistence.PaginationParams{Offset: 2, Limit: 1})
	assert.NoError(t, err)
	assert.Len(t, result3.Content, 0)
}

func TestQueryWorkflowsSort(t *testing.T, db persistence.Storage) {
	_, _ = db.CreateWorkflow(context.Background(), dau.DirectWorkflow())
	_, _ = db.CreateWorkflow(context.Background(), dau.PhasedWorkflow())

	allCount := 2
	t.Run("asc", func(t *testing.T) {
		result, err := db.QueryWorkflows(context.Background(), persistence.SortParams{Desc: false}, persistence.PaginationParams{Offset: 0, Limit: int32(allCount)})
		require.NoError(t, err)
		assert.Len(t, result.Content, allCount)

		keys := make([]string, 0, len(result.Content))
		for _, workflow := range result.Content {
			keys = append(keys, workflow.Name)
		}
		assert.IsIncreasing(t, keys)
	})

	t.Run("desc", func(t *testing.T) {
		result, err := db.QueryWorkflows(context.Background(), persistence.SortParams{Desc: true}, persistence.PaginationParams{Offset: 0, Limit: int32(allCount)})
		require.NoError(t, err)
		assert.Len(t, result.Content, allCount)

		keys := make([]string, 0, len(result.Content))
		for _, workflow := range result.Content {
			keys = append(keys, workflow.Name)
		}
		assert.IsDecreasing(t, keys)
	})
}
