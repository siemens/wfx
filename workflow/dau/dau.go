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

	"github.com/siemens/wfx/generated/model"
	"gopkg.in/yaml.v3"
)

//go:embed wfx.workflow.dau.direct.yml
var DirectYAML string

//go:embed wfx.workflow.dau.phased.yml
var PhasedYAML string

func DirectWorkflow() *model.Workflow {
	var result model.Workflow
	_ = yaml.Unmarshal([]byte(DirectYAML), &result)
	return &result
}

func PhasedWorkflow() *model.Workflow {
	var result model.Workflow
	_ = yaml.Unmarshal([]byte(PhasedYAML), &result)
	return &result
}
