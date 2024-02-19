package workflow

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"sort"

	"github.com/siemens/wfx/generated/model"
)

// FindStateGroup tries to find the group of a state. If not found, it returns the empty string.
func FindStateGroup(workflow *model.Workflow, state string) string {
	for _, group := range workflow.Groups {
		for _, s := range group.States {
			if s == state {
				return group.Name
			}
		}
	}
	return ""
}

// FollowImmediateTransitions follows the edges of type `actor` starting at the `from` state.
func FollowImmediateTransitions(workflow *model.Workflow, from string) string {
	// map of transitions which we handle
	jump := make(map[string]string, len(workflow.Transitions))
	for _, t := range workflow.Transitions {
		if t.Eligible == model.EligibleEnumWFX && t.Action == model.ActionEnumIMMEDIATE {
			jump[t.From] = t.To
		}
	}

	current := from
	for {
		// follow the path
		to, ok := jump[current]
		if !ok {
			// we have reached the final destination
			return current
		}
		current = to
	}
}

func FindInitialState(workflow *model.Workflow) *string {
	parent := make(map[string]string, len(workflow.States))
	for _, state := range workflow.States {
		parent[state.Name] = ""
	}
	for _, transition := range workflow.Transitions {
		if transition.From != transition.To {
			parent[transition.To] = transition.From
		}
	}
	// we know that there must be exactly one initial state due to model validation
	for node, predecessor := range parent {
		if predecessor == "" {
			return &node
		}
	}
	return nil
}

func FindFinalStates(workflow *model.Workflow) []string {
	finalStateMap := make(map[string]bool, len(workflow.States))
	// add all states and then remove the ones that are not final
	for _, state := range workflow.States {
		finalStateMap[state.Name] = true
	}
	for _, transition := range workflow.Transitions {
		delete(finalStateMap, transition.From)
	}
	finalStates := make([]string, 0, len(finalStateMap))
	for name := range finalStateMap {
		finalStates = append(finalStates, name)
	}
	sort.Strings(finalStates)
	return finalStates
}
