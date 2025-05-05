package main

import (
	"bytes"
	"os"
	"testing"

	"github.com/siemens/wfx/workflow/dau"
	"github.com/stretchr/testify/assert"
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

	actual := buf.String()
	expected := `@startuml
state INSTALL as "<color:black>INSTALL</color>" #00cc00: instruct client to start installation
state INSTALLING as "<color:black>INSTALLING</color>" #00cc00: installation progress update from client
state INSTALLED as "<color:black>INSTALLED</color>" #00cc00: client signaled installation success
state ACTIVATE as "<color:black>ACTIVATE</color>" #00cc00: instruct client to start activation
state ACTIVATING as "<color:black>ACTIVATING</color>" #00cc00: client activates update
state ACTIVATED as "<color:black>ACTIVATED</color>" #4993dd: client signaled activation success
state TERMINATED as "<color:black>TERMINATED</color>" #9393dd: client aborted update with error
INSTALL --> INSTALLING: CLIENT
INSTALL --> TERMINATED: CLIENT
INSTALLING --> INSTALLING: CLIENT
INSTALLING --> TERMINATED: CLIENT
INSTALLING --> INSTALLED: CLIENT
INSTALLED --> ACTIVATE: WFX [IMMEDIATE]
ACTIVATE --> ACTIVATING: CLIENT
ACTIVATE --> TERMINATED: CLIENT
ACTIVATING --> ACTIVATING: CLIENT
ACTIVATING --> TERMINATED: CLIENT
ACTIVATING --> ACTIVATED: CLIENT
legend right
  | Color | Group | Description |
  | <#00cc00> | OPEN | regular workflow-advancing states |
  | <#4993dd> | CLOSED | a successful update's terminal states |
  | <#9393dd> | FAILED | a failed update's terminal states |
  | <#000000> |  | The state doesn't belong to any group. |
endlegend
@enduml
`
	assert.Equal(t, expected, actual)
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
	actual := string(content)
	expected := `@startuml
state INSTALL as "<color:black>INSTALL</color>" #00cc00: instruct client to start installation
state INSTALLING as "<color:black>INSTALLING</color>" #00cc00: installation progress update from client
state INSTALLED as "<color:black>INSTALLED</color>" #00cc00: client signaled installation success
state ACTIVATE as "<color:black>ACTIVATE</color>" #00cc00: instruct client to start activation
state ACTIVATING as "<color:black>ACTIVATING</color>" #00cc00: client activates update
state ACTIVATED as "<color:black>ACTIVATED</color>" #4993dd: client signaled activation success
state TERMINATED as "<color:black>TERMINATED</color>" #9393dd: client aborted update with error
INSTALL --> INSTALLING: CLIENT
INSTALL --> TERMINATED: CLIENT
INSTALLING --> INSTALLING: CLIENT
INSTALLING --> TERMINATED: CLIENT
INSTALLING --> INSTALLED: CLIENT
INSTALLED --> ACTIVATE: WFX [IMMEDIATE]
ACTIVATE --> ACTIVATING: CLIENT
ACTIVATE --> TERMINATED: CLIENT
ACTIVATING --> ACTIVATING: CLIENT
ACTIVATING --> TERMINATED: CLIENT
ACTIVATING --> ACTIVATED: CLIENT
legend right
  | Color | Group | Description |
  | <#00cc00> | OPEN | regular workflow-advancing states |
  | <#4993dd> | CLOSED | a successful update's terminal states |
  | <#9393dd> | FAILED | a failed update's terminal states |
  | <#000000> |  | The state doesn't belong to any group. |
endlegend
@enduml
`
	assert.Equal(t, expected, actual)
}
