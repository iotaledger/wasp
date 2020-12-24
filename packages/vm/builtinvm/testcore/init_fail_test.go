package testcore

import (
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/testcore/test_init_fail"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSuccess(t *testing.T) {
	glb := solo.New(t, false, false)
	chain := glb.NewChain(nil, "chain1")
	err := chain.DeployContract(nil, test_init_fail.Name, test_init_fail.ProgramHash)
	require.NoError(t, err)
}

func TestFail(t *testing.T) {
	glb := solo.New(t, false, false)
	chain := glb.NewChain(nil, "chain1")
	err := chain.DeployContract(nil, test_init_fail.Name, test_init_fail.ProgramHash, test_init_fail.ParamFail, 1)
	require.Error(t, err)
}

func TestFailRepeat(t *testing.T) {
	glb := solo.New(t, false, false)
	chain := glb.NewChain(nil, "chain1")
	err := chain.DeployContract(nil, test_init_fail.Name, test_init_fail.ProgramHash, test_init_fail.ParamFail, 1)
	require.Error(t, err)
	_, rec := chain.GetInfo()
	require.EqualValues(t, 4, len(rec))

	// repeat must succeed
	err = chain.DeployContract(nil, test_init_fail.Name, test_init_fail.ProgramHash)
	require.NoError(t, err)
	_, rec = chain.GetInfo()
	require.EqualValues(t, 5, len(rec))
}
