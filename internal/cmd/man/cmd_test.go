package man

/*
 * SPDX-FileCopyrightText: 2024 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"os"
	"sort"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecuteCommand(t *testing.T) {
	dummy := &cobra.Command{
		Use:              "dummy",
		Short:            "a command for testing purposes",
		TraverseChildren: true,
	}
	dummy.AddCommand(NewCommand())

	flags := dummy.PersistentFlags()
	flags.String("foo", "bar", "example argument")

	tmpDir, _ := os.MkdirTemp("", "man")
	t.Cleanup(func() {
		_ = os.RemoveAll(tmpDir)
	})

	dummy.SetArgs([]string{"man", "--dir", tmpDir})
	err := dummy.Execute()
	require.NoError(t, err)

	dir, _ := os.Open(tmpDir)
	t.Cleanup(func() { _ = dir.Close() })

	entries, _ := dir.Readdir(-1)

	fnames := make([]string, 0, len(entries))
	for _, entry := range entries {
		fnames = append(fnames, entry.Name())
	}

	expected := []string{
		"dummy-completion-bash.1",
		"dummy-completion-fish.1",
		"dummy-completion-powershell.1",
		"dummy-completion-zsh.1",
		"dummy-completion.1",
		"dummy-man.1",
		"dummy.1",
	}

	sort.Strings(fnames)
	sort.Strings(expected)
	assert.Equal(t, expected, fnames)
}
