package colors

/*
 * SPDX-FileCopyrightText: 2024 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"image/color/palette"

	"github.com/siemens/wfx/generated/model"
	"gopkg.in/go-playground/colors.v1"
)

const (
	DefaultFgColor = "white"
	DefaultBgColor = "#000000"
)

type ColorPalette struct {
	stateToGroup map[string]string
	groupColor   map[string]colors.RGBAColor
}

func NewColorPalette(workflow *model.Workflow) ColorPalette {
	cp := ColorPalette{
		stateToGroup: make(map[string]string, len(workflow.States)),
		groupColor:   make(map[string]colors.RGBAColor),
	}

	black := colors.RGBAColor{R: 0x00, G: 0x00, B: 0x00, A: 0xff}
	white := colors.RGBAColor{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
	darkColors := make([]colors.RGBAColor, 0, len(palette.Plan9)/2)
	lightColors := make([]colors.RGBAColor, 0, len(palette.Plan9)/2)
	for _, c := range palette.Plan9 {
		r, g, b, _ := c.RGBA()

		c2, _ := colors.RGBA(uint8(r), uint8(g), uint8(b), 0)
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

	// states belonging to a group
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

		cp.groupColor[group.Name] = chosenColor
		for _, s := range group.States {
			cp.stateToGroup[s] = group.Name
		}
	}
	return cp
}

func (cp *ColorPalette) GroupColor(group string) *colors.RGBAColor {
	if color, ok := cp.groupColor[group]; ok {
		return &color
	}
	return nil
}

func (cp *ColorPalette) StateColor(state string) (string, string) {
	var fgColor, bgColor string
	if group := cp.StateToGroup(state); group != nil {
		// if missing (no group), it picks the default value (0) which is black
		c := cp.GroupColor(*group)
		bgColor = c.ToHEX().String()
		if c.IsDark() {
			fgColor = "white"
		} else {
			fgColor = "black"
		}
	} else {
		// use default values if state doesn't belong to any group
		fgColor = DefaultFgColor
		bgColor = DefaultBgColor
	}
	return fgColor, bgColor
}

func (cp *ColorPalette) StateToGroup(state string) *string {
	group, ok := cp.stateToGroup[state]
	if !ok {
		return nil
	}
	return &group
}
