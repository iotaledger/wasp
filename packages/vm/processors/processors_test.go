package processors

import (
	"github.com/iotaledger/wasp/packages/vm/builtinvm"
	"testing"

	"github.com/iotaledger/wasp/packages/coret"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/root"
	"github.com/stretchr/testify/assert"
)

func TestBasic(t *testing.T) {
	p := MustNew()

	rootproc, err := p.GetOrCreateProcessor(
		&root.RootContractRecord,
		func(hashing.HashValue) (string, []byte, error) { return builtinvm.VMType, nil, nil },
	)
	assert.NoError(t, err)

	_, exists := rootproc.GetEntryPoint(0)
	assert.False(t, exists)

	_, exists = rootproc.GetEntryPoint(coret.Hn(root.FuncDeployContract))
	assert.True(t, exists)
}
