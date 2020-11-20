package processors

import (
	"testing"

	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/root"
	"github.com/stretchr/testify/assert"
)

func TestBasic(t *testing.T) {
	p := MustNew()

	rootproc, err := p.GetOrCreateProcessor(
		&root.RootContractRecord,
		func(*hashing.HashValue) ([]byte, error) { return root.ProgramHash[:], nil },
	)
	assert.NoError(t, err)

	_, exists := rootproc.GetEntryPoint(0)
	assert.False(t, exists)

	_, exists = rootproc.GetEntryPoint(coretypes.Hn(root.FuncDeployContract))
	assert.True(t, exists)
}
