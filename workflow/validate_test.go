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

	"github.com/stretchr/testify/assert"

	"github.com/siemens/wfx/generated/model"
	"github.com/siemens/wfx/workflow/dau"
)

var (
	state1 = "state1"
	state2 = "state2"
	state3 = "state3"
	state4 = "state4"

	eligibleClient = model.EligibleEnumCLIENT
	eligibleWfx    = model.EligibleEnumWFX

	groupOpen = "OPEN"

	name = "dummy"
)

func TestValidateWorkflow_NoStates(t *testing.T) {
	err := ValidateWorkflow(&model.Workflow{})
	assert.Error(t, err)
}

func TestValidateWorkflow_Dau(t *testing.T) {
	err := ValidateWorkflow(dau.PhasedWorkflow())
	assert.NoError(t, err)

	err = ValidateWorkflow(dau.DirectWorkflow())
	assert.NoError(t, err)
}

func TestValidateWorkflow_StateNamesMustBeUnique(t *testing.T) {
	eligible := model.EligibleEnumWFX
	m := model.Workflow{
		Name: name,
		States: []*model.State{
			{
				Name: state1,
			},
			{
				Name: state1,
			},
		},
		Transitions: []*model.Transition{
			{From: state1, To: state1, Eligible: eligible},
		},
	}
	err := ValidateWorkflow(&m)
	assert.Equal(t, "state1 state has already been created", err.Error())
}

func TestValidateWorkflow_TransitionsMustExist(t *testing.T) {
	m := model.Workflow{
		Name: name,
		States: []*model.State{
			{
				Name: state1,
			},
			{
				Name: state2,
			},
		},
		Transitions: []*model.Transition{
			{
				From:     state1,
				To:       state4,
				Eligible: eligibleClient,
			},
		},
	}
	err := ValidateWorkflow(&m)
	assert.Equal(t, "transition state1 -> state4 contains unknown state name", err.Error())
}

func TestValidateWorkflow_DuplicateTransition(t *testing.T) {
	m := model.Workflow{
		Name: name,
		States: []*model.State{
			{
				Name: state1,
			},
			{
				Name: state2,
			},
		},
		Transitions: []*model.Transition{
			{
				From:     state1,
				To:       state2,
				Eligible: eligibleClient,
			},
			{
				From:     state1,
				To:       state2,
				Eligible: eligibleClient,
			},
		},
	}
	err := ValidateWorkflow(&m)
	assert.ErrorContains(t, err, "duplicate transition: state1 -> state2 eligible: CLIENT")
}

func TestValidateWorkflow_ReachableFromInitial(t *testing.T) {
	m := model.Workflow{
		Name: name,
		States: []*model.State{
			{
				Name: state1,
			},
			{
				Name: state2,
			},
			{
				Name: state3,
			},
		},
		Transitions: []*model.Transition{
			{
				From:     state1,
				To:       state2,
				Eligible: eligibleClient,
			},
		},
	}
	err := ValidateWorkflow(&m)
	assert.Equal(t, "workflow must have exactly one INITIAL state", err.Error())
}

func TestValidateWorkflow_NoCycles(t *testing.T) {
	m := model.Workflow{
		Name: name,
		States: []*model.State{
			{
				Name: state1,
			},
			{
				Name: state2,
			},
			{
				Name: state3,
			},
		},
		Transitions: []*model.Transition{
			{
				From:     state1,
				To:       state2,
				Eligible: eligibleClient,
			},
			{
				From:     state2,
				To:       state1,
				Eligible: eligibleClient,
			},
			{
				From:     state1,
				To:       state3,
				Eligible: eligibleClient,
			},
		},
	}
	err := ValidateWorkflow(&m)
	assert.Equal(t, "workflow contains cycle from state1 to state2", err.Error())
}

func TestValidateWorkflow_UnambiguousWfxTransition(t *testing.T) {
	m := model.Workflow{
		Name: name,
		States: []*model.State{
			{
				Name: state1,
			},
			{
				Name: state2,
			},
			{
				Name: state3,
			},
		},
		Transitions: []*model.Transition{
			{
				From:     state1,
				To:       state2,
				Eligible: eligibleWfx,
				Action:   model.ActionEnumIMMEDIATE,
			},
			{
				From:     state1,
				To:       state3,
				Eligible: eligibleWfx,
				Action:   model.ActionEnumIMMEDIATE,
			},
		},
	}
	err := ValidateWorkflow(&m)
	assert.ErrorContains(t, err, "more than one immediate action from state state1")
}

func TestValidateWorkflow_ImmediateActionUnique(t *testing.T) {
	state3 := "state3"
	m := model.Workflow{
		Name: name,
		States: []*model.State{
			{
				Name: state1,
			},
			{
				Name: state2,
			},
			{
				Name: state3,
			},
		},
		Transitions: []*model.Transition{
			{
				From:     state1,
				To:       state2,
				Eligible: eligibleWfx,
				Action:   model.ActionEnumIMMEDIATE,
			},
			{
				From:     state1,
				To:       state3,
				Eligible: eligibleWfx,
				Action:   model.ActionEnumWAIT,
			},
		},
	}
	err := ValidateWorkflow(&m)
	assert.ErrorContains(t, err, "transition with source state1 contains impossible transition")
}

func TestValidateWorkflow_ImmediateActionUnique2(t *testing.T) {
	m := model.Workflow{
		Name: name,
		States: []*model.State{
			{
				Name: state1,
			},
			{
				Name: state2,
			},
			{
				Name: state3,
			},
		},
		Transitions: []*model.Transition{
			{
				From:     state1,
				To:       state1,
				Eligible: eligibleWfx,
				Action:   model.ActionEnumWAIT,
			},
			{
				From:     state1,
				To:       state2,
				Eligible: eligibleWfx,
				Action:   model.ActionEnumIMMEDIATE,
			},
			{
				From:     state2,
				To:       state3,
				Eligible: eligibleClient,
			},
		},
	}
	err := ValidateWorkflow(&m)
	assert.ErrorContains(t, err, "transition with source state1 contains impossible transition")
}

func TestValidateWorkflow_AllowMultipleTransitions(t *testing.T) {
	state3 := "state3"
	m := model.Workflow{
		Name: name,
		States: []*model.State{
			{
				Name: state1,
			},
			{
				Name: state2,
			},
			{
				Name: state3,
			},
		},
		Transitions: []*model.Transition{
			{
				From:     state1,
				To:       state2,
				Eligible: eligibleClient,
			},
			{
				From:     state1,
				To:       state2,
				Eligible: eligibleWfx,
			},
			{
				From:     state1,
				To:       state3,
				Eligible: eligibleWfx,
			},
		},
	}
	err := ValidateWorkflow(&m)
	assert.NoError(t, err)
}

func TestValidateWorkflow_GroupsNoOverlap(t *testing.T) {
	groupName2 := "CLOSED"
	m := model.Workflow{
		Name: name,
		States: []*model.State{
			{
				Name: state1,
			},
			{
				Name: state2,
			},
		},
		Transitions: []*model.Transition{
			{
				From:     state1,
				To:       state2,
				Eligible: eligibleClient,
			},
		},
		Groups: []*model.Group{
			{
				Name: groupOpen,
				States: []string{
					state1,
				},
			},
			{
				Name: groupName2,
				States: []string{
					state1,
				},
			},
		},
	}
	err := ValidateWorkflow(&m)
	assert.ErrorContains(t, err, "state state1 belongs to more than one group")
}

func TestValidateWorkflow_GroupNamesUnique(t *testing.T) {
	groupName2 := "OPEN"
	m := model.Workflow{
		Name: name,
		States: []*model.State{
			{
				Name: state1,
			},
			{
				Name: state2,
			},
		},
		Transitions: []*model.Transition{
			{
				From:     state1,
				To:       state2,
				Eligible: eligibleClient,
			},
		},
		Groups: []*model.Group{
			{
				Name: groupOpen,
				States: []string{
					state1,
				},
			},
			{
				Name: groupName2,
				States: []string{
					state2,
				},
			},
		},
	}
	err := ValidateWorkflow(&m)
	assert.ErrorContains(t, err, "group name OPEN used multiple times")
}
