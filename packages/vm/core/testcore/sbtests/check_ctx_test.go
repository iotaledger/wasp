package sbtests

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/core/testcore/sbtests/sbtestsc"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

func TestMainCallsFromFullEP(t *testing.T) {
	_, chain := setupChain(t, nil)

	user, userAgentID := setupDeployer(t, chain)

	setupTestSandboxSC(t, chain, user)

	req := solo.NewCallParams(sbtestsc.FuncCheckContextFromFullEP.Message(
		chain.AdminAgentID(),
		userAgentID,
		isc.NewContractAgentID(HScName),
	), ScName).
		WithGasBudget(10 * gas.LimitsDefault.MinGasPerRequest)
	_, err := chain.PostRequestSync(req, user)
	require.NoError(t, err)
}

func TestMainCallsFromViewEP(t *testing.T) {
	_, chain := setupChain(t, nil)

	user, _ := setupDeployer(t, chain)

	setupTestSandboxSC(t, chain, user)

	err := sbtestsc.FuncCheckContextFromViewEP.Call(
		chain.AdminAgentID(),
		isc.NewContractAgentID(HScName),
		func(msg isc.Message) (isc.CallArguments, error) {
			return chain.CallViewWithContract(ScName, msg)
		},
	)
	require.NoError(t, err)
}
