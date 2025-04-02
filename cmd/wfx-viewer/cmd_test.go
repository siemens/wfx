package main

import (
	"bytes"
	"os"
	"testing"

	approvals "github.com/approvals/go-approval-tests"
	"github.com/siemens/wfx/workflow/dau"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestPlantUML_Stdout(t *testing.T) {
	f := rootCmd.PersistentFlags()
	_ = f.Set(outputFlag, "")
	_ = f.Set(outputFormatFlag, "plantuml")

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)

	tmpFile, _ := os.CreateTemp(os.TempDir(), "TestPlantUML.*")
	b, _ := yaml.Marshal(dau.DirectWorkflow())
	_, _ = tmpFile.Write(b)
	_ = tmpFile.Close()
	defer func() {
		_ = os.Remove(tmpFile.Name())
	}()

	rootCmd.SetArgs([]string{tmpFile.Name()})
	err := rootCmd.Execute()
	require.NoError(t, err)

	approvals.VerifyString(t, buf.String())
}

func TestPlantUML_File(t *testing.T) {
	f := rootCmd.PersistentFlags()

	outFile, _ := os.CreateTemp(os.TempDir(), "TestPlantUML.*")
	defer func() {
		_ = os.Remove(outFile.Name())
	}()

	_ = f.Set(outputFlag, outFile.Name())
	_ = f.Set(outputFormatFlag, "plantuml")

	tmpFile, _ := os.CreateTemp(os.TempDir(), "TestPlantUML.*")
	b, _ := yaml.Marshal(dau.DirectWorkflow())
	_, _ = tmpFile.Write(b)
	_ = tmpFile.Close()
	t.Cleanup(func() { _ = os.Remove(tmpFile.Name()) })

	err := rootCmd.RunE(rootCmd, []string{tmpFile.Name()})
	require.NoError(t, err)

	content, _ := os.ReadFile(outFile.Name())
	approvals.VerifyString(t, string(content))
}
