package testcore

import (
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/core/testcore/test_sandbox"
	"github.com/stretchr/testify/require"
	"testing"
)

func setupForTestSandbox(t *testing.T) (*solo.Solo, *solo.Chain, coretypes.ContractID) {
	glb := solo.New(t, false, false)
	chain := glb.NewChain(nil, "ch1")

	err := chain.DeployContract(nil, test_sandbox.Interface.Name, test_sandbox.Interface.ProgramHash)
	require.NoError(t, err)

	deployed := coretypes.NewContractID(chain.ChainID, coretypes.Hn(test_sandbox.Interface.Name))
	req := solo.NewCall(test_sandbox.Interface.Name, test_sandbox.FuncDoNothing)
	_, err = chain.PostRequest(req, nil)
	require.NoError(t, err)
	return glb, chain, deployed
}
