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
	"image/color/palette"
	"io"

	"github.com/rs/zerolog/log"
	"github.com/siemens/wfx/generated/model"
	"github.com/spf13/pflag"
	"gopkg.in/go-playground/colors.v1"
)

const (
	defaultTextColor       = "white"
	defaultBackgroundColor = "#000000"
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

	black := colors.RGBAColor{R: 0x00, G: 0x00, B: 0x00, A: 0xff}
	white := colors.RGBAColor{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
	darkColors := make([]colors.RGBAColor, 0, len(palette.Plan9)/2)
	lightColors := make([]colors.RGBAColor, 0, len(palette.Plan9)/2)
	for _, c := range palette.Plan9 {
		r, g, b, _ := c.RGBA()

		c2, err := colors.RGBA(uint8(r), uint8(g), uint8(b), 0)
		if err != nil {
			log.Fatal().Err(err).Msg("Error converting color")
		}
		if c2.IsDark() {
			darkColors = append(darkColors, *c2)
		} else {
			lightColors = append(lightColors, *c2)
		}
	}
	// shuffle to get more distinctive colors
	{
		n := len(lightColors)
		for i := range lightColors {
			j := (i * 60) % n
			lightColors[i], lightColors[j] = lightColors[j], lightColors[i]
		}
	}
	{
		n := len(darkColors)
		for i := range darkColors {
			j := (i * 60) % n
			darkColors[i], darkColors[j] = darkColors[j], darkColors[i]
		}
	}

	groupToColor := make(map[string]colors.RGBAColor)

	// states belonging to a group
	stateToGroup := make(map[string]string, len(workflow.States))
	for _, group := range workflow.Groups {
		var chosenColor colors.RGBAColor
		// prefer light colors
		if len(lightColors) > 0 {
			// pop
			chosenColor, lightColors = lightColors[0], lightColors[1:]
			if chosenColor == white && len(lightColors) > 0 {
				// ignore white
				chosenColor, lightColors = lightColors[0], lightColors[1:]
			}

		} else if len(darkColors) > 0 {
			chosenColor, darkColors = darkColors[0], darkColors[1:]
			if chosenColor == black && len(darkColors) > 0 {
				// ignore black
				chosenColor, darkColors = darkColors[0], darkColors[1:]
			}
		}

		groupToColor[group.Name] = chosenColor
		for _, s := range group.States {
			stateToGroup[s] = group.Name
		}
	}

	hasStatesWithoutGroup := false
	for _, state := range workflow.States {
		var textColor string
		var backgroundColor string

		if group, found := stateToGroup[state.Name]; found {
			// if missing (no group), it picks the default value (0) which is black
			c := groupToColor[group]
			backgroundColor = c.ToHEX().String()

			if c.IsDark() {
				textColor = "white"
			} else {
				textColor = "black"
			}
		} else {
			// use default values if state doesn't belong to any group
			hasStatesWithoutGroup = true
			textColor = defaultTextColor
			backgroundColor = defaultBackgroundColor
		}

		mustWrite(fmt.Sprintf("state %s as \"<color:%s>%s</color>\" %s: %s\n", state.Name, textColor, state.Name, backgroundColor, state.Description))
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
		color := groupToColor[group.Name]
		hex := color.ToHEX().String()
		mustWrite(fmt.Sprintf("  | <%s> | %s | %s |\n", hex, group.Name, group.Description))
	}
	if hasStatesWithoutGroup {
		mustWrite(fmt.Sprintf("  | <%s> | %s | %s |\n", defaultBackgroundColor, "", "The state doesn't belong to any group."))
	}
	mustWrite("endlegend\n")

	mustWrite("@enduml\n")
	return nil
}
