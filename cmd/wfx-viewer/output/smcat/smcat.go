package smcat

/*
 * SPDX-FileCopyrightText: 2024 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"fmt"
	"io"
	"strings"

	"github.com/siemens/wfx/cmd/wfx-viewer/colors"
	"github.com/siemens/wfx/generated/model"
	"github.com/siemens/wfx/internal/workflow"
	"github.com/spf13/pflag"
)

type Generator struct{}

func NewGenerator() *Generator {
	return &Generator{}
}

func (g *Generator) RegisterFlags(_ *pflag.FlagSet) {}

func (g *Generator) Generate(out io.Writer, wf *model.Workflow) error {
	cp := colors.NewColorPalette(wf)

	states := make([]string, 0, len(wf.States))
	states = append(states, "initial")

	for _, state := range wf.States {
		_, bgColor := cp.StateColor(state.Name)
		states = append(states, fmt.Sprintf(`%s [color="%s"]`, state.Name, bgColor))
	}
	states = append(states, "final")

	_, _ = out.Write([]byte(strings.Join(states, ",\n")))
	_, _ = out.Write([]byte(";\n\n"))

	initialState := *workflow.FindInitialState(wf)
	_, _ = out.Write([]byte(fmt.Sprintf("initial => %s;\n", initialState)))
	for _, transition := range wf.Transitions {
		_, _ = out.Write([]byte(transition.From))
		_, _ = out.Write([]byte(" => "))
		_, _ = out.Write([]byte(transition.To))
		_, _ = out.Write([]byte(": "))
		_, _ = out.Write([]byte(transition.Eligible))
		_, _ = out.Write([]byte(";\n"))
	}

	finalStates := workflow.FindFinalStates(wf)
	for _, state := range finalStates {
		_, _ = out.Write([]byte(state))
		_, _ = out.Write([]byte(" => final;\n"))
	}

	return nil
}
