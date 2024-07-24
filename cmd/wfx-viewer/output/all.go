package output

/*
 * SPDX-FileCopyrightText: 2024 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"io"

	"github.com/siemens/wfx/cmd/wfx-viewer/output/mermaid"
	"github.com/siemens/wfx/cmd/wfx-viewer/output/plantuml"
	"github.com/siemens/wfx/cmd/wfx-viewer/output/smcat"
	"github.com/siemens/wfx/cmd/wfx-viewer/output/svg"
	"github.com/siemens/wfx/generated/api"
	"github.com/spf13/pflag"
)

type Generator interface {
	RegisterFlags(f *pflag.FlagSet)
	Generate(out io.Writer, workflow *api.Workflow) error
}

var Generators = make(map[string]Generator)

func init() {
	Generators["svg"] = svg.NewGenerator()
	Generators["plantuml"] = plantuml.NewGenerator()
	Generators["smcat"] = smcat.NewGenerator()
	Generators["mermaid"] = mermaid.NewGenerator()
}
