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

	"github.com/rs/zerolog/log"
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
	mustWrite := func(s string) {
		_, err := out.Write([]byte(s))
		if err != nil {
			log.Fatal().Err(err).Msg("Failed writing string")
		}
	}

	mustWrite("@startuml\n")

	allStates := make(map[string]*model.State, len(workflow.States))
	for _, state := range workflow.States {
		allStates[state.Name] = state
	}

	cp := colors.NewColorPalette(workflow)

	for _, state := range workflow.States {
		fgColor, bgColor := cp.StateColor(state.Name)
		mustWrite(fmt.Sprintf("state %s as \"<color:%s>%s</color>\" %s: %s\n", state.Name, fgColor, state.Name, bgColor, state.Description))
	}

	// add transitions
	for _, transition := range workflow.Transitions {
		mustWrite(fmt.Sprintf("%s --> %s: %s", transition.From, transition.To, string(transition.Eligible)))
		if string(transition.Action) != "" {
			mustWrite(fmt.Sprintf(" [%s]", string(transition.Action)))
		}
		mustWrite("\n")
	}

	// add legend
	mustWrite("legend right\n")
	mustWrite("  | Color | Group | Description |\n")
	for _, group := range workflow.Groups {
		color := cp.GroupColor(group.Name)
		hex := color.ToHEX().String()
		mustWrite(fmt.Sprintf("  | <%s> | %s | %s |\n", hex, group.Name, group.Description))
	}
	mustWrite(fmt.Sprintf("  | <%s> | %s | %s |\n", colors.DefaultBgColor, "", "The state doesn't belong to any group."))
	mustWrite("endlegend\n")

	mustWrite("@enduml\n")
	return nil
}
