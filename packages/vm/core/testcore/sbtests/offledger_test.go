//nolint:dupl
package sbtests

import (
	"strings"
	"testing"

	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/packages/vm/core/testcore/sbtests/sbtestsc"
	"github.com/stretchr/testify/require"
)

func TestOffLedgerFailNoAccount(t *testing.T) {
	run2(t, func(t *testing.T, w bool) {
		env, chain := setupChain(t, nil)
		cAID, _ := setupTestSandboxSC(t, chain, nil, w)

		owner, ownerAddr := env.NewKeyPairWithFunds()
		ownerAgentID := iscp.NewAgentID(ownerAddr, 0)

		chain.AssertL2AccountIotas(ownerAgentID, 0)
		chain.AssertL2AccountIotas(cAID, 1)

		// NOTE: NO deposit into owner account
		//req := solo.NewCallParams(accounts.Interface.Name, accounts.FuncDeposit)
		//_, err := chain.PostRequestSync(req.WithIotas(10), owner)
		//require.NoError(t, err)

		chain.AssertL2AccountIotas(ownerAgentID, 0)
		chain.AssertL2AccountIotas(cAID, 1)

		req := solo.NewCallParams(ScName, sbtestsc.FuncSetInt.Name,
			sbtestsc.ParamIntParamName, "ppp",
			sbtestsc.ParamIntParamValue, 314,
		)
		_, err := chain.PostRequestOffLedger(req, owner)
		require.Error(t, err)
		require.True(t, strings.Contains(err.Error(), "unverified account"))

		chain.AssertL2AccountIotas(ownerAgentID, 0)
		chain.AssertL2AccountIotas(cAID, 1)
	})
}

func TestOffLedgerNoFeeNoTransfer(t *testing.T) {
	run2(t, func(t *testing.T, w bool) {
		env, chain := setupChain(t, nil)
		cAID, _ := setupTestSandboxSC(t, chain, nil, w)

		owner, ownerAddr := env.NewKeyPairWithFunds()
		ownerAgentID := iscp.NewAgentID(ownerAddr, 0)

		chain.AssertL2AccountIotas(ownerAgentID, 0)
		chain.AssertL2AccountIotas(cAID, 1)

		// deposit into owner account
		req := solo.NewCallParams(accounts.Contract.Name, accounts.FuncDeposit.Name)
		_, err := chain.PostRequestSync(req.WithIotas(10), owner)
		require.NoError(t, err)

		chain.AssertL2AccountIotas(ownerAgentID, 10)
		chain.AssertL2AccountIotas(cAID, 1)

		req = solo.NewCallParams(ScName, sbtestsc.FuncSetInt.Name,
			sbtestsc.ParamIntParamName, "ppp",
			sbtestsc.ParamIntParamValue, 314,
		)
		// Look, Ma! No .WithIotas() necessary when doing off-ledger request!
		_, err = chain.PostRequestOffLedger(req, owner)
		require.NoError(t, err)

		chain.AssertL2AccountIotas(ownerAgentID, 10)
		chain.AssertL2AccountIotas(cAID, 1)

		ret, err := chain.CallView(ScName, sbtestsc.FuncGetInt.Name,
			sbtestsc.ParamIntParamName, "ppp")
		require.NoError(t, err)

		retInt, err := codec.DecodeInt64(ret.MustGet("ppp"))
		require.NoError(t, err)
		require.EqualValues(t, 314, retInt)
	})
}

func TestOffLedgerFeesEnough(t *testing.T) {
	run2(t, func(t *testing.T, w bool) {
		env, chain := setupChain(t, nil)
		cAID, extraToken := setupTestSandboxSC(t, chain, nil, w)
		user, userAddr, userAgentID := setupDeployer(t, chain)

		req := solo.NewCallParams(governance.Contract.Name, governance.FuncSetContractFee.Name,
			root.ParamHname, HScName,
			governance.ParamOwnerFee, 10)
		_, err := chain.PostRequestSync(req.WithIotas(1), nil)
		require.NoError(t, err)

		chain.AssertL2AccountIotas(userAgentID, 0)
		chain.AssertL2AccountIotas(cAID, 1)

		req = solo.NewCallParams(accounts.Contract.Name, accounts.FuncDeposit.Name)
		_, err = chain.PostRequestSync(req.WithIotas(10), user)
		require.NoError(t, err)

		chain.AssertL2AccountIotas(userAgentID, 10)
		chain.AssertL2AccountIotas(cAID, 1)

		req = solo.NewCallParams(ScName, sbtestsc.FuncDoNothing.Name)
		_, err = chain.PostRequestOffLedger(req.WithIotas(10), user)
		require.NoError(t, err)

		t.Logf("dump accounts:\n%s", chain.DumpAccounts())
		chain.AssertL2AccountIotas(chain.OriginatorAgentID, 0)
		chain.AssertL2AccountIotas(userAgentID, 0)
		chain.AssertL2AccountIotas(cAID, 1)
		env.AssertAddressIotas(userAddr, solo.Saldo-10)
		env.AssertAddressIotas(chain.OriginatorAddress, solo.Saldo-solo.ChainDustThreshold-5-extraToken)
		chain.AssertCommonAccountIotas(4 + extraToken + 10)
	})
}

func TestOffLedgerFeesNotEnough(t *testing.T) {
	run2(t, func(t *testing.T, w bool) {
		env, chain := setupChain(t, nil)
		cAID, extraToken := setupTestSandboxSC(t, chain, nil, w)
		user, userAddr, userAgentID := setupDeployer(t, chain)

		req := solo.NewCallParams(governance.Contract.Name, governance.FuncSetContractFee.Name,
			root.ParamHname, HScName,
			governance.ParamOwnerFee, 10)
		_, err := chain.PostRequestSync(req.WithIotas(1), nil)
		require.NoError(t, err)

		chain.AssertL2AccountIotas(userAgentID, 0)
		chain.AssertL2AccountIotas(cAID, 1)

		req = solo.NewCallParams(accounts.Contract.Name, accounts.FuncDeposit.Name)
		_, err = chain.PostRequestSync(req.WithIotas(9), user)
		require.NoError(t, err)

		chain.AssertL2AccountIotas(userAgentID, 9)
		chain.AssertL2AccountIotas(cAID, 1)

		req = solo.NewCallParams(ScName, sbtestsc.FuncDoNothing.Name)
		_, err = chain.PostRequestOffLedger(req.WithIotas(10), user)
		require.Error(t, err)
		require.Contains(t, err.Error(), "not enough fees")

		t.Logf("dump accounts:\n%s", chain.DumpAccounts())
		chain.AssertL2AccountIotas(chain.OriginatorAgentID, 0)
		chain.AssertL2AccountIotas(userAgentID, 0)
		chain.AssertL2AccountIotas(cAID, 1)
		env.AssertAddressIotas(userAddr, solo.Saldo-9)
		env.AssertAddressIotas(chain.OriginatorAddress, solo.Saldo-solo.ChainDustThreshold-5-extraToken)
		chain.AssertCommonAccountIotas(4 + extraToken + 9)
	})
}

func TestOffLedgerFeesExtra(t *testing.T) {
	run2(t, func(t *testing.T, w bool) {
		env, chain := setupChain(t, nil)
		cAID, extraToken := setupTestSandboxSC(t, chain, nil, w)
		user, userAddr, userAgentID := setupDeployer(t, chain)

		req := solo.NewCallParams(governance.Contract.Name, governance.FuncSetContractFee.Name,
			root.ParamHname, HScName,
			governance.ParamOwnerFee, 10)
		_, err := chain.PostRequestSync(req.WithIotas(1), nil)
		require.NoError(t, err)

		chain.AssertL2AccountIotas(userAgentID, 0)
		chain.AssertL2AccountIotas(cAID, 1)

		req = solo.NewCallParams(accounts.Contract.Name, accounts.FuncDeposit.Name)
		_, err = chain.PostRequestSync(req.WithIotas(11), user)
		require.NoError(t, err)

		chain.AssertL2AccountIotas(userAgentID, 11)
		chain.AssertL2AccountIotas(cAID, 1)

		req = solo.NewCallParams(ScName, sbtestsc.FuncDoNothing.Name)
		_, err = chain.PostRequestOffLedger(req.WithIotas(10), user)
		require.NoError(t, err)

		t.Logf("dump accounts:\n%s", chain.DumpAccounts())
		chain.AssertL2AccountIotas(chain.OriginatorAgentID, 0)
		chain.AssertL2AccountIotas(userAgentID, 1)
		chain.AssertL2AccountIotas(cAID, 1)
		env.AssertAddressIotas(userAddr, solo.Saldo-11)
		env.AssertAddressIotas(chain.OriginatorAddress, solo.Saldo-solo.ChainDustThreshold-5-extraToken)
		chain.AssertCommonAccountIotas(4 + extraToken + 10)
	})
}

func TestOffLedgerTransferWithFeesEnough(t *testing.T) {
	run2(t, func(t *testing.T, w bool) {
		env, chain := setupChain(t, nil)
		cAID, extraToken := setupTestSandboxSC(t, chain, nil, w)
		user, userAddr, userAgentID := setupDeployer(t, chain)

		req := solo.NewCallParams(governance.Contract.Name, governance.FuncSetContractFee.Name,
			root.ParamHname, HScName,
			governance.ParamOwnerFee, 10)
		_, err := chain.PostRequestSync(req.WithIotas(1), nil)
		require.NoError(t, err)

		chain.AssertL2AccountIotas(userAgentID, 0)
		chain.AssertL2AccountIotas(cAID, 1)

		req = solo.NewCallParams(accounts.Contract.Name, accounts.FuncDeposit.Name)
		_, err = chain.PostRequestSync(req.WithIotas(10+42), user)
		require.NoError(t, err)

		chain.AssertL2AccountIotas(userAgentID, 10+42)
		chain.AssertL2AccountIotas(cAID, 1)

		req = solo.NewCallParams(ScName, sbtestsc.FuncDoNothing.Name)
		_, err = chain.PostRequestOffLedger(req.WithIotas(10+42), user)
		require.NoError(t, err)

		t.Logf("dump accounts:\n%s", chain.DumpAccounts())
		chain.AssertL2AccountIotas(chain.OriginatorAgentID, 0)
		chain.AssertL2AccountIotas(userAgentID, 0)
		chain.AssertL2AccountIotas(cAID, 1+42)
		env.AssertAddressIotas(userAddr, solo.Saldo-10-42)
		env.AssertAddressIotas(chain.OriginatorAddress, solo.Saldo-solo.ChainDustThreshold-5-extraToken)
		chain.AssertCommonAccountIotas(4 + extraToken + 10)
	})
}

func TestOffLedgerTransferWithFeesNotEnough(t *testing.T) {
	run2(t, func(t *testing.T, w bool) {
		env, chain := setupChain(t, nil)
		cAID, extraToken := setupTestSandboxSC(t, chain, nil, w)
		user, userAddr, userAgentID := setupDeployer(t, chain)

		req := solo.NewCallParams(governance.Contract.Name, governance.FuncSetContractFee.Name,
			root.ParamHname, HScName,
			governance.ParamOwnerFee, 10)
		_, err := chain.PostRequestSync(req.WithIotas(1), nil)
		require.NoError(t, err)

		chain.AssertL2AccountIotas(userAgentID, 0)
		chain.AssertL2AccountIotas(cAID, 1)

		req = solo.NewCallParams(accounts.Contract.Name, accounts.FuncDeposit.Name)
		_, err = chain.PostRequestSync(req.WithIotas(10+41), user)
		require.NoError(t, err)

		chain.AssertL2AccountIotas(userAgentID, 10+41)
		chain.AssertL2AccountIotas(cAID, 1)

		req = solo.NewCallParams(ScName, sbtestsc.FuncDoNothing.Name)
		_, err = chain.PostRequestOffLedger(req.WithIotas(10+42), user)
		require.NoError(t, err)

		t.Logf("dump accounts:\n%s", chain.DumpAccounts())
		chain.AssertL2AccountIotas(chain.OriginatorAgentID, 0)
		chain.AssertL2AccountIotas(userAgentID, 0)
		chain.AssertL2AccountIotas(cAID, 1+41)
		env.AssertAddressIotas(userAddr, solo.Saldo-10-41)
		env.AssertAddressIotas(chain.OriginatorAddress, solo.Saldo-solo.ChainDustThreshold-5-extraToken)
		chain.AssertCommonAccountIotas(4 + extraToken + 10)
	})
}

func TestOffLedgerTransferWithFeesExtra(t *testing.T) {
	run2(t, func(t *testing.T, w bool) {
		env, chain := setupChain(t, nil)
		cAID, extraToken := setupTestSandboxSC(t, chain, nil, w)
		user, userAddr, userAgentID := setupDeployer(t, chain)

		req := solo.NewCallParams(governance.Contract.Name, governance.FuncSetContractFee.Name,
			root.ParamHname, HScName,
			governance.ParamOwnerFee, 10)
		_, err := chain.PostRequestSync(req.WithIotas(1), nil)
		require.NoError(t, err)

		chain.AssertL2AccountIotas(userAgentID, 0)
		chain.AssertL2AccountIotas(cAID, 1)

		req = solo.NewCallParams(accounts.Contract.Name, accounts.FuncDeposit.Name)
		_, err = chain.PostRequestSync(req.WithIotas(10+43), user)
		require.NoError(t, err)

		chain.AssertL2AccountIotas(userAgentID, 10+43)
		chain.AssertL2AccountIotas(cAID, 1)

		req = solo.NewCallParams(ScName, sbtestsc.FuncDoNothing.Name)
		_, err = chain.PostRequestOffLedger(req.WithIotas(10+42), user)
		require.NoError(t, err)

		t.Logf("dump accounts:\n%s", chain.DumpAccounts())
		chain.AssertL2AccountIotas(chain.OriginatorAgentID, 0)
		chain.AssertL2AccountIotas(userAgentID, 1)
		chain.AssertL2AccountIotas(cAID, 1+42)
		env.AssertAddressIotas(userAddr, solo.Saldo-10-43)
		env.AssertAddressIotas(chain.OriginatorAddress, solo.Saldo-solo.ChainDustThreshold-5-extraToken)
		chain.AssertCommonAccountIotas(4 + extraToken + 10)
	})
}

func TestOffLedgerTransferEnough(t *testing.T) {
	run2(t, func(t *testing.T, w bool) {
		env, chain := setupChain(t, nil)
		cAID, extraToken := setupTestSandboxSC(t, chain, nil, w)
		user, userAddr, userAgentID := setupDeployer(t, chain)

		chain.AssertL2AccountIotas(userAgentID, 0)
		chain.AssertL2AccountIotas(cAID, 1)

		req := solo.NewCallParams(accounts.Contract.Name, accounts.FuncDeposit.Name)
		_, err := chain.PostRequestSync(req.WithIotas(42), user)
		require.NoError(t, err)

		chain.AssertL2AccountIotas(userAgentID, 42)
		chain.AssertL2AccountIotas(cAID, 1)

		req = solo.NewCallParams(ScName, sbtestsc.FuncDoNothing.Name)
		_, err = chain.PostRequestOffLedger(req.WithIotas(42), user)
		require.NoError(t, err)

		t.Logf("dump accounts:\n%s", chain.DumpAccounts())
		chain.AssertL2AccountIotas(chain.OriginatorAgentID, 0)
		chain.AssertL2AccountIotas(userAgentID, 0)
		chain.AssertL2AccountIotas(cAID, 1+42)
		env.AssertAddressIotas(userAddr, solo.Saldo-42)
		env.AssertAddressIotas(chain.OriginatorAddress, solo.Saldo-solo.ChainDustThreshold-4-extraToken)
		chain.AssertCommonAccountIotas(3 + extraToken)
	})
}

func TestOffLedgerTransferNotEnough(t *testing.T) {
	run2(t, func(t *testing.T, w bool) {
		env, chain := setupChain(t, nil)
		cAID, extraToken := setupTestSandboxSC(t, chain, nil, w)
		user, userAddr, userAgentID := setupDeployer(t, chain)

		chain.AssertL2AccountIotas(userAgentID, 0)
		chain.AssertL2AccountIotas(cAID, 1)

		req := solo.NewCallParams(accounts.Contract.Name, accounts.FuncDeposit.Name)
		_, err := chain.PostRequestSync(req.WithIotas(41), user)
		require.NoError(t, err)

		chain.AssertL2AccountIotas(userAgentID, 41)
		chain.AssertL2AccountIotas(cAID, 1)

		req = solo.NewCallParams(ScName, sbtestsc.FuncDoNothing.Name)
		_, err = chain.PostRequestOffLedger(req.WithIotas(42), user)
		require.NoError(t, err)

		t.Logf("dump accounts:\n%s", chain.DumpAccounts())
		chain.AssertL2AccountIotas(chain.OriginatorAgentID, 0)
		chain.AssertL2AccountIotas(userAgentID, 0)
		chain.AssertL2AccountIotas(cAID, 1+41)
		env.AssertAddressIotas(userAddr, solo.Saldo-41)
		env.AssertAddressIotas(chain.OriginatorAddress, solo.Saldo-solo.ChainDustThreshold-4-extraToken)
		chain.AssertCommonAccountIotas(3 + extraToken)
	})
}

func TestOffLedgerTransferExtra(t *testing.T) {
	run2(t, func(t *testing.T, w bool) {
		env, chain := setupChain(t, nil)
		cAID, extraToken := setupTestSandboxSC(t, chain, nil, w)
		user, userAddr, userAgentID := setupDeployer(t, chain)

		chain.AssertL2AccountIotas(userAgentID, 0)
		chain.AssertL2AccountIotas(cAID, 1)

		req := solo.NewCallParams(accounts.Contract.Name, accounts.FuncDeposit.Name)
		_, err := chain.PostRequestSync(req.WithIotas(43), user)
		require.NoError(t, err)

		chain.AssertL2AccountIotas(userAgentID, 43)
		chain.AssertL2AccountIotas(cAID, 1)

		req = solo.NewCallParams(ScName, sbtestsc.FuncDoNothing.Name)
		_, err = chain.PostRequestOffLedger(req.WithIotas(42), user)
		require.NoError(t, err)

		t.Logf("dump accounts:\n%s", chain.DumpAccounts())
		chain.AssertL2AccountIotas(chain.OriginatorAgentID, 0)
		chain.AssertL2AccountIotas(userAgentID, 1)
		chain.AssertL2AccountIotas(cAID, 1+42)
		env.AssertAddressIotas(userAddr, solo.Saldo-43)
		env.AssertAddressIotas(chain.OriginatorAddress, solo.Saldo-solo.ChainDustThreshold-4-extraToken)
		chain.AssertCommonAccountIotas(3 + extraToken)
	})
}
