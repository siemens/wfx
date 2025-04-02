package mermaid

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
	"github.com/siemens/wfx/generated/api"
	"github.com/siemens/wfx/internal/workflow"
	"github.com/spf13/pflag"
)

type Generator struct{}

func NewGenerator() *Generator {
	return &Generator{}
}

func (g *Generator) RegisterFlags(_ *pflag.FlagSet) {}

func (g *Generator) Generate(out io.Writer, wf *api.Workflow) error {
	_, _ = out.Write([]byte("stateDiagram-v2\n"))

	initialState := *workflow.FindInitialState(wf)
	_, _ = fmt.Fprintf(out, "    [*] --> %s\n", initialState)
	for _, transition := range wf.Transitions {
		_, _ = out.Write([]byte("    "))
		_, _ = out.Write([]byte(transition.From))
		_, _ = out.Write([]byte(" --> "))
		_, _ = out.Write([]byte(transition.To))
		_, _ = out.Write([]byte(": "))
		_, _ = out.Write([]byte(transition.Eligible))
		_, _ = out.Write([]byte("\n"))
	}

	finalStates := workflow.FindFinalStates(wf)
	for _, state := range finalStates {
		_, _ = out.Write([]byte("    "))
		_, _ = out.Write([]byte(state))
		_, _ = out.Write([]byte(" --> [*]\n"))
	}

	// colors
	cp := colors.NewColorPalette(wf)
	for _, state := range wf.States {
		fgColor, bgColor := cp.StateColor(state.Name)
		_, _ = fmt.Fprintf(out, "    classDef cl_%s color:%s,fill:%s\n", state.Name, fgColor, bgColor)
		_, _ = fmt.Fprintf(out, "    class %s cl_%s\n", state.Name, state.Name)
	}

	// add legend
	_, _ = fmt.Fprintf(out, "    Note right of %s: <b>Group to Color Mapping</b><br/>", initialState)
	lines := make([]string, 0)
	for _, group := range wf.Groups {
		hex := cp.GroupColor(group.Name).ToHEX().String()
		lines = append(lines, fmt.Sprintf(`<font color="%s">%s</font> - %s`, hex, group.Name, group.Description))
	}
	_, _ = out.Write([]byte(strings.Join(lines, "<br/>")))
	_, _ = out.Write([]byte("\n"))

	return nil
}
