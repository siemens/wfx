package workflow

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"errors"
	"fmt"

	"github.com/Southclaws/fault"
	"github.com/go-openapi/strfmt"
	"github.com/yourbasic/graph"

	"github.com/siemens/wfx/generated/model"
)

type edge struct {
	From string
	To   string
}

func ValidateWorkflow(workflow *model.Workflow) error {
	if err := workflow.Validate(strfmt.Default); err != nil {
		return fault.Wrap(err)
	}

	stateCount := len(workflow.States)
	stateToNode := make(map[string]int, stateCount)

	for i, s := range workflow.States {
		if _, found := stateToNode[s.Name]; found {
			// State name must be unique
			return fmt.Errorf("%s state has already been created", s.Name)
		}
		stateToNode[s.Name] = i
	}

	{
		// for each state, count in how many groups we can find it
		stateCount := make(map[string]int)
		groupNameCount := make(map[string]int)
		for _, group := range workflow.Groups {
			for _, s := range group.States {
				stateCount[s]++
			}
			groupNameCount[group.Name]++
		}

		// check groups do not overlap
		for state, count := range stateCount {
			if count > 1 {
				return fmt.Errorf("state %s belongs to more than one group", state)
			}
		}

		// ensure that no two groups have the same name
		for name, count := range groupNameCount {
			if count > 1 {
				return fmt.Errorf("group name %s used multiple times", name)
			}
		}
	}

	// build a graph from the transitions
	g := graph.New(stateCount)

	// (from, to) -> [client, wfx]
	transitions := make(map[edge]([]model.EligibleEnum))
	// for each edge (from, _), count the ones containing actor WFX
	outgoingActions := make(map[string]([]model.ActionEnum))

	for _, t := range workflow.Transitions {
		from, foundFrom := stateToNode[t.From]
		to, foundTo := stateToNode[t.To]
		if !foundFrom || !foundTo {
			return fmt.Errorf("transition %s -> %s contains unknown state name", t.From, t.To)
		}
		e := edge{From: t.From, To: t.To}
		transitions[e] = append(transitions[e], t.Eligible)

		action := t.Action
		if action == "" {
			// default action
			action = model.ActionEnumWAIT
		}
		outgoingActions[t.From] = append(outgoingActions[t.From], action)

		if from != to {
			// we allow trivial loops
			g.Add(from, to)
		}
	}

	for e, eligible := range transitions {
		if len(eligible) > 1 {
			eligible := findDuplicate(eligible)
			if eligible != nil {
				return fmt.Errorf("duplicate transition: %s -> %s eligible: %s", e.From, e.To, *eligible)
			}
		}
	}
	for from, actions := range outgoingActions {
		// count immediate actions
		count := 0
		for _, act := range actions {
			if act == model.ActionEnumIMMEDIATE {
				count++
			}
		}
		if count > 1 {
			return fmt.Errorf("more than one immediate action from state %s", from)
		}
		if count != 0 && len(actions) > 1 {
			return fmt.Errorf("transition with source %s contains impossible transition", from)
		}
	}

	dfsResult := dfs(g)

	{
		// Workflow should have exactly one initial state.
		var initialStates []string
		for v, parent := range dfsResult.Prev {
			if parent == noParent {
				initialStates = append(initialStates, workflow.States[v].Name)
			}
		}
		if len(initialStates) != 1 {
			return errors.New("workflow must have exactly one INITIAL state")
		}
	}

	for _, cycle := range dfsResult.Cycles {
		return fmt.Errorf("workflow contains cycle from %s to %s",
			workflow.States[cycle.From].Name, workflow.States[cycle.To].Name)
	}

	return nil
}

func findDuplicate[T comparable](values []T) *T {
	n := len(values)
	seen := make(map[T]bool, len(values))
	for i := 0; i < n; i++ {
		if seen[values[i]] {
			return &values[i]
		}
		seen[values[i]] = true
	}
	return nil
}
