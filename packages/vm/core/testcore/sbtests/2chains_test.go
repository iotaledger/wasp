package sbtests

import (
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/testcore/sbtests/sbtestsc"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test2Chains(t *testing.T) { run2(t, test2Chains) }
func test2Chains(t *testing.T, w bool) {
	env := solo.New(t, false, false)
	chain1 := env.NewChain(nil, "ch1")
	chain2 := env.NewChain(nil, "ch2")
	chain1.CheckAccountLedger()
	chain2.CheckAccountLedger()

	contractAgentID1, _ := setupTestSandboxSC(t, chain1, nil, w)
	contractAgentID2, _ := setupTestSandboxSC(t, chain2, nil, w)

	userWallet, userAddress := env.NewKeyPairWithFunds()
	userAgentID := coretypes.NewAgentID(userAddress, 0)
	env.AssertAddressIotas(userAddress, solo.Saldo)

	chain1.AssertIotas(contractAgentID1, 0)
	chain1.AssertIotas(contractAgentID2, 0)
	chain2.AssertIotas(contractAgentID1, 0)
	chain2.AssertIotas(contractAgentID2, 0)

	req := solo.NewCallParams(accounts.Interface.Name, accounts.FuncDeposit,
		accounts.ParamAgentID, contractAgentID2,
	).WithIotas(42)
	_, err := chain1.PostRequestSync(req, userWallet)
	require.NoError(t, err)

	accountsAgentID1 := coretypes.NewAgentID(chain1.ChainID.AsAddress(), coretypes.HnameAccounts)
	accountsAgentID2 := coretypes.NewAgentID(chain2.ChainID.AsAddress(), coretypes.HnameAccounts)

	env.AssertAddressIotas(userAddress, solo.Saldo-43)
	chain1.AssertIotas(userAgentID, 1)
	chain2.AssertIotas(userAgentID, 0)
	chain1.AssertIotas(contractAgentID1, 0)
	chain1.AssertIotas(contractAgentID2, 42)
	chain2.AssertIotas(contractAgentID1, 0)
	chain2.AssertIotas(contractAgentID2, 0)

	chain1.AssertIotas(accountsAgentID1, 0)
	chain1.AssertIotas(accountsAgentID2, 0)
	chain2.AssertIotas(accountsAgentID1, 0)
	chain2.AssertIotas(accountsAgentID2, 0)

	req = solo.NewCallParams(sbtestsc.Name, sbtestsc.FuncWithdrawToChain,
		sbtestsc.ParamChainID, chain1.ChainID,
	).WithIotas(3)

	_, err = chain2.PostRequestSync(req, userWallet)
	require.NoError(t, err)

	chain1.WaitForEmptyBacklog()
	chain2.WaitForEmptyBacklog()

	env.AssertAddressIotas(userAddress, solo.Saldo-47)
	chain1.AssertIotas(userAgentID, 1)
	chain2.AssertIotas(userAgentID, 0)
	chain1.AssertIotas(contractAgentID1, 0)
	chain1.AssertIotas(contractAgentID2, 42)
	chain2.AssertIotas(contractAgentID1, 0)
	chain2.AssertIotas(contractAgentID2, 43)

	chain1.AssertIotas(accountsAgentID1, 0)
	chain1.AssertIotas(accountsAgentID2, 0)
	chain2.AssertIotas(accountsAgentID1, 0)
	chain2.AssertIotas(accountsAgentID2, 0)
}
