package validate

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"bytes"
	"os"
	"testing"

	"github.com/siemens/wfx/workflow/dau"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommand_NoWorkflowGiven(t *testing.T) {
	err := NewCommand().Execute()
	require.Error(t, err)
	assert.ErrorContains(t, err, "workflow must be provided either via file or stdin")
}

func TestCommand_Stdin(t *testing.T) {
	buf := new(bytes.Buffer)
	_, _ = buf.WriteString(dau.DirectYAML)
	cmd := NewCommand()
	cmd.SetIn(buf)
	cmd.SetArgs([]string{"-"})
	err := cmd.Execute()
	require.NoError(t, err)
}

func TestCommand_Fname(t *testing.T) {
	tmpFile, _ := os.CreateTemp(os.TempDir(), "workflow")
	_, _ = tmpFile.Write([]byte(dau.PhasedYAML))
	_ = tmpFile.Close()
	t.Cleanup(func() { _ = os.Remove(tmpFile.Name()) })

	cmd := NewCommand()
	cmd.SetArgs([]string{tmpFile.Name()})
	err := cmd.Execute()
	require.NoError(t, err)
}
