package workflow

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"testing"

	"github.com/siemens/wfx/generated/model"
	"github.com/siemens/wfx/workflow/dau"
	"github.com/stretchr/testify/assert"
)

func TestFindStateGroup(t *testing.T) {
	workflow := dau.PhasedWorkflow()
	{
		group := FindStateGroup(workflow, "DOWNLOAD")
		assert.Equal(t, "OPEN", group)
	}
	{
		group := FindStateGroup(workflow, "FOO")
		assert.Equal(t, "", group)
	}
}

func TestFollowTransitions(t *testing.T) {
	a := "a"
	b := "b"
	c := "c"
	d := "d"

	eligibleWfx := model.EligibleEnumWFX

	transitions := []*model.Transition{
		{From: a, To: b, Eligible: eligibleWfx, Action: model.ActionEnumIMMEDIATE},
		{From: b, To: c, Eligible: eligibleWfx, Action: model.ActionEnumIMMEDIATE},
		{From: c, To: d, Eligible: eligibleWfx, Action: model.ActionEnumIMMEDIATE},
	}

	actual := FollowImmediateTransitions(&model.Workflow{Transitions: transitions}, "a")
	assert.Equal(t, d, actual, "should warp from a to d")
}

func TestFindInitialState(t *testing.T) {
	wf := dau.DirectWorkflow()
	initial := FindInitialState(wf)
	assert.Equal(t, "INSTALL", *initial)
}

func TestFindFinalStates(t *testing.T) {
	wf := dau.DirectWorkflow()
	finaleStates := FindFinalStates(wf)
	assert.Equal(t, []string{"ACTIVATED", "TERMINATED"}, finaleStates)
	assert.IsIncreasing(t, finaleStates)
}
