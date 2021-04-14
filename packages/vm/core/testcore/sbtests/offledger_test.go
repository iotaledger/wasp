package sbtests

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/packages/vm/core/testcore/sbtests/sbtestsc"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestOffLedgerNoFeeNoTransfer(t *testing.T) { run2(t, testOffLedgerNoFeeNoTransfer) }
func testOffLedgerNoFeeNoTransfer(t *testing.T, w bool) {
	env, chain := setupChain(t, nil)
	setupTestSandboxSC(t, chain, nil, w)

	owner, ownerAddr := env.NewKeyPairWithFunds()

	req := solo.NewCallParams(SandboxSCName, sbtestsc.FuncSetInt,
		sbtestsc.ParamIntParamName, "ppp",
		sbtestsc.ParamIntParamValue, 314,
	)
	err := chain.PostRequestOffLedger(req, owner)
	require.NoError(t, err)

	ownerAgentID := coretypes.NewAgentID(ownerAddr, 0)
	chain.AssertAccountBalance(ownerAgentID, ledgerstate.ColorIOTA, 0)

	ret, err := chain.CallView(SandboxSCName, sbtestsc.FuncGetInt,
		sbtestsc.ParamIntParamName, "ppp")
	require.NoError(t, err)

	retInt, exists, err := codec.DecodeInt64(ret.MustGet("ppp"))
	require.NoError(t, err)
	require.True(t, exists)
	require.EqualValues(t, 314, retInt)
}

func TestOffLedgerFeesEnough(t *testing.T) { run2(t, testOffLedgerFeesEnough) }
func testOffLedgerFeesEnough(t *testing.T, w bool) {
	env, chain := setupChain(t, nil)
	cAID, extraToken := setupTestSandboxSC(t, chain, nil, w)
	user, userAddr, userAgentID := setupDeployer(t, chain)

	req := solo.NewCallParams(root.Interface.Name, root.FuncSetContractFee,
		root.ParamHname, coretypes.Hn(SandboxSCName),
		root.ParamOwnerFee, 10,
	).WithIotas(1)
	_, err := chain.PostRequestSync(req, nil)
	require.NoError(t, err)

	chain.AssertIotas(userAgentID, 0)
	chain.AssertIotas(cAID, 1)

	req = solo.NewCallParams(accounts.Interface.Name, accounts.FuncDeposit,
	).WithIotas(10)
	_, err = chain.PostRequestSync(req, user)
	require.NoError(t, err)

	chain.AssertIotas(userAgentID, 10)
	chain.AssertIotas(cAID, 1)

	req = solo.NewCallParams(SandboxSCName, sbtestsc.FuncDoNothing,
	).WithIotas(10)
	err = chain.PostRequestOffLedger(req, user)
	require.NoError(t, err)

	t.Logf("dump accounts:\n%s", chain.DumpAccounts())
	chain.AssertIotas(&chain.OriginatorAgentID, 0)
	chain.AssertIotas(userAgentID, 0)
	chain.AssertIotas(cAID, 1)
	env.AssertAddressIotas(userAddr, solo.Saldo-10)
	env.AssertAddressIotas(chain.OriginatorAddress, solo.Saldo-solo.ChainDustThreshold-5-extraToken)
	chain.AssertOwnersIotas(4 + extraToken + 10)
}

func TestOffLedgerFeesNotEnough(t *testing.T) { run2(t, testOffLedgerFeesNotEnough) }
func testOffLedgerFeesNotEnough(t *testing.T, w bool) {
	env, chain := setupChain(t, nil)
	cAID, extraToken := setupTestSandboxSC(t, chain, nil, w)
	user, userAddr, userAgentID := setupDeployer(t, chain)

	req := solo.NewCallParams(root.Interface.Name, root.FuncSetContractFee,
		root.ParamHname, coretypes.Hn(SandboxSCName),
		root.ParamOwnerFee, 10,
	).WithIotas(1)
	_, err := chain.PostRequestSync(req, nil)
	require.NoError(t, err)

	chain.AssertIotas(userAgentID, 0)
	chain.AssertIotas(cAID, 1)

	req = solo.NewCallParams(accounts.Interface.Name, accounts.FuncDeposit,
	).WithIotas(9)
	_, err = chain.PostRequestSync(req, user)
	require.NoError(t, err)

	chain.AssertIotas(userAgentID, 9)
	chain.AssertIotas(cAID, 1)

	req = solo.NewCallParams(SandboxSCName, sbtestsc.FuncDoNothing,
	).WithIotas(10)
	err = chain.PostRequestOffLedger(req, user)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "not enough fees"))

	t.Logf("dump accounts:\n%s", chain.DumpAccounts())
	chain.AssertIotas(&chain.OriginatorAgentID, 0)
	chain.AssertIotas(userAgentID, 0)
	chain.AssertIotas(cAID, 1)
	env.AssertAddressIotas(userAddr, solo.Saldo-9)
	env.AssertAddressIotas(chain.OriginatorAddress, solo.Saldo-solo.ChainDustThreshold-5-extraToken)
	chain.AssertOwnersIotas(4 + extraToken + 9)
}

func TestOffLedgerFeesExtra(t *testing.T) { run2(t, testOffLedgerFeesExtra) }
func testOffLedgerFeesExtra(t *testing.T, w bool) {
	env, chain := setupChain(t, nil)
	cAID, extraToken := setupTestSandboxSC(t, chain, nil, w)
	user, userAddr, userAgentID := setupDeployer(t, chain)

	req := solo.NewCallParams(root.Interface.Name, root.FuncSetContractFee,
		root.ParamHname, coretypes.Hn(SandboxSCName),
		root.ParamOwnerFee, 10,
	).WithIotas(1)
	_, err := chain.PostRequestSync(req, nil)
	require.NoError(t, err)

	chain.AssertIotas(userAgentID, 0)
	chain.AssertIotas(cAID, 1)

	req = solo.NewCallParams(accounts.Interface.Name, accounts.FuncDeposit,
	).WithIotas(11)
	_, err = chain.PostRequestSync(req, user)
	require.NoError(t, err)

	chain.AssertIotas(userAgentID, 11)
	chain.AssertIotas(cAID, 1)

	req = solo.NewCallParams(SandboxSCName, sbtestsc.FuncDoNothing,
	).WithIotas(10)
	err = chain.PostRequestOffLedger(req, user)
	require.NoError(t, err)

	t.Logf("dump accounts:\n%s", chain.DumpAccounts())
	chain.AssertIotas(&chain.OriginatorAgentID, 0)
	chain.AssertIotas(userAgentID, 1)
	chain.AssertIotas(cAID, 1)
	env.AssertAddressIotas(userAddr, solo.Saldo-11)
	env.AssertAddressIotas(chain.OriginatorAddress, solo.Saldo-solo.ChainDustThreshold-5-extraToken)
	chain.AssertOwnersIotas(4 + extraToken + 10)
}

func TestOffLedgerTransferWithFeesEnough(t *testing.T) { run2(t, testOffLedgerTransferWithFeesEnough) }
func testOffLedgerTransferWithFeesEnough(t *testing.T, w bool) {
	env, chain := setupChain(t, nil)
	cAID, extraToken := setupTestSandboxSC(t, chain, nil, w)
	user, userAddr, userAgentID := setupDeployer(t, chain)

	req := solo.NewCallParams(root.Interface.Name, root.FuncSetContractFee,
		root.ParamHname, coretypes.Hn(SandboxSCName),
		root.ParamOwnerFee, 10,
	).WithIotas(1)
	_, err := chain.PostRequestSync(req, nil)
	require.NoError(t, err)

	chain.AssertIotas(userAgentID, 0)
	chain.AssertIotas(cAID, 1)

	req = solo.NewCallParams(accounts.Interface.Name, accounts.FuncDeposit,
	).WithIotas(10+42)
	_, err = chain.PostRequestSync(req, user)
	require.NoError(t, err)

	chain.AssertIotas(userAgentID, 10+42)
	chain.AssertIotas(cAID, 1)

	req = solo.NewCallParams(SandboxSCName, sbtestsc.FuncDoNothing,
	).WithIotas(10+42)
	err = chain.PostRequestOffLedger(req, user)
	require.NoError(t, err)

	t.Logf("dump accounts:\n%s", chain.DumpAccounts())
	chain.AssertIotas(&chain.OriginatorAgentID, 0)
	chain.AssertIotas(userAgentID, 0)
	chain.AssertIotas(cAID, 1+42)
	env.AssertAddressIotas(userAddr, solo.Saldo-10-42)
	env.AssertAddressIotas(chain.OriginatorAddress, solo.Saldo-solo.ChainDustThreshold-5-extraToken)
	chain.AssertOwnersIotas(4 + extraToken + 10)
}

func TestOffLedgerTransferWithFeesNotEnough(t *testing.T) { run2(t, testOffLedgerTransferWithFeesNotEnough) }
func testOffLedgerTransferWithFeesNotEnough(t *testing.T, w bool) {
	env, chain := setupChain(t, nil)
	cAID, extraToken := setupTestSandboxSC(t, chain, nil, w)
	user, userAddr, userAgentID := setupDeployer(t, chain)

	req := solo.NewCallParams(root.Interface.Name, root.FuncSetContractFee,
		root.ParamHname, coretypes.Hn(SandboxSCName),
		root.ParamOwnerFee, 10,
	).WithIotas(1)
	_, err := chain.PostRequestSync(req, nil)
	require.NoError(t, err)

	chain.AssertIotas(userAgentID, 0)
	chain.AssertIotas(cAID, 1)

	req = solo.NewCallParams(accounts.Interface.Name, accounts.FuncDeposit,
	).WithIotas(10+41)
	_, err = chain.PostRequestSync(req, user)
	require.NoError(t, err)

	chain.AssertIotas(userAgentID, 10+41)
	chain.AssertIotas(cAID, 1)

	req = solo.NewCallParams(SandboxSCName, sbtestsc.FuncDoNothing,
	).WithIotas(10+42)
	err = chain.PostRequestOffLedger(req, user)
	require.NoError(t, err)

	t.Logf("dump accounts:\n%s", chain.DumpAccounts())
	chain.AssertIotas(&chain.OriginatorAgentID, 0)
	chain.AssertIotas(userAgentID, 0)
	chain.AssertIotas(cAID, 1+41)
	env.AssertAddressIotas(userAddr, solo.Saldo-10-41)
	env.AssertAddressIotas(chain.OriginatorAddress, solo.Saldo-solo.ChainDustThreshold-5-extraToken)
	chain.AssertOwnersIotas(4 + extraToken + 10)
}

func TestOffLedgerTransferWithFeesExtra(t *testing.T) { run2(t, testOffLedgerTransferWithFeesExtra) }
func testOffLedgerTransferWithFeesExtra(t *testing.T, w bool) {
	env, chain := setupChain(t, nil)
	cAID, extraToken := setupTestSandboxSC(t, chain, nil, w)
	user, userAddr, userAgentID := setupDeployer(t, chain)

	req := solo.NewCallParams(root.Interface.Name, root.FuncSetContractFee,
		root.ParamHname, coretypes.Hn(SandboxSCName),
		root.ParamOwnerFee, 10,
	).WithIotas(1)
	_, err := chain.PostRequestSync(req, nil)
	require.NoError(t, err)

	chain.AssertIotas(userAgentID, 0)
	chain.AssertIotas(cAID, 1)

	req = solo.NewCallParams(accounts.Interface.Name, accounts.FuncDeposit,
	).WithIotas(10+43)
	_, err = chain.PostRequestSync(req, user)
	require.NoError(t, err)

	chain.AssertIotas(userAgentID, 10+43)
	chain.AssertIotas(cAID, 1)

	req = solo.NewCallParams(SandboxSCName, sbtestsc.FuncDoNothing,
	).WithIotas(10+42)
	err = chain.PostRequestOffLedger(req, user)
	require.NoError(t, err)

	t.Logf("dump accounts:\n%s", chain.DumpAccounts())
	chain.AssertIotas(&chain.OriginatorAgentID, 0)
	chain.AssertIotas(userAgentID, 1)
	chain.AssertIotas(cAID, 1+42)
	env.AssertAddressIotas(userAddr, solo.Saldo-10-43)
	env.AssertAddressIotas(chain.OriginatorAddress, solo.Saldo-solo.ChainDustThreshold-5-extraToken)
	chain.AssertOwnersIotas(4 + extraToken + 10)
}

func TestOffLedgerTransferEnough(t *testing.T) { run2(t, testOffLedgerTransferEnough) }
func testOffLedgerTransferEnough(t *testing.T, w bool) {
	env, chain := setupChain(t, nil)
	cAID, extraToken := setupTestSandboxSC(t, chain, nil, w)
	user, userAddr, userAgentID := setupDeployer(t, chain)

	chain.AssertIotas(userAgentID, 0)
	chain.AssertIotas(cAID, 1)

	req := solo.NewCallParams(accounts.Interface.Name, accounts.FuncDeposit,
	).WithIotas(42)
	_, err := chain.PostRequestSync(req, user)
	require.NoError(t, err)

	chain.AssertIotas(userAgentID, 42)
	chain.AssertIotas(cAID, 1)

	req = solo.NewCallParams(SandboxSCName, sbtestsc.FuncDoNothing,
	).WithIotas(42)
	err = chain.PostRequestOffLedger(req, user)
	require.NoError(t, err)

	t.Logf("dump accounts:\n%s", chain.DumpAccounts())
	chain.AssertIotas(&chain.OriginatorAgentID, 0)
	chain.AssertIotas(userAgentID, 0)
	chain.AssertIotas(cAID, 1+42)
	env.AssertAddressIotas(userAddr, solo.Saldo-42)
	env.AssertAddressIotas(chain.OriginatorAddress, solo.Saldo-solo.ChainDustThreshold-4-extraToken)
	chain.AssertOwnersIotas(3 + extraToken)
}

func TestOffLedgerTransferNotEnough(t *testing.T) { run2(t, testOffLedgerTransferNotEnough) }
func testOffLedgerTransferNotEnough(t *testing.T, w bool) {
	env, chain := setupChain(t, nil)
	cAID, extraToken := setupTestSandboxSC(t, chain, nil, w)
	user, userAddr, userAgentID := setupDeployer(t, chain)

	chain.AssertIotas(userAgentID, 0)
	chain.AssertIotas(cAID, 1)

	req := solo.NewCallParams(accounts.Interface.Name, accounts.FuncDeposit,
	).WithIotas(41)
	_, err := chain.PostRequestSync(req, user)
	require.NoError(t, err)

	chain.AssertIotas(userAgentID, 41)
	chain.AssertIotas(cAID, 1)

	req = solo.NewCallParams(SandboxSCName, sbtestsc.FuncDoNothing,
	).WithIotas(42)
	err = chain.PostRequestOffLedger(req, user)
	require.NoError(t, err)

	t.Logf("dump accounts:\n%s", chain.DumpAccounts())
	chain.AssertIotas(&chain.OriginatorAgentID, 0)
	chain.AssertIotas(userAgentID, 0)
	chain.AssertIotas(cAID, 1+41)
	env.AssertAddressIotas(userAddr, solo.Saldo-41)
	env.AssertAddressIotas(chain.OriginatorAddress, solo.Saldo-solo.ChainDustThreshold-4-extraToken)
	chain.AssertOwnersIotas(3 + extraToken)
}

func TestOffLedgerTransferExtra(t *testing.T) { run2(t, testOffLedgerTransferExtra) }
func testOffLedgerTransferExtra(t *testing.T, w bool) {
	env, chain := setupChain(t, nil)
	cAID, extraToken := setupTestSandboxSC(t, chain, nil, w)
	user, userAddr, userAgentID := setupDeployer(t, chain)

	chain.AssertIotas(userAgentID, 0)
	chain.AssertIotas(cAID, 1)

	req := solo.NewCallParams(accounts.Interface.Name, accounts.FuncDeposit,
	).WithIotas(43)
	_, err := chain.PostRequestSync(req, user)
	require.NoError(t, err)

	chain.AssertIotas(userAgentID, 43)
	chain.AssertIotas(cAID, 1)

	req = solo.NewCallParams(SandboxSCName, sbtestsc.FuncDoNothing,
	).WithIotas(42)
	err = chain.PostRequestOffLedger(req, user)
	require.NoError(t, err)

	t.Logf("dump accounts:\n%s", chain.DumpAccounts())
	chain.AssertIotas(&chain.OriginatorAgentID, 0)
	chain.AssertIotas(userAgentID, 1)
	chain.AssertIotas(cAID, 1+42)
	env.AssertAddressIotas(userAddr, solo.Saldo-43)
	env.AssertAddressIotas(chain.OriginatorAddress, solo.Saldo-solo.ChainDustThreshold-4-extraToken)
	chain.AssertOwnersIotas(3 + extraToken)
}
