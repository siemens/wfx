package plantuml

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"fmt"
	"io"

	"github.com/siemens/wfx/cmd/wfx-viewer/colors"
	"github.com/siemens/wfx/generated/model"
	"github.com/spf13/pflag"
)

type Generator struct{}

func NewGenerator() *Generator {
	return &Generator{}
}

func (g *Generator) RegisterFlags(_ *pflag.FlagSet) {}

func (g *Generator) Generate(out io.Writer, workflow *model.Workflow) error {
	_, _ = out.Write([]byte("@startuml\n"))

	allStates := make(map[string]*model.State, len(workflow.States))
	for _, state := range workflow.States {
		allStates[state.Name] = state
	}

	cp := colors.NewColorPalette(workflow)

	for _, state := range workflow.States {
		fgColor, bgColor := cp.StateColor(state.Name)
		_, _ = out.Write([]byte((fmt.Sprintf("state %s as \"<color:%s>%s</color>\" %s: %s\n", state.Name, fgColor, state.Name, bgColor, state.Description))))
	}

	// add transitions
	for _, transition := range workflow.Transitions {
		_, _ = out.Write([]byte((fmt.Sprintf("%s --> %s: %s", transition.From, transition.To, string(transition.Eligible)))))
		if string(transition.Action) != "" {
			_, _ = out.Write([]byte((fmt.Sprintf(" [%s]", string(transition.Action)))))
		}
		_, _ = out.Write([]byte("\n"))
	}

	// add legend
	_, _ = out.Write([]byte("legend right\n"))
	_, _ = out.Write([]byte(("  | Color | Group | Description |\n")))
	for _, group := range workflow.Groups {
		color := cp.GroupColor(group.Name)
		hex := color.ToHEX().String()
		_, _ = out.Write([]byte((fmt.Sprintf("  | <%s> | %s | %s |\n", hex, group.Name, group.Description))))
	}
	_, _ = out.Write([]byte((fmt.Sprintf("  | <%s> | %s | %s |\n", colors.DefaultBgColor, "", "The state doesn't belong to any group."))))
	_, _ = out.Write([]byte(("endlegend\n")))

	_, _ = out.Write([]byte(("@enduml\n")))
	return nil
}
