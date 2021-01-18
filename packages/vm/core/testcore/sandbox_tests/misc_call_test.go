package sandbox_tests

import (
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/core/testcore/test_sandbox"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestChainOwnerID(t *testing.T) {
	_, chain := setupChain(t, nil)
	setupTestSandboxSC(t, chain, nil)

	req := solo.NewCall(SandboxSCName,
		test_sandbox.FuncChainOwnerID,
	)
	ret, err := chain.PostRequest(req, nil)
	require.NoError(t, err)

	c := ret.MustGet(test_sandbox.VarChainOwner)

	require.EqualValues(t, chain.OriginatorAgentID.Bytes(), c)
}

func TestChainID(t *testing.T) {
	_, chain := setupChain(t, nil)
	setupTestSandboxSC(t, chain, nil)

	ret, err := chain.CallView(SandboxSCName, test_sandbox.FuncChainID)
	require.NoError(t, err)

	require.EqualValues(t, chain.ChainID.Bytes(), ret.MustGet(test_sandbox.VarChainID))
}

func TestSandboxCall(t *testing.T) {
	_, chain := setupChain(t, nil)
	setupTestSandboxSC(t, chain, nil)

	ret, err := chain.CallView(SandboxSCName, test_sandbox.FuncSandboxCall)
	require.NoError(t, err)

	d := ret.MustGet(test_sandbox.VarSandboxCall)
	require.EqualValues(t, "'solo' testing chain", string(d))
}
