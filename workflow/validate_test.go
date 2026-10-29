package workflow

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/siemens/wfx/generated/api"
	"github.com/siemens/wfx/workflow/dau"
)

var (
	state1 = "state1"
	state2 = "state2"
	state3 = "state3"
	state4 = "state4"

	eligibleClient = api.CLIENT
	eligibleWfx    = api.WFX

	groupOpen = "OPEN"

	name = "dummy"
)

func TestValidateWorkflow_NoStates(t *testing.T) {
	err := ValidateWorkflow(&api.Workflow{})
	assert.Error(t, err)
}

func TestValidateWorkflow_Dau(t *testing.T) {
	err := ValidateWorkflow(dau.PhasedWorkflow())
	assert.NoError(t, err)

	err = ValidateWorkflow(dau.DirectWorkflow())
	assert.NoError(t, err)
}

func TestValidateWorkflow_StateNamesMustBeUnique(t *testing.T) {
	eligible := api.WFX
	m := api.Workflow{
		Name: name,
		States: []api.State{
			{
				Name: state1,
			},
			{
				Name: state1,
			},
		},
		Transitions: []api.Transition{
			{From: state1, To: state1, Eligible: eligible},
		},
	}
	err := ValidateWorkflow(&m)
	assert.Equal(t, "state1 state has already been created", err.Error())
}

func TestValidateWorkflow_TransitionsMustExist(t *testing.T) {
	m := api.Workflow{
		Name: name,
		States: []api.State{
			{
				Name: state1,
			},
			{
				Name: state2,
			},
		},
		Transitions: []api.Transition{
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
	m := api.Workflow{
		Name: name,
		States: []api.State{
			{
				Name: state1,
			},
			{
				Name: state2,
			},
		},
		Transitions: []api.Transition{
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
	m := api.Workflow{
		Name: name,
		States: []api.State{
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
		Transitions: []api.Transition{
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
	m := api.Workflow{
		Name: name,
		States: []api.State{
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
		Transitions: []api.Transition{
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
	immediate := api.IMMEDIATE
	m := api.Workflow{
		Name: name,
		States: []api.State{
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
		Transitions: []api.Transition{
			{
				From:     state1,
				To:       state2,
				Eligible: eligibleWfx,
				Action:   &immediate,
			},
			{
				From:     state1,
				To:       state3,
				Eligible: eligibleWfx,
				Action:   &immediate,
			},
		},
	}
	err := ValidateWorkflow(&m)
	assert.ErrorContains(t, err, "more than one immediate action from state state1")
}

func TestValidateWorkflow_ImmediateActionUnique(t *testing.T) {
	immediate := api.IMMEDIATE
	wait := api.WAIT
	state3 := "state3"
	m := api.Workflow{
		Name: name,
		States: []api.State{
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
		Transitions: []api.Transition{
			{
				From:     state1,
				To:       state2,
				Eligible: eligibleWfx,
				Action:   &immediate,
			},
			{
				From:     state1,
				To:       state3,
				Eligible: eligibleWfx,
				Action:   &wait,
			},
		},
	}
	err := ValidateWorkflow(&m)
	assert.ErrorContains(t, err, "transition with source state1 contains impossible transition")
}

func TestValidateWorkflow_ImmediateActionUnique2(t *testing.T) {
	immediate := api.IMMEDIATE
	wait := api.WAIT
	m := api.Workflow{
		Name: name,
		States: []api.State{
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
		Transitions: []api.Transition{
			{
				From:     state1,
				To:       state1,
				Eligible: eligibleWfx,
				Action:   &wait,
			},
			{
				From:     state1,
				To:       state2,
				Eligible: eligibleWfx,
				Action:   &immediate,
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
	m := api.Workflow{
		Name: name,
		States: []api.State{
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
		Transitions: []api.Transition{
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
	m := api.Workflow{
		Name: name,
		States: []api.State{
			{
				Name: state1,
			},
			{
				Name: state2,
			},
		},
		Transitions: []api.Transition{
			{
				From:     state1,
				To:       state2,
				Eligible: eligibleClient,
			},
		},
		Groups: []api.Group{
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
	m := api.Workflow{
		Name: name,
		States: []api.State{
			{
				Name: state1,
			},
			{
				Name: state2,
			},
		},
		Transitions: []api.Transition{
			{
				From:     state1,
				To:       state2,
				Eligible: eligibleClient,
			},
		},
		Groups: []api.Group{
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

func TestValidateWorkflow_StateOrderDoesNotChangeInitialState(t *testing.T) {
	orders := []struct {
		name   string
		states []string
	}{
		{"initial-first", []string{"start", "middle", "end"}},
		{"initial-last", []string{"end", "middle", "start"}},
		{"initial-between", []string{"middle", "start", "end"}},
	}
	for _, order := range orders {
		for _, selfLoop := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/self-loop=%v", order.name, selfLoop), func(t *testing.T) {
				model := api.Workflow{
					Name: "ordered-chain",
					Transitions: []api.Transition{
						{From: "start", To: "middle", Eligible: api.CLIENT},
						{From: "middle", To: "end", Eligible: api.CLIENT},
					},
				}
				for _, state := range order.states {
					model.States = append(model.States, api.State{Name: state})
				}
				if selfLoop {
					model.Transitions = append(model.Transitions,
						api.Transition{From: "start", To: "start", Eligible: api.CLIENT})
				}
				assert.NoError(t, ValidateWorkflow(&model))
			})
		}
	}
}

func TestValidateWorkflow_MultipleInitialStatesRemainInvalid(t *testing.T) {
	for _, states := range [][]string{
		{"left", "right", "end"},
		{"end", "right", "left"},
	} {
		t.Run(strings.Join(states, "-"), func(t *testing.T) {
			model := api.Workflow{
				Name: "two-initial-states",
				Transitions: []api.Transition{
					{From: "left", To: "end", Eligible: api.CLIENT},
					{From: "right", To: "end", Eligible: api.CLIENT},
				},
			}
			for _, state := range states {
				model.States = append(model.States, api.State{Name: state})
			}
			err := ValidateWorkflow(&model)
			assert.EqualError(t, err, "workflow must have exactly one INITIAL state")
		})
	}
}

func TestValidateWorkflow_CycleTakesPrecedenceOverMissingInitialState(t *testing.T) {
	model := api.Workflow{
		Name:   "zero-sources",
		States: []api.State{{Name: "a"}, {Name: "b"}, {Name: "c"}},
		Transitions: []api.Transition{
			{From: "a", To: "a", Eligible: api.CLIENT},
			{From: "b", To: "c", Eligible: api.CLIENT},
			{From: "c", To: "b", Eligible: api.CLIENT},
		},
	}
	assert.ErrorContains(t, ValidateWorkflow(&model), "contains cycle")
}
