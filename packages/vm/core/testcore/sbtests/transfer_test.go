package sbtests

import (
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/packages/vm/core/testcore/sbtests/sbtestsc"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestDoNothing(t *testing.T) { run2(t, testDoNothing) }
func testDoNothing(t *testing.T, w bool) {
	env, chain := setupChain(t, nil)
	cAID, extraToken := setupTestSandboxSC(t, chain, nil, w)
	req := solo.NewCallParams(SandboxSCName, sbtestsc.FuncDoNothing).
		WithIotas(42)
	_, err := chain.PostRequestSync(req, nil)
	require.NoError(t, err)

	t.Logf("dump accounts:\n%s", chain.DumpAccounts())
	chain.AssertIotas(&chain.OriginatorAgentID, extraToken)
	chain.AssertIotas(cAID, 43)
	chain.AssertOwnersIotas(2)
	env.AssertAddressIotas(chain.OriginatorAddress, solo.Saldo-100-2-1-42-extraToken)
}

func TestDoNothingUser(t *testing.T) { run2(t, testDoNothingUser) }
func testDoNothingUser(t *testing.T, w bool) {
	env, chain := setupChain(t, nil)
	cAID, extraToken := setupTestSandboxSC(t, chain, nil, w)
	user, userAddr, userAgentID := setupDeployer(t, chain)
	req := solo.NewCallParams(SandboxSCName, sbtestsc.FuncDoNothing).WithIotas(42)
	_, err := chain.PostRequestSync(req, user)
	require.NoError(t, err)

	t.Logf("dump accounts:\n%s", chain.DumpAccounts())
	chain.AssertIotas(&chain.OriginatorAgentID, 4+extraToken)
	chain.AssertIotas(userAgentID, 1)
	chain.AssertIotas(cAID, 42)
	env.AssertAddressIotas(chain.OriginatorAddress, solo.Saldo-1-4-extraToken)
	env.AssertAddressIotas(userAddr, solo.Saldo-1-42)
}

func TestWithdrawToAddress(t *testing.T) { run2(t, testWithdrawToAddress) }
func testWithdrawToAddress(t *testing.T, w bool) {
	env, chain := setupChain(t, nil)
	cAID, extraToken := setupTestSandboxSC(t, chain, nil, w)
	user, userAddress, userAgentID := setupDeployer(t, chain)
	t.Logf("contract agentID: %s", cAID)

	req := solo.NewCallParams(SandboxSCName, sbtestsc.FuncDoNothing).WithIotas(42)
	_, err := chain.PostRequestSync(req, user)
	require.NoError(t, err)

	t.Logf("dump accounts 1:\n%s", chain.DumpAccounts())
	chain.AssertIotas(&chain.OriginatorAgentID, 4+extraToken)
	chain.AssertIotas(userAgentID, 1)
	chain.AssertIotas(cAID, 42)
	env.AssertAddressIotas(chain.OriginatorAddress, solo.Saldo-1-4-extraToken)
	env.AssertAddressIotas(userAddress, solo.Saldo-1-42)

	req = solo.NewCallParams(SandboxSCName, sbtestsc.FuncSendToAddress,
		sbtestsc.ParamAddress, userAddress,
	)
	_, err = chain.PostRequestSync(req, nil)
	require.NoError(t, err)

	t.Logf("dump accounts 2:\n%s", chain.DumpAccounts())
	chain.AssertIotas(&chain.OriginatorAgentID, 5+extraToken)
	chain.AssertIotas(userAgentID, 1)
	chain.AssertIotas(cAID, 0)
	env.AssertAddressIotas(chain.OriginatorAddress, solo.Saldo-1-5-extraToken)
	env.AssertAddressIotas(userAddress, solo.Saldo-1)
}

func TestDoPanicUser(t *testing.T) { run2(t, testDoPanicUser) }
func testDoPanicUser(t *testing.T, w bool) {
	env, chain := setupChain(t, nil)
	cAID, extraToken := setupTestSandboxSC(t, chain, nil, w)
	user, userAddress, userAgentID := setupDeployer(t, chain)

	t.Logf("dump accounts 1:\n%s", chain.DumpAccounts())
	chain.AssertIotas(&chain.OriginatorAgentID, 4+extraToken)
	chain.AssertIotas(userAgentID, 0)
	chain.AssertIotas(cAID, 0)
	env.AssertAddressIotas(chain.OriginatorAddress, solo.Saldo-1-4-extraToken)
	env.AssertAddressIotas(userAddress, solo.Saldo)

	req := solo.NewCallParams(sbtestsc.Interface.Name, sbtestsc.FuncPanicFullEP).
		WithIotas(42)
	_, err := chain.PostRequestSync(req, user)
	require.Error(t, err)

	t.Logf("dump accounts 2:\n%s", chain.DumpAccounts())
	chain.AssertIotas(&chain.OriginatorAgentID, 4+extraToken)
	chain.AssertIotas(userAgentID, 1)
	chain.AssertIotas(cAID, 0)
	env.AssertAddressIotas(chain.OriginatorAddress, solo.Saldo-1-4-extraToken)
	env.AssertAddressIotas(userAddress, solo.Saldo-1)
}

func TestDoPanicUserFeeless(t *testing.T) { run2(t, testDoPanicUserFeeless) }
func testDoPanicUserFeeless(t *testing.T, w bool) {
	env, chain := setupChain(t, nil)
	cAID, extraToken := setupTestSandboxSC(t, chain, nil, w)
	user, userAddress, userAgentID := setupDeployer(t, chain)

	t.Logf("dump accounts 1:\n%s", chain.DumpAccounts())
	chain.AssertIotas(&chain.OriginatorAgentID, 4+extraToken)
	chain.AssertIotas(userAgentID, 0)
	chain.AssertIotas(cAID, 0)
	env.AssertAddressIotas(chain.OriginatorAddress, solo.Saldo-1-4-extraToken)
	env.AssertAddressIotas(userAddress, solo.Saldo)

	req := solo.NewCallParams(sbtestsc.Interface.Name, sbtestsc.FuncPanicFullEP).
		WithIotas(42)
	_, err := chain.PostRequestSync(req, user)
	require.Error(t, err)

	t.Logf("dump accounts 2:\n%s", chain.DumpAccounts())
	chain.AssertIotas(&chain.OriginatorAgentID, 4+extraToken)
	chain.AssertIotas(userAgentID, 1)
	chain.AssertIotas(cAID, 0)
	env.AssertAddressIotas(chain.OriginatorAddress, solo.Saldo-1-4-extraToken)
	env.AssertAddressIotas(userAddress, solo.Saldo-1)

	req = solo.NewCallParams(accounts.Interface.Name, accounts.FuncWithdraw)
	_, err = chain.PostRequestSync(req, user)
	require.NoError(t, err)

	chain.AssertIotas(&chain.OriginatorAgentID, 4+extraToken)
	chain.AssertIotas(userAgentID, 0)
	chain.AssertIotas(cAID, 0)
	env.AssertAddressIotas(chain.OriginatorAddress, solo.Saldo-1-4-extraToken)
	env.AssertAddressIotas(userAddress, solo.Saldo)
}

func TestDoPanicUserFee(t *testing.T) { run2(t, testDoPanicUserFee) }
func testDoPanicUserFee(t *testing.T, w bool) {
	env, chain := setupChain(t, nil)
	cAID, extraToken := setupTestSandboxSC(t, chain, nil, w)
	user, userAddress, userAgentID := setupDeployer(t, chain)

	t.Logf("dump accounts 1:\n%s", chain.DumpAccounts())
	chain.AssertIotas(&chain.OriginatorAgentID, 4+extraToken)
	chain.AssertIotas(userAgentID, 0)
	chain.AssertIotas(cAID, 0)
	env.AssertAddressIotas(chain.OriginatorAddress, solo.Saldo-1-4-extraToken)
	env.AssertAddressIotas(userAddress, solo.Saldo)

	req := solo.NewCallParams(root.Interface.Name, root.FuncSetContractFee,
		root.ParamHname, cAID.Hname(),
		root.ParamOwnerFee, 10,
	)
	_, err := chain.PostRequestSync(req, nil)
	require.NoError(t, err)

	chain.AssertIotas(&chain.OriginatorAgentID, 5+extraToken)
	chain.AssertIotas(userAgentID, 0)
	chain.AssertIotas(cAID, 0)
	env.AssertAddressIotas(chain.OriginatorAddress, solo.Saldo-1-5-extraToken)
	env.AssertAddressIotas(userAddress, solo.Saldo)

	req = solo.NewCallParams(sbtestsc.Interface.Name, sbtestsc.FuncPanicFullEP).WithIotas(42)
	_, err = chain.PostRequestSync(req, user)
	require.Error(t, err)

	t.Logf("dump accounts 2:\n%s", chain.DumpAccounts())
	chain.AssertIotas(&chain.OriginatorAgentID, 5+10+extraToken)
	chain.AssertIotas(userAgentID, 1)
	chain.AssertIotas(cAID, 0)
	env.AssertAddressIotas(chain.OriginatorAddress, solo.Saldo-1-5-extraToken)
	env.AssertAddressIotas(userAddress, solo.Saldo-1-10)
}

func TestRequestToView(t *testing.T) { run2(t, testRequestToView) }
func testRequestToView(t *testing.T, w bool) {
	env, chain := setupChain(t, nil)
	cAID, extraToken := setupTestSandboxSC(t, chain, nil, w)
	user, userAddress, userAgentID := setupDeployer(t, chain)

	t.Logf("dump accounts 1:\n%s", chain.DumpAccounts())
	chain.AssertIotas(&chain.OriginatorAgentID, 4+extraToken)
	chain.AssertIotas(userAgentID, 0)
	chain.AssertIotas(cAID, 0)
	env.AssertAddressIotas(chain.OriginatorAddress, solo.Saldo-1-4-extraToken)
	env.AssertAddressIotas(userAddress, solo.Saldo)

	// sending request to the view entry point should return an error and invoke fallback for tokens
	req := solo.NewCallParams(sbtestsc.Interface.Name, sbtestsc.FuncJustView).WithIotas(42)
	_, err := chain.PostRequestSync(req, user)
	require.Error(t, err)

	t.Logf("dump accounts 2:\n%s", chain.DumpAccounts())
	chain.AssertIotas(&chain.OriginatorAgentID, 4+extraToken)
	chain.AssertIotas(userAgentID, 1)
	chain.AssertIotas(cAID, 0)
	env.AssertAddressIotas(chain.OriginatorAddress, solo.Saldo-1-4-extraToken)
	env.AssertAddressIotas(userAddress, solo.Saldo-1)
}
