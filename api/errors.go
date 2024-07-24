package api

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import "github.com/siemens/wfx/generated/api"

/* these are defined in the swagger yaml */

var InvalidRequest = api.Error{
	Code:    "wfx.invalidRequest",
	Message: "The request was invalid and/or could not be completed by the storage",
	Logref:  "96a37ea1f7d205ffbfa12334c6812727",
}

var JobNotFound = api.Error{
	Code:    "wfx.jobNotFound",
	Logref:  "11cc67762090e15b79a1387eca65ba65",
	Message: "Job ID was not found",
}

var WorkflowNotFound = api.Error{
	Code:    "wfx.workflowNotFound",
	Logref:  "c452719774086b6e803bb8f6ecea9899",
	Message: "Workflow not found for given name",
}

var WorkflowNotUnique = api.Error{
	Code:    "wfx.workflowNotUnique",
	Logref:  "e1ee1f2aea859b9dd34579610e386da6",
	Message: "Workflow with name already exists",
}

var WorkflowInvalid = api.Error{
	Code:    "wfx.workflowInvalid",
	Logref:  "18f57adc70dd79c7fb4f1246be8a6e04",
	Message: "Workflow validation failed",
}
