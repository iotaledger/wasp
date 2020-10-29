package processors

import (
	"github.com/iotaledger/wasp/packages/vm/builtinvm"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/nilprocessor"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBasic(t *testing.T) {
	err := RegisterVMType("builtinvm", builtinvm.Constructor)
	assert.NoError(t, err)

	p, err := New()
	assert.NoError(t, err)
	assert.True(t, p.ExistsProcessor(0))
	assert.False(t, p.ExistsProcessor(1))

	idx, err := p.AddProcessor(nilprocessor.ProgramHash[:], "builtinvm")
	assert.NoError(t, err)
	assert.EqualValues(t, idx, 1)

	p1, exists := p.GetProcessor(1)
	assert.True(t, exists)

	_, exists = p1.GetEntryPoint(0)
	assert.False(t, exists)
}
