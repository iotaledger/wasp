package sbtests

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/solo"
	"github.com/iotaledger/wasp/v2/packages/vm/core/testcore/sbtests/sbtestsc"
)

func TestMainCallsFromFullEP(t *testing.T) {
	_, chain := setupChain(t)

	user, userAgentID := setupDeployer(t, chain)

	setupTestSandboxSC(t, chain, user)

	req := solo.NewCallParams(sbtestsc.FuncCheckContextFromFullEP.Message(
		chain.AdminAgentID(),
		userAgentID,
		isc.NewContractAgentID(HScName),
	), ScName)
	_, err := chain.PostRequestSync(req, user)
	require.NoError(t, err)
}

func TestMainCallsFromViewEP(t *testing.T) {
	_, chain := setupChain(t)

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
