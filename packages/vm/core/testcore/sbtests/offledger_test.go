package sbtests

import (
	"testing"

	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/kvdecoder"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/testutil/testmisc"
	"github.com/iotaledger/wasp/packages/vm/core/testcore/sbtests/sbtestsc"
	"github.com/stretchr/testify/require"
)

func TestOffLedgerFailNoAccount(t *testing.T) {
	run2(t, func(t *testing.T, w bool) {
		t.SkipNow() // TODO EMPTY BLOCKS NOT SUPPORTED IN SOLO

		env, chain := setupChain(t, nil)
		cAID := setupTestSandboxSC(t, chain, nil, w)

		user, userAddr := env.NewKeyPairWithFunds()
		userAgentID := iscp.NewAgentID(userAddr)

		chain.AssertL2Iotas(userAgentID, 0)
		chain.AssertL2Iotas(cAID, 0)

		req := solo.NewCallParams(ScName, sbtestsc.FuncSetInt.Name,
			sbtestsc.ParamIntParamName, "ppp",
			sbtestsc.ParamIntParamValue, 314,
		)
		_, err := chain.PostRequestOffLedger(req, user)
		require.Error(t, err)
		testmisc.RequireErrorToBe(t, err, "unverified account")

		chain.AssertL2Iotas(userAgentID, 0)
		chain.AssertL2Iotas(cAID, 0)
	})
}

func TestOffLedgerSuccess(t *testing.T) {
	run2(t, func(t *testing.T, w bool) {
		env, ch := setupChain(t, nil)
		cAID := setupTestSandboxSC(t, ch, nil, w)

		user, userAddr := env.NewKeyPairWithFunds()
		userAgentID := iscp.NewAgentID(userAddr)

		ch.AssertL2Iotas(userAgentID, 0)
		ch.AssertL2Iotas(cAID, 0)

		depositIotas := 1 * iscp.Mi
		err := ch.DepositIotasToL2(depositIotas, user)
		expectedUser := depositIotas - ch.LastReceipt().GasFeeCharged
		ch.AssertL2Iotas(userAgentID, expectedUser)
		require.NoError(t, err)

		req := solo.NewCallParams(ScName, sbtestsc.FuncSetInt.Name,
			sbtestsc.ParamIntParamName, "ppp",
			sbtestsc.ParamIntParamValue, 314,
		).WithGasBudget(100_000)
		_, err = ch.PostRequestOffLedger(req, user)
		require.NoError(t, err)
		rec := ch.LastReceipt()
		require.NoError(t, rec.Error.AsGoError())
		t.Logf("receipt: %s", rec)

		res, err := ch.CallView(ScName, sbtestsc.FuncGetInt.Name,
			sbtestsc.ParamIntParamName, "ppp",
		)
		require.NoError(t, err)
		require.EqualValues(t, 314, kvdecoder.New(res).MustGetUint64("ppp"))
		ch.AssertL2Iotas(userAgentID, expectedUser-rec.GasFeeCharged)
	})
}
