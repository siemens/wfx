package dau

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	_ "embed"

	"github.com/siemens/wfx/generated/api"
	"gopkg.in/yaml.v3"
)

//go:embed wfx.workflow.dau.direct.yml
var DirectYAML string

//go:embed wfx.workflow.dau.phased.yml
var PhasedYAML string

func DirectWorkflow() *api.Workflow {
	var result api.Workflow
	_ = yaml.Unmarshal([]byte(DirectYAML), &result)
	return &result
}

func PhasedWorkflow() *api.Workflow {
	var result api.Workflow
	_ = yaml.Unmarshal([]byte(PhasedYAML), &result)
	return &result
}
