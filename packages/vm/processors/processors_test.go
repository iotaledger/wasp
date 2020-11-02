package processors

import (
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/vm/builtinvm"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/dummyprocessor"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/root"
	"github.com/iotaledger/wasp/packages/vm/examples"
	"github.com/iotaledger/wasp/packages/vm/examples/donatewithfeedback/dwfimpl"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBasic(t *testing.T) {
	p := MustNew()
	dummyp, ok := p.GetProcessor(dummyprocessor.ProgramHash)
	assert.False(t, ok)

	proc, ok := p.GetProcessor(root.ProgramHash)
	assert.True(t, ok)

	_, err := p.NewProcessor(root.ProgramHash[:], builtinvm.VMType)
	assert.Error(t, err)

	_, exists := proc.GetEntryPoint(0)
	assert.False(t, exists)

	_, exists = proc.GetEntryPoint(coretypes.NewEntryPointCodeFromFunctionName("initialize"))
	assert.True(t, exists)

	_, err = p.NewProcessor([]byte(dwfimpl.ProgramHash), examples.VMType)
	assert.NoError(t, err)

	_, err = p.NewProcessor(dummyprocessor.ProgramHash[:], builtinvm.VMType)
	assert.NoError(t, err)

	dummyp, ok = p.GetProcessor(dummyprocessor.ProgramHash)
	assert.True(t, ok)

	_, exists = dummyp.GetEntryPoint(0)
	assert.False(t, exists)
}
