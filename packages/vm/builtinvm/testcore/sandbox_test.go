package testcore

import (
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/root"
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

func TestMainCallsFromFullEP(t *testing.T) {
	glb := solo.New(t, false, false)
	chain := glb.NewChain(nil, "ch1")
	user := glb.NewSignatureSchemeWithFunds()
	userAddress := user.Address()
	userAgentID := coretypes.NewAgentIDFromAddress(userAddress)

	req := solo.NewCall(root.Interface.Name, root.FuncGrantDeploy,
		root.ParamDeployer, userAgentID,
	)
	_, err := chain.PostRequest(req, nil)
	require.NoError(t, err)

	err = chain.DeployContract(user, test_sandbox.Interface.Name, test_sandbox.Interface.ProgramHash)
	require.NoError(t, err)

	contractID := coretypes.NewContractID(chain.ChainID, coretypes.Hn(test_sandbox.Interface.Name))
	agentID := coretypes.NewAgentIDFromContractID(contractID)

	req = solo.NewCall(test_sandbox.Interface.Name, test_sandbox.FuncCheckContextFromFullEP,
		test_sandbox.ParamChainID, chain.ChainID,
		test_sandbox.ParamAgentID, agentID,
		test_sandbox.ParamCaller, userAgentID,
		test_sandbox.ParamChainOwnerID, chain.OriginatorAgentID,
		test_sandbox.ParamContractID, contractID,
		test_sandbox.ParamContractCreator, userAgentID,
	)
	_, err = chain.PostRequest(req, user)
	require.NoError(t, err)
}

func TestMainCallsFromViewEP(t *testing.T) {
	glb := solo.New(t, false, false)
	chain := glb.NewChain(nil, "ch1")
	user := glb.NewSignatureSchemeWithFunds()
	userAddress := user.Address()
	userAgentID := coretypes.NewAgentIDFromAddress(userAddress)

	req := solo.NewCall(root.Interface.Name, root.FuncGrantDeploy,
		root.ParamDeployer, userAgentID,
	)
	_, err := chain.PostRequest(req, nil)
	require.NoError(t, err)

	err = chain.DeployContract(user, test_sandbox.Interface.Name, test_sandbox.Interface.ProgramHash)
	require.NoError(t, err)

	contractID := coretypes.NewContractID(chain.ChainID, coretypes.Hn(test_sandbox.Interface.Name))

	_, err = chain.CallView(test_sandbox.Interface.Name, test_sandbox.FuncCheckContextFromViewEP,
		test_sandbox.ParamChainID, chain.ChainID,
		test_sandbox.ParamContractID, contractID,
	)
	require.NoError(t, err)
}
