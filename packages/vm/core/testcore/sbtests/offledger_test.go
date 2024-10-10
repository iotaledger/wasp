package sbtests

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/kvdecoder"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/testutil/testmisc"
	"github.com/iotaledger/wasp/packages/vm/core/testcore/sbtests/sbtestsc"
)

func TestOffLedgerFailNoAccount(t *testing.T) {
	run2(t, func(t *testing.T) {
		t.SkipNow() // TODO EMPTY BLOCKS NOT SUPPORTED IN SOLO

		env, chain := setupChain(t, nil)
		cAID := setupTestSandboxSC(t, chain, nil)

		user, userAddr := env.NewKeyPairWithFunds()
		userAgentID := isc.NewAddressAgentID(userAddr)

		chain.AssertL2BaseTokens(userAgentID, 0)
		chain.AssertL2BaseTokens(cAID, 0)

		req := solo.NewCallParamsEx(ScName, sbtestsc.FuncSetInt.Name,
			sbtestsc.ParamIntParamName, "ppp",
			sbtestsc.ParamIntParamValue, 314,
		)
		_, err := chain.PostRequestOffLedger(req, user)
		require.Error(t, err)
		testmisc.RequireErrorToBe(t, err, "unverified account")

		chain.AssertL2BaseTokens(userAgentID, 0)
		chain.AssertL2BaseTokens(cAID, 0)
	})
}

func TestOffLedgerSuccess(t *testing.T) {
	run2(t, func(t *testing.T) {
		env, ch := setupChain(t, nil)
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

		req := solo.NewCallParamsEx(ScName, sbtestsc.FuncSetInt.Name,
			sbtestsc.ParamIntParamName, "ppp",
			sbtestsc.ParamIntParamValue, 314,
		).WithGasBudget(100_000)
		_, err = ch.PostRequestOffLedger(req, user)
		require.NoError(t, err)
		rec := ch.LastReceipt()
		require.NoError(t, rec.Error.AsGoError())
		t.Logf("receipt: %s", rec)

		res, err := ch.CallViewEx(ScName, sbtestsc.FuncGetInt.Name,
			sbtestsc.ParamIntParamName, "ppp",
		)
		require.NoError(t, err)
		require.EqualValues(t, 314, kvdecoder.New(res).MustGetUint64("ppp"))
		ch.AssertL2BaseTokens(userAgentID, expectedUser-rec.GasFeeCharged)
	})
}
