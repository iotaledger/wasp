package testcore

import (
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/testcore/test_sandbox"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestBasic(t *testing.T) {
	glb := solo.New(t, false, false)
	chain := glb.NewChain(nil, "chain1")

	chain.CheckChain()
	err := chain.DeployContract(nil, test_sandbox.Interface.Name, test_sandbox.Interface.ProgramHash)
	require.NoError(t, err)
}
