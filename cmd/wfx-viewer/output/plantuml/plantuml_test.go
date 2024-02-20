package plantuml

import (
	"bytes"
	"testing"

	approvals "github.com/approvals/go-approval-tests"
	"github.com/siemens/wfx/workflow/dau"
	"github.com/stretchr/testify/require"
)

func TestGenerate(t *testing.T) {
	buf := new(bytes.Buffer)
	gen := NewGenerator()
	err := gen.Generate(buf, dau.DirectWorkflow())
	require.NoError(t, err)
	approvals.VerifyString(t, buf.String())
}
