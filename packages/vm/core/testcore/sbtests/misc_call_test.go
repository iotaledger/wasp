package sbtests

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/core/testcore/sbtests/sbtestsc"
)

func TestChainAdminView(t *testing.T) {
	_, chain := setupChain(t)
	setupTestSandboxSC(t, chain, nil)

	chainOwnderID, err := sbtestsc.FuncChainAdminView.Call(func(msg isc.Message) (isc.CallArguments, error) {
		return chain.CallViewWithContract(ScName, msg)
	})
	require.NoError(t, err)
	require.EqualValues(t, chain.AdminAgentID().Bytes(), chainOwnderID.Bytes())
}

func TestChainAdminFull(t *testing.T) {
	_, chain := setupChain(t)
	setupTestSandboxSC(t, chain, nil)

	chainOwnderID, err := sbtestsc.FuncChainAdminFull.Call(func(msg isc.Message) (isc.CallArguments, error) {
		req := solo.NewCallParams(msg, ScName).
			WithGasBudget(100_000)
		return chain.PostRequestSync(req, nil)
	})
	require.NoError(t, err)
	require.True(t, chain.AdminAgentID().Equals(chainOwnderID))
}

func TestSandboxCall(t *testing.T) {
	_, chain := setupChain(t)
	setupTestSandboxSC(t, chain, nil)

	ret, err := chain.CallViewEx(ScName, sbtestsc.FuncSandboxCall.Name)
	require.NoError(t, err)
	require.NotNil(t, ret)
}

func TestCustomError(t *testing.T) {
	_, chain := setupChain(t)
	setupTestSandboxSC(t, chain, nil)

	req := solo.NewCallParamsEx(ScName, sbtestsc.FuncTestCustomError.Name).
		WithGasBudget(100_000)
	ret, err := chain.PostRequestSync(req, nil)

	require.Error(t, err)
	require.IsType(t, &isc.VMError{}, err)
	require.Nil(t, ret)
}
