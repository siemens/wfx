//go:build testing

package tests

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 */

var AllTests = []PersistenceTest{
	TestCRDWorkflow,
	TestDeleteJob,
	TestDeleteJobNotFound,
	TestGetJob,
	TestGetJobMaxHistorySize,
	TestGetJobWithHistory,
	TestGetJobsSorted,
	TestJobAddTags,
	TestJobAddTagsOverlap,
	TestJobDeleteTags,
	TestJobDeleteTagsNonExisting,
	TestJobReuseExistingTags,
	TestJobsPagination,
	TestQueryJobsFilter,
	TestQueryWorkflows,
	TestUpdateJobDefinition,
	TestUpdateJobStatus,
	TestUpdateJobStatusNonExisting,
	TestWorkflowsPagination,
	TestQueryWorkflowsSort,
}
