package sbtests

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/core/testcore/sbtests/sbtestsc"
)

func TestMainCallsFromFullEP(t *testing.T) { run2(t, testMainCallsFromFullEP) }
func testMainCallsFromFullEP(t *testing.T, w bool) {
	_, chain := setupChain(t, nil)

	user, userAgentID := setupDeployer(t, chain)

	setupTestSandboxSC(t, chain, user, w)

	req := solo.NewCallParams(ScName, sbtestsc.FuncCheckContextFromFullEP.Name,
		sbtestsc.ParamChainID, chain.ChainID,
		sbtestsc.ParamAgentID, isc.NewContractAgentID(chain.ChainID, HScName),
		sbtestsc.ParamCaller, userAgentID,
		sbtestsc.ParamChainOwnerID, chain.OriginatorAgentID,
	).
		WithGasBudget(120_000)
	_, err := chain.PostRequestSync(req, user)
	require.NoError(t, err)
}

func TestMainCallsFromViewEP(t *testing.T) { run2(t, testMainCallsFromViewEP) }
func testMainCallsFromViewEP(t *testing.T, w bool) {
	_, chain := setupChain(t, nil)

	user, _ := setupDeployer(t, chain)

	setupTestSandboxSC(t, chain, user, w)

	_, err := chain.CallView(ScName, sbtestsc.FuncCheckContextFromViewEP.Name,
		sbtestsc.ParamChainID, chain.ChainID,
		sbtestsc.ParamAgentID, isc.NewContractAgentID(chain.ChainID, HScName),
		sbtestsc.ParamChainOwnerID, chain.OriginatorAgentID,
	)
	require.NoError(t, err)
}
