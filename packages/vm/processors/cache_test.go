package processors

import (
	"testing"

	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
	"github.com/stretchr/testify/assert"
)

func TestBasic(t *testing.T) {
	p := MustNew(NewConfig())

	rec := root.ContractRecordFromContractInfo(root.Contract, &iscp.NilAgentID)
	rootproc, err := p.GetOrCreateProcessor(
		rec,
		func(hashing.HashValue) (string, []byte, error) { return vmtypes.Core, nil, nil },
	)
	assert.NoError(t, err)

	// TODO always exists because it returns default handler
	ep, exists := rootproc.GetEntryPoint(0)
	assert.True(t, exists)
	assert.Same(t, ep.(*coreutil.EntryPointHandler).Info, &coreutil.FuncFallback)

	ep, exists = rootproc.GetEntryPoint(root.FuncDeployContract.Hname())
	assert.True(t, exists)
	assert.Same(t, ep.(*coreutil.EntryPointHandler).Info, &root.FuncDeployContract)
}
