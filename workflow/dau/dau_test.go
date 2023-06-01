package dau

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"testing"

	"github.com/siemens/wfx/workflow"
	"github.com/stretchr/testify/assert"
)

func TestDirectWorkflow(t *testing.T) {
	wf := DirectWorkflow()
	assert.NotNil(t, wf)
	err := workflow.ValidateWorkflow(wf)
	assert.NoError(t, err)
}

func TestPhasedWorkflow(t *testing.T) {
	wf := PhasedWorkflow()
	assert.NotNil(t, wf)
	err := workflow.ValidateWorkflow(wf)
	assert.NoError(t, err)
}
