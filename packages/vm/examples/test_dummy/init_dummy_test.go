package test_env

import (
	"github.com/iotaledger/wasp/packages/vm/alone"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSuccess(t *testing.T) {
	glb := alone.New(t, false, false)
	chain := glb.NewChain(nil, "chain1")
	err := chain.DeployContract(nil, "dummy", ProgramHash)
	require.NoError(t, err)
}

func TestFail(t *testing.T) {
	glb := alone.New(t, false, false)
	chain := glb.NewChain(nil, "chain1")
	err := chain.DeployContract(nil, "dummy", ProgramHash, "fail", 1)
	require.Error(t, err)
}

func TestFailRepeat(t *testing.T) {
	glb := alone.New(t, false, false)
	chain := glb.NewChain(nil, "chain1")
	err := chain.DeployContract(nil, "dummy", ProgramHash, "fail", 1)
	require.Error(t, err)
	_, _, rec := chain.GetInfo()
	require.EqualValues(t, 3, len(rec))

	// repeat must succeed
	err = chain.DeployContract(nil, "dummy", ProgramHash)
	require.NoError(t, err)
	_, _, rec = chain.GetInfo()
	require.EqualValues(t, 4, len(rec))
}
