package sbtests

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/testutil/testmisc"
	"github.com/iotaledger/wasp/packages/vm/core/testcore/sbtests/sbtestsc"
)

func TestOffLedgerFailNoAccount(t *testing.T) {
	t.SkipNow() // TODO EMPTY BLOCKS NOT SUPPORTED IN SOLO

	env, chain := setupChain(t)
	cAID := setupTestSandboxSC(t, chain, nil)

	user, userAddr := env.NewKeyPairWithFunds()
	userAgentID := isc.NewAddressAgentID(userAddr)

	chain.AssertL2BaseTokens(userAgentID, 0)
	chain.AssertL2BaseTokens(cAID, 0)

	msg := sbtestsc.FuncSetInt.Message("ppp", 314)

	_, err := chain.PostRequestOffLedger(solo.NewCallParams(msg, ScName), user)
	require.Error(t, err)
	testmisc.RequireErrorToBe(t, err, "unverified account")

	chain.AssertL2BaseTokens(userAgentID, 0)
	chain.AssertL2BaseTokens(cAID, 0)
}

func TestOffLedgerSuccess(t *testing.T) {
	env, ch := setupChain(t)
	cAID := setupTestSandboxSC(t, ch, nil)

	user, userAddr := env.NewKeyPairWithFunds()
	userAgentID := isc.NewAddressAgentID(userAddr)

	ch.AssertL2BaseTokens(userAgentID, 0)
	ch.AssertL2BaseTokens(cAID, 0)

	var depositBaseTokens coin.Value = 1 * isc.Million
	err := ch.DepositBaseTokensToL2(depositBaseTokens, user)
	expectedUser := depositBaseTokens - ch.LastReceipt().GasFeeCharged
	ch.AssertL2BaseTokens(userAgentID, expectedUser)
	require.NoError(t, err)

	req := solo.NewCallParams(sbtestsc.FuncSetInt.Message("ppp", 314), ScName).
		WithGasBudget(100_000)
	_, err = ch.PostRequestOffLedger(req, user)
	require.NoError(t, err)
	rec := ch.LastReceipt()
	require.NoError(t, rec.Error.AsGoError())
	t.Logf("receipt: %s", rec)

	r, err := sbtestsc.FuncGetInt.Call("ppp", func(msg isc.Message) (isc.CallArguments, error) {
		return ch.CallViewWithContract(ScName, msg)
	})
	require.NoError(t, err)
	require.EqualValues(t, 314, r)
	ch.AssertL2BaseTokens(userAgentID, expectedUser-rec.GasFeeCharged)
}
