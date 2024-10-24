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

	chainOwnderID, err := sbtestsc.FuncChainOwnerIDView.Call(func(msg isc.Message) (isc.CallArguments, error) {
		return chain.CallViewWithContract(ScName, msg)
	})
	require.NoError(t, err)
	require.EqualValues(t, chain.OriginatorAgentID.Bytes(), chainOwnderID.Bytes())
}

func TestChainOwnerIDFull(t *testing.T) { run2(t, testChainOwnerIDFull) }
func testChainOwnerIDFull(t *testing.T) {
	_, chain := setupChain(t, nil)
	setupTestSandboxSC(t, chain, nil)

	chainOwnderID, err := sbtestsc.FuncChainOwnerIDFull.Call(func(msg isc.Message) (isc.CallArguments, error) {
		req := solo.NewCallParams(msg, ScName).
			WithGasBudget(100_000)
		return chain.PostRequestSync(req, nil)
	})
	require.NoError(t, err)
	require.EqualValues(t, chain.OriginatorAgentID.Bytes(), chainOwnderID)
}

func TestSandboxCall(t *testing.T) { run2(t, testSandboxCall) }
func testSandboxCall(t *testing.T) {
	_, chain := setupChain(t, nil)
	setupTestSandboxSC(t, chain, nil)

	ret, err := chain.CallViewEx(ScName, sbtestsc.FuncSandboxCall.Name)
	require.NoError(t, err)
	require.NotNil(t, ret)
}

func TestCustomError(t *testing.T) { run2(t, testCustomError) }
func testCustomError(t *testing.T) {
	_, chain := setupChain(t, nil)
	setupTestSandboxSC(t, chain, nil)

	req := solo.NewCallParamsEx(ScName, sbtestsc.FuncTestCustomError.Name).
		WithGasBudget(100_000)
	ret, err := chain.PostRequestSync(req, nil)

	require.Error(t, err)
	require.IsType(t, &isc.VMError{}, err)
	require.Nil(t, ret)
}
