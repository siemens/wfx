package server

/*
 * SPDX-FileCopyrightText: 2024 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/ftag"
	wfxAPI "github.com/siemens/wfx/api"
	"github.com/siemens/wfx/generated/api"
	"github.com/siemens/wfx/persistence"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var allOrientations = []string{"north", "south"}

func createServerForTesting(t *testing.T, orientation string, db persistence.Storage) api.StrictServerInterface {
	wfx := wfxAPI.NewWfxServer(db)
	wfx.Start()
	t.Cleanup(func() { wfx.Stop() })
	switch orientation {
	case "north":
		return NewNorthboundServer(wfx)
	case "south":
		return NewSouthboundServer(wfx)
	}
	panic("invalid orientation")
}

func TestGetJobsIDStatus_NotFound(t *testing.T) {
	jobID := "42"

	for _, orientation := range allOrientations {
		t.Run(orientation, func(t *testing.T) {
			dbMock := persistence.NewHealthyMockStorage(t)
			dbMock.EXPECT().GetJob(t.Context(), jobID, persistence.FetchParams{}).Return(nil, fault.Wrap(fmt.Errorf("job with id %s does not exist", jobID), ftag.With(ftag.NotFound)))

			server := createServerForTesting(t, orientation, dbMock)
			resp, err := server.GetJobsIdStatus(t.Context(), api.GetJobsIdStatusRequestObject{Id: jobID})
			assert.NoError(t, err)

			recorder := httptest.NewRecorder()
			_ = resp.VisitGetJobsIdStatusResponse(recorder)
			response := recorder.Result()
			assert.Equal(t, http.StatusNotFound, response.StatusCode)
		})
	}
}

func TestGetJobsIDStatus_InternalError(t *testing.T) {
	jobID := "42"
	for _, orientation := range allOrientations {
		t.Run(orientation, func(t *testing.T) {
			dbMock := persistence.NewHealthyMockStorage(t)
			dbMock.EXPECT().GetJob(t.Context(), jobID, persistence.FetchParams{}).Return(nil, errors.New("something went wrong"))

			server := createServerForTesting(t, orientation, dbMock)
			resp, err := server.GetJobsIdStatus(t.Context(), api.GetJobsIdStatusRequestObject{Id: jobID})
			assert.Error(t, err)
			assert.Nil(t, resp)
		})
	}
}

func TestPutJobsIDStatus_NotFound(t *testing.T) {
	jobID := "42"

	for _, orientation := range allOrientations {
		t.Run(orientation, func(t *testing.T) {
			dbMock := persistence.NewHealthyMockStorage(t)
			dbMock.EXPECT().GetJob(mock.AnythingOfType("*context.valueCtx"), jobID, persistence.FetchParams{}).Return(nil, fault.Wrap(fmt.Errorf("job with id %s does not exist", jobID), ftag.With(ftag.NotFound)))

			server := createServerForTesting(t, orientation, dbMock)
			resp, err := server.PutJobsIdStatus(t.Context(), api.PutJobsIdStatusRequestObject{
				Id:   jobID,
				Body: &api.JobStatus{},
			})
			assert.NoError(t, err)

			recorder := httptest.NewRecorder()
			_ = resp.VisitPutJobsIdStatusResponse(recorder)
			response := recorder.Result()

			assert.Equal(t, http.StatusNotFound, response.StatusCode)
		})
	}
}

func TestPutJobsIDStatus_InternalError(t *testing.T) {
	jobID := "42"

	for _, orientation := range allOrientations {
		t.Run(orientation, func(t *testing.T) {
			dbMock := persistence.NewHealthyMockStorage(t)
			dbMock.EXPECT().GetJob(mock.AnythingOfType("*context.valueCtx"), jobID, persistence.FetchParams{}).Return(nil, errors.New("something went wrong"))

			server := createServerForTesting(t, orientation, dbMock)
			resp, err := server.PutJobsIdStatus(t.Context(), api.PutJobsIdStatusRequestObject{Id: jobID})
			assert.Error(t, err)
			assert.Nil(t, resp)
		})
	}
}

func TestGetJobs_InternalError(t *testing.T) {
	for _, orientation := range allOrientations {
		t.Run(orientation, func(t *testing.T) {
			dbMock := persistence.NewHealthyMockStorage(t)
			dbMock.EXPECT().
				QueryJobs(t.Context(), persistence.FilterParams{}, persistence.SortParams{}, persistence.PaginationParams{Limit: 10}).
				Return(nil, errors.New("something went wrong"))

			server := createServerForTesting(t, orientation, dbMock)
			resp, err := server.GetJobs(t.Context(), api.GetJobsRequestObject{Params: api.GetJobsParams{}})
			assert.Error(t, err)
			assert.Nil(t, resp)
		})
	}
}

func TestGetJobsID_NotFound(t *testing.T) {
	jobID := "42"
	for _, orientation := range allOrientations {
		t.Run(orientation, func(t *testing.T) {
			dbMock := persistence.NewHealthyMockStorage(t)
			dbMock.EXPECT().GetJob(t.Context(), jobID, persistence.FetchParams{}).Return(nil, fault.Wrap(fmt.Errorf("job with id %s does not exist", jobID), ftag.With(ftag.NotFound)))

			server := createServerForTesting(t, orientation, dbMock)
			resp, err := server.GetJobsId(t.Context(), api.GetJobsIdRequestObject{Id: jobID})
			assert.NoError(t, err)

			recorder := httptest.NewRecorder()
			_ = resp.VisitGetJobsIdResponse(recorder)
			response := recorder.Result()

			assert.Equal(t, http.StatusNotFound, response.StatusCode)
		})
	}
}

func TestGetJobsID_InternalError(t *testing.T) {
	jobID := "42"
	for _, orientation := range allOrientations {
		t.Run(orientation, func(t *testing.T) {
			history := true
			dbMock := persistence.NewHealthyMockStorage(t)
			dbMock.EXPECT().GetJob(t.Context(), jobID, persistence.FetchParams{History: history}).Return(nil, errors.New("something went wrong"))

			server := createServerForTesting(t, orientation, dbMock)
			resp, err := server.GetJobsId(t.Context(), api.GetJobsIdRequestObject{Id: jobID, Params: api.GetJobsIdParams{ParamHistory: &history}})
			assert.Error(t, err)
			assert.Nil(t, resp)
		})
	}
}

func TestGetJobsIDDefinition_NotFound(t *testing.T) {
	jobID := "42"
	for _, orientation := range allOrientations {
		t.Run(orientation, func(t *testing.T) {
			dbMock := persistence.NewHealthyMockStorage(t)
			dbMock.EXPECT().GetJob(t.Context(), jobID, persistence.FetchParams{}).Return(nil, fault.Wrap(fmt.Errorf("job with id %s does not exist", jobID), ftag.With(ftag.NotFound)))

			server := createServerForTesting(t, orientation, dbMock)
			resp, err := server.GetJobsIdDefinition(t.Context(), api.GetJobsIdDefinitionRequestObject{Id: jobID})
			assert.NoError(t, err)

			recorder := httptest.NewRecorder()
			_ = resp.VisitGetJobsIdDefinitionResponse(recorder)
			response := recorder.Result()

			assert.Equal(t, http.StatusNotFound, response.StatusCode)
		})
	}
}

func TestGetJobsIDDefinition_InternalError(t *testing.T) {
	jobID := "42"
	for _, orientation := range allOrientations {
		t.Run(orientation, func(t *testing.T) {
			dbMock := persistence.NewHealthyMockStorage(t)
			dbMock.EXPECT().GetJob(t.Context(), jobID, persistence.FetchParams{}).Return(nil, errors.New("something went wrong"))

			server := createServerForTesting(t, orientation, dbMock)
			resp, err := server.GetJobsIdDefinition(t.Context(), api.GetJobsIdDefinitionRequestObject{Id: jobID})
			assert.Error(t, err)
			assert.Nil(t, resp)
		})
	}
}

func TestPutJobsIDDefinition_NotFound(t *testing.T) {
	jobID := "42"
	for _, orientation := range allOrientations {
		t.Run(orientation, func(t *testing.T) {
			dbMock := persistence.NewHealthyMockStorage(t)
			dbMock.EXPECT().GetJob(t.Context(), jobID, persistence.FetchParams{}).Return(nil, fault.Wrap(fmt.Errorf("job with id %s does not exist", jobID), ftag.With(ftag.NotFound)))

			server := createServerForTesting(t, orientation, dbMock)
			resp, err := server.PutJobsIdDefinition(t.Context(), api.PutJobsIdDefinitionRequestObject{Id: jobID})
			assert.NoError(t, err)

			recorder := httptest.NewRecorder()
			_ = resp.VisitPutJobsIdDefinitionResponse(recorder)
			response := recorder.Result()

			assert.Equal(t, http.StatusNotFound, response.StatusCode)
		})
	}
}

func TestPutJobsIDDefinition_InternalError(t *testing.T) {
	jobID := "42"
	for _, orientation := range allOrientations {
		t.Run(orientation, func(t *testing.T) {
			dbMock := persistence.NewHealthyMockStorage(t)
			dbMock.EXPECT().GetJob(t.Context(), jobID, persistence.FetchParams{}).Return(nil, errors.New("something went wrong"))

			server := createServerForTesting(t, orientation, dbMock)
			resp, err := server.PutJobsIdDefinition(t.Context(), api.PutJobsIdDefinitionRequestObject{Id: jobID})
			assert.Error(t, err)
			assert.Nil(t, resp)
		})
	}
}

func TestGetWorkflowsName_InternalError(t *testing.T) {
	workflow := "wfx.test.workflow"
	for _, orientation := range allOrientations {
		t.Run(orientation, func(t *testing.T) {
			dbMock := persistence.NewHealthyMockStorage(t)
			dbMock.EXPECT().GetWorkflow(t.Context(), workflow).Return(nil, errors.New("something went wrong"))

			server := createServerForTesting(t, orientation, dbMock)
			resp, err := server.GetWorkflowsName(t.Context(), api.GetWorkflowsNameRequestObject{Name: workflow})
			assert.Error(t, err)
			assert.Nil(t, resp)
		})
	}
}

func TestGetWorkflows_InternalError(t *testing.T) {
	for _, orientation := range allOrientations {
		t.Run(orientation, func(t *testing.T) {
			dbMock := persistence.NewHealthyMockStorage(t)
			dbMock.EXPECT().QueryWorkflows(t.Context(), persistence.SortParams{}, persistence.PaginationParams{Limit: 10}).Return(nil, errors.New("something went wrong"))

			server := createServerForTesting(t, orientation, dbMock)
			resp, err := server.GetWorkflows(t.Context(), api.GetWorkflowsRequestObject{Params: api.GetWorkflowsParams{}})
			assert.Error(t, err)
			assert.Nil(t, resp)
		})
	}
}

func TestGetWorkflows_Empty(t *testing.T) {
	for _, orientation := range allOrientations {
		t.Run(orientation, func(t *testing.T) {
			dbMock := persistence.NewHealthyMockStorage(t)
			dbMock.EXPECT().QueryWorkflows(t.Context(), persistence.SortParams{Desc: false}, persistence.PaginationParams{Limit: 10}).Return(&api.PaginatedWorkflowList{}, nil)

			server := createServerForTesting(t, orientation, dbMock)
			resp, err := server.GetWorkflows(t.Context(), api.GetWorkflowsRequestObject{Params: api.GetWorkflowsParams{}})
			assert.NoError(t, err)

			recorder := httptest.NewRecorder()
			_ = resp.VisitGetWorkflowsResponse(recorder)
			response := recorder.Result()

			assert.Equal(t, http.StatusOK, response.StatusCode)
		})
	}
}

func TestGetJobsIDTags_NotFound(t *testing.T) {
	jobID := "42"
	for _, orientation := range allOrientations {
		t.Run(orientation, func(t *testing.T) {
			dbMock := persistence.NewHealthyMockStorage(t)
			dbMock.EXPECT().GetJob(t.Context(), jobID, persistence.FetchParams{}).Return(nil, fault.Wrap(errors.New("not found"), ftag.With(ftag.NotFound)))

			server := createServerForTesting(t, orientation, dbMock)
			resp, err := server.GetJobsIdTags(t.Context(), api.GetJobsIdTagsRequestObject{Id: jobID})
			assert.NoError(t, err)

			recorder := httptest.NewRecorder()
			_ = resp.VisitGetJobsIdTagsResponse(recorder)
			response := recorder.Result()

			assert.Equal(t, http.StatusNotFound, response.StatusCode)
		})
	}
}

func TestGetJobsIDTags_InternalError(t *testing.T) {
	jobID := "42"
	for _, orientation := range allOrientations {
		t.Run(orientation, func(t *testing.T) {
			dbMock := persistence.NewHealthyMockStorage(t)
			dbMock.EXPECT().GetJob(t.Context(), jobID, persistence.FetchParams{}).Return(nil, fault.Wrap(errors.New("something went wrong"), ftag.With(ftag.Internal)))

			server := createServerForTesting(t, orientation, dbMock)
			resp, err := server.GetJobsIdTags(t.Context(), api.GetJobsIdTagsRequestObject{Id: jobID})
			assert.Error(t, err)
			assert.Nil(t, resp)
		})
	}
}

func TestHealth(t *testing.T) {
	for _, orientation := range allOrientations {
		t.Run(orientation, func(t *testing.T) {
			dbMock := persistence.NewHealthyMockStorage(t)
			server := createServerForTesting(t, orientation, dbMock)
			ok := false
			for range 10 {
				resp, err := server.GetHealth(t.Context(), api.GetHealthRequestObject{})
				require.NoError(t, err)
				recorder := httptest.NewRecorder()
				_ = resp.VisitGetHealthResponse(recorder)
				response := recorder.Result()
				if response.StatusCode == http.StatusOK {
					ok = true
					break
				}
				time.Sleep(200 * time.Millisecond)
			}
			assert.True(t, ok)
		})
	}
}

func TestVersion(t *testing.T) {
	dbMock := persistence.NewHealthyMockStorage(t)
	for _, orientation := range allOrientations {
		t.Run(orientation, func(t *testing.T) {
			server := createServerForTesting(t, orientation, dbMock)
			resp, err := server.GetVersion(t.Context(), api.GetVersionRequestObject{})
			require.NoError(t, err)

			recorder := httptest.NewRecorder()
			_ = resp.VisitGetVersionResponse(recorder)
			response := recorder.Result()
			assert.Equal(t, http.StatusOK, response.StatusCode)
		})
	}
}
