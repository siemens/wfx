package smcat

/*
 * SPDX-FileCopyrightText: 2024 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"bytes"
	"testing"

	"github.com/siemens/wfx/workflow/dau"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerate(t *testing.T) {
	buf := new(bytes.Buffer)
	gen := NewGenerator()
	err := gen.Generate(buf, dau.DirectWorkflow())
	require.NoError(t, err)
	actual := buf.String()
	expected := `initial,
INSTALL [color="#00cc00"],
INSTALLING [color="#00cc00"],
INSTALLED [color="#00cc00"],
ACTIVATE [color="#00cc00"],
ACTIVATING [color="#00cc00"],
ACTIVATED [color="#4993dd"],
TERMINATED [color="#9393dd"],
final;

initial => INSTALL;
INSTALL => INSTALLING: CLIENT;
INSTALL => TERMINATED: CLIENT;
INSTALLING => INSTALLING: CLIENT;
INSTALLING => TERMINATED: CLIENT;
INSTALLING => INSTALLED: CLIENT;
INSTALLED => ACTIVATE: WFX;
ACTIVATE => ACTIVATING: CLIENT;
ACTIVATE => TERMINATED: CLIENT;
ACTIVATING => ACTIVATING: CLIENT;
ACTIVATING => TERMINATED: CLIENT;
ACTIVATING => ACTIVATED: CLIENT;
ACTIVATED => final;
TERMINATED => final;
`
	assert.Equal(t, expected, actual)
}
