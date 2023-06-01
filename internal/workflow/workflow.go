package workflow

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import "github.com/siemens/wfx/generated/model"

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
