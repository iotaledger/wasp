package sbtests

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/core/testcore/sbtests/sbtestsc"
)

func TestChainOwnerIDView(t *testing.T) { run2(t, testChainOwnerIDView) }
func testChainOwnerIDView(t *testing.T) {
	_, chain := setupChain(t, nil)
	setupTestSandboxSC(t, chain, nil)

	ret, err := chain.CallView(ScName, sbtestsc.FuncChainOwnerIDView.Name)
	require.NoError(t, err)

	c := ret.Get(sbtestsc.ParamChainOwnerID)

	require.EqualValues(t, chain.OriginatorAgentID.Bytes(), c)
}

func TestChainOwnerIDFull(t *testing.T) { run2(t, testChainOwnerIDFull) }
func testChainOwnerIDFull(t *testing.T) {
	_, chain := setupChain(t, nil)
	setupTestSandboxSC(t, chain, nil)

	req := solo.NewCallParams(ScName, sbtestsc.FuncChainOwnerIDFull.Name).
		WithGasBudget(100_000)
	ret, err := chain.PostRequestSync(req, nil)
	require.NoError(t, err)

	c := ret.Get(sbtestsc.ParamChainOwnerID)
	require.EqualValues(t, chain.OriginatorAgentID.Bytes(), c)
}

func TestSandboxCall(t *testing.T) { run2(t, testSandboxCall) }
func testSandboxCall(t *testing.T) {
	_, chain := setupChain(t, nil)
	setupTestSandboxSC(t, chain, nil)

	ret, err := chain.CallView(ScName, sbtestsc.FuncSandboxCall.Name)
	require.NoError(t, err)
	require.NotNil(t, ret)
}

func TestCustomError(t *testing.T) { run2(t, testCustomError) }
func testCustomError(t *testing.T) {
	_, chain := setupChain(t, nil)
	setupTestSandboxSC(t, chain, nil)

	req := solo.NewCallParams(ScName, sbtestsc.FuncTestCustomError.Name).
		WithGasBudget(100_000)
	ret, err := chain.PostRequestSync(req, nil)

	require.Error(t, err)
	require.IsType(t, &isc.VMError{}, err)
	require.Nil(t, ret)
}
