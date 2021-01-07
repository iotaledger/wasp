package testcore

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/packages/vm/core/testcore/test_sandbox"
	"github.com/stretchr/testify/require"
	"testing"
)

func setupForTransfer(t *testing.T) (*solo.Solo, *solo.Chain, coretypes.ContractID) {
	glb := solo.New(t, false, false)
	chain := glb.NewChain(nil, "ch1")

	err := chain.DeployContract(nil, test_sandbox.Interface.Name, test_sandbox.Interface.ProgramHash)
	require.NoError(t, err)

	deployed := coretypes.NewContractID(chain.ChainID, coretypes.Hn(test_sandbox.Interface.Name))
	req := solo.NewCall(test_sandbox.Interface.Name, test_sandbox.FuncDoNothing)
	_, err = chain.PostRequest(req, nil)
	require.NoError(t, err)
	return glb, chain, deployed
}

func TestDoNothing(t *testing.T) {
	glb, chain, cID := setupForTransfer(t)
	cAID := coretypes.NewAgentIDFromContractID(cID)
	req := solo.NewCall(test_sandbox.Interface.Name, test_sandbox.FuncDoNothing).
		WithTransfer(balance.ColorIOTA, 42)
	_, err := chain.PostRequest(req, nil)
	require.NoError(t, err)

	t.Logf("dump accounts:\n%s", chain.DumpAccounts())
	chain.AssertAccountBalance(chain.OriginatorAgentID, balance.ColorIOTA, 4)
	chain.AssertAccountBalance(cAID, balance.ColorIOTA, 42)
	glb.AssertAddressBalance(chain.OriginatorAddress, balance.ColorIOTA, testutil.RequestFundsAmount-1-4-42)
}

func TestDoNothingUser(t *testing.T) {
	glb, chain, cID := setupForTransfer(t)
	cAID := coretypes.NewAgentIDFromContractID(cID)
	user := glb.NewSignatureSchemeWithFunds()
	userAddress := user.Address()
	userAgentID := coretypes.NewAgentIDFromAddress(userAddress)

	req := solo.NewCall(test_sandbox.Interface.Name, test_sandbox.FuncDoNothing).
		WithTransfer(balance.ColorIOTA, 42)
	_, err := chain.PostRequest(req, user)
	require.NoError(t, err)

	t.Logf("dump accounts:\n%s", chain.DumpAccounts())
	chain.AssertAccountBalance(chain.OriginatorAgentID, balance.ColorIOTA, 3)
	chain.AssertAccountBalance(userAgentID, balance.ColorIOTA, 1)
	chain.AssertAccountBalance(cAID, balance.ColorIOTA, 42)
	glb.AssertAddressBalance(chain.OriginatorAddress, balance.ColorIOTA, testutil.RequestFundsAmount-1-3)
	glb.AssertAddressBalance(userAddress, balance.ColorIOTA, testutil.RequestFundsAmount-1-42)
}

func TestWithdrawToAddress(t *testing.T) {
	glb, chain, cID := setupForTransfer(t)
	cAID := coretypes.NewAgentIDFromContractID(cID)
	t.Logf("contract agentID: %s", cAID)

	user := glb.NewSignatureSchemeWithFunds()
	userAddress := user.Address()
	userAgentID := coretypes.NewAgentIDFromAddress(userAddress)

	req := solo.NewCall(test_sandbox.Interface.Name, test_sandbox.FuncDoNothing).
		WithTransfer(balance.ColorIOTA, 42)
	_, err := chain.PostRequest(req, user)
	require.NoError(t, err)

	t.Logf("dump accounts 1:\n%s", chain.DumpAccounts())
	chain.AssertAccountBalance(chain.OriginatorAgentID, balance.ColorIOTA, 3)
	chain.AssertAccountBalance(userAgentID, balance.ColorIOTA, 1)
	chain.AssertAccountBalance(cAID, balance.ColorIOTA, 42)
	glb.AssertAddressBalance(chain.OriginatorAddress, balance.ColorIOTA, testutil.RequestFundsAmount-1-3)
	glb.AssertAddressBalance(userAddress, balance.ColorIOTA, testutil.RequestFundsAmount-1-42)

	req = solo.NewCall(test_sandbox.Interface.Name, test_sandbox.FuncSendToAddress,
		test_sandbox.ParamAddress, userAddress,
	)
	_, err = chain.PostRequest(req, nil)
	require.NoError(t, err)

	t.Logf("dump accounts 2:\n%s", chain.DumpAccounts())
	chain.AssertAccountBalance(chain.OriginatorAgentID, balance.ColorIOTA, 4)
	chain.AssertAccountBalance(userAgentID, balance.ColorIOTA, 1)
	chain.AssertAccountBalance(cAID, balance.ColorIOTA, 0)
	glb.AssertAddressBalance(chain.OriginatorAddress, balance.ColorIOTA, testutil.RequestFundsAmount-1-4)
	glb.AssertAddressBalance(userAddress, balance.ColorIOTA, testutil.RequestFundsAmount-1)
}

func TestDoPanicUser(t *testing.T) {
	glb, chain, cID := setupForTransfer(t)
	cAID := coretypes.NewAgentIDFromContractID(cID)
	user := glb.NewSignatureSchemeWithFunds()
	userAddress := user.Address()
	userAgentID := coretypes.NewAgentIDFromAddress(userAddress)

	t.Logf("dump accounts 1:\n%s", chain.DumpAccounts())
	chain.AssertAccountBalance(chain.OriginatorAgentID, balance.ColorIOTA, 3)
	chain.AssertAccountBalance(userAgentID, balance.ColorIOTA, 0)
	chain.AssertAccountBalance(cAID, balance.ColorIOTA, 0)
	glb.AssertAddressBalance(chain.OriginatorAddress, balance.ColorIOTA, testutil.RequestFundsAmount-1-3)
	glb.AssertAddressBalance(userAddress, balance.ColorIOTA, testutil.RequestFundsAmount)

	req := solo.NewCall(test_sandbox.Interface.Name, test_sandbox.FuncPanicFullEP).
		WithTransfer(balance.ColorIOTA, 42)
	_, err := chain.PostRequest(req, user)
	require.Error(t, err)

	t.Logf("dump accounts 2:\n%s", chain.DumpAccounts())
	chain.AssertAccountBalance(chain.OriginatorAgentID, balance.ColorIOTA, 3)
	chain.AssertAccountBalance(userAgentID, balance.ColorIOTA, 1)
	chain.AssertAccountBalance(cAID, balance.ColorIOTA, 0)
	glb.AssertAddressBalance(chain.OriginatorAddress, balance.ColorIOTA, testutil.RequestFundsAmount-1-3)
	glb.AssertAddressBalance(userAddress, balance.ColorIOTA, testutil.RequestFundsAmount-1)
}

func TestDoPanicUserFeeless(t *testing.T) {
	glb, chain, cID := setupForTransfer(t)
	cAID := coretypes.NewAgentIDFromContractID(cID)
	user := glb.NewSignatureSchemeWithFunds()
	userAddress := user.Address()
	userAgentID := coretypes.NewAgentIDFromAddress(userAddress)

	t.Logf("dump accounts 1:\n%s", chain.DumpAccounts())
	chain.AssertAccountBalance(chain.OriginatorAgentID, balance.ColorIOTA, 3)
	chain.AssertAccountBalance(userAgentID, balance.ColorIOTA, 0)
	chain.AssertAccountBalance(cAID, balance.ColorIOTA, 0)
	glb.AssertAddressBalance(chain.OriginatorAddress, balance.ColorIOTA, testutil.RequestFundsAmount-1-3)
	glb.AssertAddressBalance(userAddress, balance.ColorIOTA, testutil.RequestFundsAmount)

	req := solo.NewCall(test_sandbox.Interface.Name, test_sandbox.FuncPanicFullEP).
		WithTransfer(balance.ColorIOTA, 42)
	_, err := chain.PostRequest(req, user)
	require.Error(t, err)

	t.Logf("dump accounts 2:\n%s", chain.DumpAccounts())
	chain.AssertAccountBalance(chain.OriginatorAgentID, balance.ColorIOTA, 3)
	chain.AssertAccountBalance(userAgentID, balance.ColorIOTA, 1)
	chain.AssertAccountBalance(cAID, balance.ColorIOTA, 0)
	glb.AssertAddressBalance(chain.OriginatorAddress, balance.ColorIOTA, testutil.RequestFundsAmount-1-3)
	glb.AssertAddressBalance(userAddress, balance.ColorIOTA, testutil.RequestFundsAmount-1)

	req = solo.NewCall(accounts.Interface.Name, accounts.FuncWithdraw)
	_, err = chain.PostRequest(req, user)
	require.NoError(t, err)

	chain.AssertAccountBalance(chain.OriginatorAgentID, balance.ColorIOTA, 3)
	chain.AssertAccountBalance(userAgentID, balance.ColorIOTA, 0)
	chain.AssertAccountBalance(cAID, balance.ColorIOTA, 0)
	glb.AssertAddressBalance(chain.OriginatorAddress, balance.ColorIOTA, testutil.RequestFundsAmount-1-3)
	glb.AssertAddressBalance(userAddress, balance.ColorIOTA, testutil.RequestFundsAmount)
}

func TestDoPanicUserFee(t *testing.T) {
	glb, chain, cID := setupForTransfer(t)
	cAID := coretypes.NewAgentIDFromContractID(cID)
	user := glb.NewSignatureSchemeWithFunds()
	userAddress := user.Address()
	userAgentID := coretypes.NewAgentIDFromAddress(userAddress)

	t.Logf("dump accounts 1:\n%s", chain.DumpAccounts())
	chain.AssertAccountBalance(chain.OriginatorAgentID, balance.ColorIOTA, 3)
	chain.AssertAccountBalance(userAgentID, balance.ColorIOTA, 0)
	chain.AssertAccountBalance(cAID, balance.ColorIOTA, 0)
	glb.AssertAddressBalance(chain.OriginatorAddress, balance.ColorIOTA, testutil.RequestFundsAmount-1-3)
	glb.AssertAddressBalance(userAddress, balance.ColorIOTA, testutil.RequestFundsAmount)

	req := solo.NewCall(root.Interface.Name, root.FuncSetContractFee,
		root.ParamHname, cID.Hname(),
		root.ParamOwnerFee, 10,
	)
	_, err := chain.PostRequest(req, nil)
	require.NoError(t, err)

	chain.AssertAccountBalance(chain.OriginatorAgentID, balance.ColorIOTA, 4)
	chain.AssertAccountBalance(userAgentID, balance.ColorIOTA, 0)
	chain.AssertAccountBalance(cAID, balance.ColorIOTA, 0)
	glb.AssertAddressBalance(chain.OriginatorAddress, balance.ColorIOTA, testutil.RequestFundsAmount-1-4)
	glb.AssertAddressBalance(userAddress, balance.ColorIOTA, testutil.RequestFundsAmount)

	req = solo.NewCall(test_sandbox.Interface.Name, test_sandbox.FuncPanicFullEP).
		WithTransfer(balance.ColorIOTA, 42)
	_, err = chain.PostRequest(req, user)
	require.Error(t, err)

	t.Logf("dump accounts 2:\n%s", chain.DumpAccounts())
	chain.AssertAccountBalance(chain.OriginatorAgentID, balance.ColorIOTA, 4+10)
	chain.AssertAccountBalance(userAgentID, balance.ColorIOTA, 1)
	chain.AssertAccountBalance(cAID, balance.ColorIOTA, 0)
	glb.AssertAddressBalance(chain.OriginatorAddress, balance.ColorIOTA, testutil.RequestFundsAmount-1-4)
	glb.AssertAddressBalance(userAddress, balance.ColorIOTA, testutil.RequestFundsAmount-1-10)
}

func TestRequestToView(t *testing.T) {
	glb, chain, cID := setupForTransfer(t)
	cAID := coretypes.NewAgentIDFromContractID(cID)
	user := glb.NewSignatureSchemeWithFunds()
	userAddress := user.Address()
	userAgentID := coretypes.NewAgentIDFromAddress(userAddress)

	t.Logf("dump accounts 1:\n%s", chain.DumpAccounts())
	chain.AssertAccountBalance(chain.OriginatorAgentID, balance.ColorIOTA, 3)
	chain.AssertAccountBalance(userAgentID, balance.ColorIOTA, 0)
	chain.AssertAccountBalance(cAID, balance.ColorIOTA, 0)
	glb.AssertAddressBalance(chain.OriginatorAddress, balance.ColorIOTA, testutil.RequestFundsAmount-1-3)
	glb.AssertAddressBalance(userAddress, balance.ColorIOTA, testutil.RequestFundsAmount)

	// sending request to the view entry point should return an error and invoke fallback for tokens
	req := solo.NewCall(test_sandbox.Interface.Name, test_sandbox.FuncJustView).
		WithTransfer(balance.ColorIOTA, 42)
	_, err := chain.PostRequest(req, user)
	require.Error(t, err)

	t.Logf("dump accounts 2:\n%s", chain.DumpAccounts())
	chain.AssertAccountBalance(chain.OriginatorAgentID, balance.ColorIOTA, 3)
	chain.AssertAccountBalance(userAgentID, balance.ColorIOTA, 1)
	chain.AssertAccountBalance(cAID, balance.ColorIOTA, 0)
	glb.AssertAddressBalance(chain.OriginatorAddress, balance.ColorIOTA, testutil.RequestFundsAmount-1-3)
	glb.AssertAddressBalance(userAddress, balance.ColorIOTA, testutil.RequestFundsAmount-1)

}
