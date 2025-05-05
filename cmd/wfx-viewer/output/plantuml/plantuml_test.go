package plantuml

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
	exepcted := `@startuml
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
	assert.Equal(t, exepcted, actual)
}
