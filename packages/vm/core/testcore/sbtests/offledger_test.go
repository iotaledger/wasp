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

// TODO rewrite off-ledger tests because the fee concept changed

//func TestOffLedgerFeesEnough(t *testing.T) {
//	run2(t, func(t *testing.T, w bool) {
//		env, chain := setupChain(t, nil)
//		cAID, extraToken := setupTestSandboxSC(t, chain, nil, w)
//		user, userAddr, userAgentID := setupDeployer(t, chain)
//
//		req := solo.NewCallParams(governance.Contract.Name, governance.FuncSetContractFee.Name,
//			root.ParamHname, HScName,
//			governance.ParamOwnerFee, 10)
//		_, err := chain.PostRequestSync(req.AddIotas(1), nil)
//		require.NoError(t, err)
//
//		chain.AssertL2Iotas(userAgentID, 0)
//		chain.AssertL2Iotas(cAID, 1)
//
//		req = solo.NewCallParams(accounts.Contract.Name, accounts.FuncDeposit.Name)
//		_, err = chain.PostRequestSync(req.AddIotas(10), user)
//		require.NoError(t, err)
//
//		chain.AssertL2Iotas(userAgentID, 10)
//		chain.AssertL2Iotas(cAID, 1)
//
//		req = solo.NewCallParams(ScName, sbtestsc.FuncDoNothing.Name)
//		_, err = chain.PostRequestOffLedger(req.AddIotas(10), user)
//		require.NoError(t, err)
//
//		t.Logf("dump accounts:\n%s", chain.DumpAccounts())
//		chain.AssertL2Iotas(chain.OriginatorAgentID, 0)
//		chain.AssertL2Iotas(userAgentID, 0)
//		chain.AssertL2Iotas(cAID, 1)
//		env.AssertAddressIotas(userAddr, utxodb.FundsFromFaucetAmount-10)
//		env.AssertAddressIotas(chain.OriginatorAddress, utxodb.FundsFromFaucetAmount-solo.ChainDustThreshold-5-extraToken)
//		chain.AssertCommonAccountIotas(4 + extraToken + 10)
//	})
//}
//
//func TestOffLedgerFeesNotEnough(t *testing.T) {
//	run2(t, func(t *testing.T, w bool) {
//		env, chain := setupChain(t, nil)
//		cAID, extraToken := setupTestSandboxSC(t, chain, nil, w)
//		user, userAddr, userAgentID := setupDeployer(t, chain)
//
//		req := solo.NewCallParams(governance.Contract.Name, governance.FuncSetContractFee.Name,
//			root.ParamHname, HScName,
//			governance.ParamOwnerFee, 10)
//		_, err := chain.PostRequestSync(req.AddIotas(1), nil)
//		require.NoError(t, err)
//
//		chain.AssertL2Iotas(userAgentID, 0)
//		chain.AssertL2Iotas(cAID, 1)
//
//		req = solo.NewCallParams(accounts.Contract.Name, accounts.FuncDeposit.Name)
//		_, err = chain.PostRequestSync(req.AddIotas(9), user)
//		require.NoError(t, err)
//
//		chain.AssertL2Iotas(userAgentID, 9)
//		chain.AssertL2Iotas(cAID, 1)
//
//		req = solo.NewCallParams(ScName, sbtestsc.FuncDoNothing.Name)
//		_, err = chain.PostRequestOffLedger(req.AddIotas(10), user)
//		require.Error(t, err)
//		require.Contains(t, err.Error(), "not enough fees")
//
//		t.Logf("dump accounts:\n%s", chain.DumpAccounts())
//		chain.AssertL2Iotas(chain.OriginatorAgentID, 0)
//		chain.AssertL2Iotas(userAgentID, 0)
//		chain.AssertL2Iotas(cAID, 1)
//		env.AssertAddressIotas(userAddr, utxodb.FundsFromFaucetAmount-9)
//		env.AssertAddressIotas(chain.OriginatorAddress, utxodb.FundsFromFaucetAmount-solo.ChainDustThreshold-5-extraToken)
//		chain.AssertCommonAccountIotas(4 + extraToken + 9)
//	})
//}
//
//func TestOffLedgerFeesExtra(t *testing.T) {
//	run2(t, func(t *testing.T, w bool) {
//		env, chain := setupChain(t, nil)
//		cAID, extraToken := setupTestSandboxSC(t, chain, nil, w)
//		user, userAddr, userAgentID := setupDeployer(t, chain)
//
//		req := solo.NewCallParams(governance.Contract.Name, governance.FuncSetContractFee.Name,
//			root.ParamHname, HScName,
//			governance.ParamOwnerFee, 10)
//		_, err := chain.PostRequestSync(req.AddIotas(1), nil)
//		require.NoError(t, err)
//
//		chain.AssertL2Iotas(userAgentID, 0)
//		chain.AssertL2Iotas(cAID, 1)
//
//		req = solo.NewCallParams(accounts.Contract.Name, accounts.FuncDeposit.Name)
//		_, err = chain.PostRequestSync(req.AddIotas(11), user)
//		require.NoError(t, err)
//
//		chain.AssertL2Iotas(userAgentID, 11)
//		chain.AssertL2Iotas(cAID, 1)
//
//		req = solo.NewCallParams(ScName, sbtestsc.FuncDoNothing.Name)
//		_, err = chain.PostRequestOffLedger(req.AddIotas(10), user)
//		require.NoError(t, err)
//
//		t.Logf("dump accounts:\n%s", chain.DumpAccounts())
//		chain.AssertL2Iotas(chain.OriginatorAgentID, 0)
//		chain.AssertL2Iotas(userAgentID, 1)
//		chain.AssertL2Iotas(cAID, 1)
//		env.AssertAddressIotas(userAddr, utxodb.FundsFromFaucetAmount-11)
//		env.AssertAddressIotas(chain.OriginatorAddress, utxodb.FundsFromFaucetAmount-solo.ChainDustThreshold-5-extraToken)
//		chain.AssertCommonAccountIotas(4 + extraToken + 10)
//	})
//}
//
//func TestOffLedgerTransferWithFeesEnough(t *testing.T) {
//	run2(t, func(t *testing.T, w bool) {
//		env, chain := setupChain(t, nil)
//		cAID, extraToken := setupTestSandboxSC(t, chain, nil, w)
//		user, userAddr, userAgentID := setupDeployer(t, chain)
//
//		req := solo.NewCallParams(governance.Contract.Name, governance.FuncSetContractFee.Name,
//			root.ParamHname, HScName,
//			governance.ParamOwnerFee, 10)
//		_, err := chain.PostRequestSync(req.AddIotas(1), nil)
//		require.NoError(t, err)
//
//		chain.AssertL2Iotas(userAgentID, 0)
//		chain.AssertL2Iotas(cAID, 1)
//
//		req = solo.NewCallParams(accounts.Contract.Name, accounts.FuncDeposit.Name)
//		_, err = chain.PostRequestSync(req.AddIotas(10+42), user)
//		require.NoError(t, err)
//
//		chain.AssertL2Iotas(userAgentID, 10+42)
//		chain.AssertL2Iotas(cAID, 1)
//
//		req = solo.NewCallParams(ScName, sbtestsc.FuncDoNothing.Name)
//		_, err = chain.PostRequestOffLedger(req.AddIotas(10+42), user)
//		require.NoError(t, err)
//
//		t.Logf("dump accounts:\n%s", chain.DumpAccounts())
//		chain.AssertL2Iotas(chain.OriginatorAgentID, 0)
//		chain.AssertL2Iotas(userAgentID, 0)
//		chain.AssertL2Iotas(cAID, 1+42)
//		env.AssertAddressIotas(userAddr, utxodb.FundsFromFaucetAmount-10-42)
//		env.AssertAddressIotas(chain.OriginatorAddress, utxodb.FundsFromFaucetAmount-solo.ChainDustThreshold-5-extraToken)
//		chain.AssertCommonAccountIotas(4 + extraToken + 10)
//	})
//}
//
//func TestOffLedgerTransferWithFeesNotEnough(t *testing.T) {
//	run2(t, func(t *testing.T, w bool) {
//		env, chain := setupChain(t, nil)
//		cAID, extraToken := setupTestSandboxSC(t, chain, nil, w)
//		user, userAddr, userAgentID := setupDeployer(t, chain)
//
//		req := solo.NewCallParams(governance.Contract.Name, governance.FuncSetContractFee.Name,
//			root.ParamHname, HScName,
//			governance.ParamOwnerFee, 10)
//		_, err := chain.PostRequestSync(req.AddIotas(1), nil)
//		require.NoError(t, err)
//
//		chain.AssertL2Iotas(userAgentID, 0)
//		chain.AssertL2Iotas(cAID, 1)
//
//		req = solo.NewCallParams(accounts.Contract.Name, accounts.FuncDeposit.Name)
//		_, err = chain.PostRequestSync(req.AddIotas(10+41), user)
//		require.NoError(t, err)
//
//		chain.AssertL2Iotas(userAgentID, 10+41)
//		chain.AssertL2Iotas(cAID, 1)
//
//		req = solo.NewCallParams(ScName, sbtestsc.FuncDoNothing.Name)
//		_, err = chain.PostRequestOffLedger(req.AddIotas(10+42), user)
//		require.NoError(t, err)
//
//		t.Logf("dump accounts:\n%s", chain.DumpAccounts())
//		chain.AssertL2Iotas(chain.OriginatorAgentID, 0)
//		chain.AssertL2Iotas(userAgentID, 0)
//		chain.AssertL2Iotas(cAID, 1+41)
//		env.AssertAddressIotas(userAddr, utxodb.FundsFromFaucetAmount-10-41)
//		env.AssertAddressIotas(chain.OriginatorAddress, utxodb.FundsFromFaucetAmount-solo.ChainDustThreshold-5-extraToken)
//		chain.AssertCommonAccountIotas(4 + extraToken + 10)
//	})
//}
//
//func TestOffLedgerTransferWithFeesExtra(t *testing.T) {
//	run2(t, func(t *testing.T, w bool) {
//		env, chain := setupChain(t, nil)
//		cAID, extraToken := setupTestSandboxSC(t, chain, nil, w)
//		user, userAddr, userAgentID := setupDeployer(t, chain)
//
//		req := solo.NewCallParams(governance.Contract.Name, governance.FuncSetContractFee.Name,
//			root.ParamHname, HScName,
//			governance.ParamOwnerFee, 10)
//		_, err := chain.PostRequestSync(req.AddIotas(1), nil)
//		require.NoError(t, err)
//
//		chain.AssertL2Iotas(userAgentID, 0)
//		chain.AssertL2Iotas(cAID, 1)
//
//		req = solo.NewCallParams(accounts.Contract.Name, accounts.FuncDeposit.Name)
//		_, err = chain.PostRequestSync(req.AddIotas(10+43), user)
//		require.NoError(t, err)
//
//		chain.AssertL2Iotas(userAgentID, 10+43)
//		chain.AssertL2Iotas(cAID, 1)
//
//		req = solo.NewCallParams(ScName, sbtestsc.FuncDoNothing.Name)
//		_, err = chain.PostRequestOffLedger(req.AddIotas(10+42), user)
//		require.NoError(t, err)
//
//		t.Logf("dump accounts:\n%s", chain.DumpAccounts())
//		chain.AssertL2Iotas(chain.OriginatorAgentID, 0)
//		chain.AssertL2Iotas(userAgentID, 1)
//		chain.AssertL2Iotas(cAID, 1+42)
//		env.AssertAddressIotas(userAddr, utxodb.FundsFromFaucetAmount-10-43)
//		env.AssertAddressIotas(chain.OriginatorAddress, utxodb.FundsFromFaucetAmount-solo.ChainDustThreshold-5-extraToken)
//		chain.AssertCommonAccountIotas(4 + extraToken + 10)
//	})
//}
//
//func TestOffLedgerTransferEnough(t *testing.T) {
//	run2(t, func(t *testing.T, w bool) {
//		env, chain := setupChain(t, nil)
//		cAID, extraToken := setupTestSandboxSC(t, chain, nil, w)
//		user, userAddr, userAgentID := setupDeployer(t, chain)
//
//		chain.AssertL2Iotas(userAgentID, 0)
//		chain.AssertL2Iotas(cAID, 1)
//
//		req := solo.NewCallParams(accounts.Contract.Name, accounts.FuncDeposit.Name)
//		_, err := chain.PostRequestSync(req.AddIotas(42), user)
//		require.NoError(t, err)
//
//		chain.AssertL2Iotas(userAgentID, 42)
//		chain.AssertL2Iotas(cAID, 1)
//
//		req = solo.NewCallParams(ScName, sbtestsc.FuncDoNothing.Name)
//		_, err = chain.PostRequestOffLedger(req.AddIotas(42), user)
//		require.NoError(t, err)
//
//		t.Logf("dump accounts:\n%s", chain.DumpAccounts())
//		chain.AssertL2Iotas(chain.OriginatorAgentID, 0)
//		chain.AssertL2Iotas(userAgentID, 0)
//		chain.AssertL2Iotas(cAID, 1+42)
//		env.AssertAddressIotas(userAddr, utxodb.FundsFromFaucetAmount-42)
//		env.AssertAddressIotas(chain.OriginatorAddress, utxodb.FundsFromFaucetAmount-solo.ChainDustThreshold-4-extraToken)
//		chain.AssertCommonAccountIotas(3 + extraToken)
//	})
//}
//
//func TestOffLedgerTransferNotEnough(t *testing.T) {
//	run2(t, func(t *testing.T, w bool) {
//		env, chain := setupChain(t, nil)
//		cAID, extraToken := setupTestSandboxSC(t, chain, nil, w)
//		user, userAddr, userAgentID := setupDeployer(t, chain)
//
//		chain.AssertL2Iotas(userAgentID, 0)
//		chain.AssertL2Iotas(cAID, 1)
//
//		req := solo.NewCallParams(accounts.Contract.Name, accounts.FuncDeposit.Name)
//		_, err := chain.PostRequestSync(req.AddIotas(41), user)
//		require.NoError(t, err)
//
//		chain.AssertL2Iotas(userAgentID, 41)
//		chain.AssertL2Iotas(cAID, 1)
//
//		req = solo.NewCallParams(ScName, sbtestsc.FuncDoNothing.Name)
//		_, err = chain.PostRequestOffLedger(req.AddIotas(42), user)
//		require.NoError(t, err)
//
//		t.Logf("dump accounts:\n%s", chain.DumpAccounts())
//		chain.AssertL2Iotas(chain.OriginatorAgentID, 0)
//		chain.AssertL2Iotas(userAgentID, 0)
//		chain.AssertL2Iotas(cAID, 1+41)
//		env.AssertAddressIotas(userAddr, utxodb.FundsFromFaucetAmount-41)
//		env.AssertAddressIotas(chain.OriginatorAddress, utxodb.FundsFromFaucetAmount-solo.ChainDustThreshold-4-extraToken)
//		chain.AssertCommonAccountIotas(3 + extraToken)
//	})
//}
//
//func TestOffLedgerTransferExtra(t *testing.T) {
//	run2(t, func(t *testing.T, w bool) {
//		env, chain := setupChain(t, nil)
//		cAID, extraToken := setupTestSandboxSC(t, chain, nil, w)
//		user, userAddr, userAgentID := setupDeployer(t, chain)
//
//		chain.AssertL2Iotas(userAgentID, 0)
//		chain.AssertL2Iotas(cAID, 1)
//
//		req := solo.NewCallParams(accounts.Contract.Name, accounts.FuncDeposit.Name)
//		_, err := chain.PostRequestSync(req.AddIotas(43), user)
//		require.NoError(t, err)
//
//		chain.AssertL2Iotas(userAgentID, 43)
//		chain.AssertL2Iotas(cAID, 1)
//
//		req = solo.NewCallParams(ScName, sbtestsc.FuncDoNothing.Name)
//		_, err = chain.PostRequestOffLedger(req.AddIotas(42), user)
//		require.NoError(t, err)
//
//		t.Logf("dump accounts:\n%s", chain.DumpAccounts())
//		chain.AssertL2Iotas(chain.OriginatorAgentID, 0)
//		chain.AssertL2Iotas(userAgentID, 1)
//		chain.AssertL2Iotas(cAID, 1+42)
//		env.AssertAddressIotas(userAddr, utxodb.FundsFromFaucetAmount-43)
//		env.AssertAddressIotas(chain.OriginatorAddress, utxodb.FundsFromFaucetAmount-solo.ChainDustThreshold-4-extraToken)
//		chain.AssertCommonAccountIotas(3 + extraToken)
//	})
//}
