package processors

import (
	"github.com/iotaledger/wasp/packages/vm/core"
	"testing"

	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/stretchr/testify/assert"
)

func TestBasic(t *testing.T) {
	p := MustNew()

	rec := root.NewContractRecord(root.Interface, coretypes.AgentID{})
	rootproc, err := p.GetOrCreateProcessor(
		&rec,
		func(hashing.HashValue) (string, []byte, error) { return core.VMType, nil, nil },
	)
	assert.NoError(t, err)

	_, _ := rootproc.GetEntryPoint(0)
	assert.False(t, exists)

	_, _ = rootproc.GetEntryPoint(coretypes.Hn(root.FuncDeployContract))
	assert.True(t, exists)
}
