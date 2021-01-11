// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package testcore

import (
	"github.com/iotaledger/wasp/packages/vm/core/testcore/test_sandbox"
	"testing"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/stretchr/testify/require"
)

func TestAccountsBase(t *testing.T) {
	glb := solo.New(t, false, false)
	chain := glb.NewChain(nil, "chain1")
	chain.CheckAccountLedger()
}

func TestAccountsRepeatInit(t *testing.T) {
	glb := solo.New(t, false, false)
	chain := glb.NewChain(nil, "chain1")
	req := solo.NewCall(accounts.Interface.Name, "init")
	_, err := chain.PostRequest(req, nil)
	require.Error(t, err)
	chain.CheckAccountLedger()
}

func TestAccountsBase1(t *testing.T) {
	glb := solo.New(t, false, false)
	chain := glb.NewChain(nil, "chain1")
	chain.CheckAccountLedger()

	newOwner := glb.NewSignatureSchemeWithFunds()
	newOwnerAgentID := coretypes.NewAgentIDFromAddress(newOwner.Address())
	req := solo.NewCall(root.Interface.Name, root.FuncDelegateChainOwnership, root.ParamChainOwner, newOwnerAgentID)
	_, err := chain.PostRequest(req, nil)
	require.NoError(t, err)

	chain.AssertAccountBalance(chain.OriginatorAgentID, balance.ColorIOTA, 2)
	chain.AssertAccountBalance(newOwnerAgentID, balance.ColorIOTA, 0)
	chain.CheckAccountLedger()

	req = solo.NewCall(root.Interface.Name, root.FuncClaimChainOwnership)
	_, err = chain.PostRequest(req, newOwner)
	require.NoError(t, err)

	chain.AssertAccountBalance(chain.OriginatorAgentID, balance.ColorIOTA, 2)
	chain.AssertAccountBalance(newOwnerAgentID, balance.ColorIOTA, 1)
	chain.CheckAccountLedger()
}

func TestAccountsDepositWithdrawToAddress(t *testing.T) {
	glb := solo.New(t, false, false)
	chain := glb.NewChain(nil, "chain1")
	chain.CheckAccountLedger()

	newOwner := glb.NewSignatureSchemeWithFunds()
	newOwnerAgentID := coretypes.NewAgentIDFromAddress(newOwner.Address())
	req := solo.NewCall(accounts.Interface.Name, accounts.FuncDeposit).
		WithTransfer(balance.ColorIOTA, 42)
	_, err := chain.PostRequest(req, newOwner)
	require.NoError(t, err)

	chain.AssertAccountBalance(newOwnerAgentID, balance.ColorIOTA, 42+1)

	req = solo.NewCall(accounts.Interface.Name, accounts.FuncWithdrawToAddress)
	_, err = chain.PostRequest(req, newOwner)
	require.NoError(t, err)
	chain.AssertAccountBalance(newOwnerAgentID, balance.ColorIOTA, 0)
	glb.AssertAddressBalance(newOwner.Address(), balance.ColorIOTA, testutil.RequestFundsAmount)
}

func TestAccountsDepositWithdrawToChainFail(t *testing.T) {
	glb := solo.New(t, false, false)
	chain := glb.NewChain(nil, "chain1")
	chain.CheckAccountLedger()

	newOwner := glb.NewSignatureSchemeWithFunds()
	newOwnerAgentID := coretypes.NewAgentIDFromAddress(newOwner.Address())
	req := solo.NewCall(accounts.Interface.Name, accounts.FuncDeposit).
		WithTransfer(balance.ColorIOTA, 42)
	_, err := chain.PostRequest(req, newOwner)
	require.NoError(t, err)

	chain.AssertAccountBalance(newOwnerAgentID, balance.ColorIOTA, 42+1)

	req = solo.NewCall(accounts.Interface.Name, accounts.FuncWithdrawToChain)
	_, err = chain.PostRequest(req, newOwner)
	require.Error(t, err)
	chain.AssertAccountBalance(newOwnerAgentID, balance.ColorIOTA, 42+2)
	glb.AssertAddressBalance(newOwner.Address(), balance.ColorIOTA, testutil.RequestFundsAmount-42-2)
}

func Test2Chains(t *testing.T) {
	env := solo.New(t, false, false)
	chain1 := env.NewChain(nil, "ch1")
	chain2 := env.NewChain(nil, "ch2")
	chain1.CheckAccountLedger()
	chain2.CheckAccountLedger()

	err := chain1.DeployContract(nil, test_sandbox.Interface.Name, test_sandbox.Interface.ProgramHash)
	require.NoError(t, err)
	chain1.CheckChain()
	contractID1 := coretypes.NewContractID(chain1.ChainID, test_sandbox.Interface.Hname())
	contractAgentID1 := coretypes.NewAgentIDFromContractID(contractID1)

	err = chain2.DeployContract(nil, test_sandbox.Interface.Name, test_sandbox.Interface.ProgramHash)
	require.NoError(t, err)
	chain2.CheckChain()
	contractID2 := coretypes.NewContractID(chain2.ChainID, test_sandbox.Interface.Hname())
	contractAgentID2 := coretypes.NewAgentIDFromContractID(contractID2)

	userWallet := env.NewSignatureSchemeWithFunds()
	userAddress := userWallet.Address()
	userAgentID := coretypes.NewAgentIDFromAddress(userAddress)
	env.AssertAddressBalance(userAddress, balance.ColorIOTA, 1337)

	chain1.AssertAccountBalance(contractAgentID1, balance.ColorIOTA, 0)
	chain1.AssertAccountBalance(contractAgentID2, balance.ColorIOTA, 0)
	chain2.AssertAccountBalance(contractAgentID1, balance.ColorIOTA, 0)
	chain2.AssertAccountBalance(contractAgentID2, balance.ColorIOTA, 0)

	req := solo.NewCall(accounts.Interface.Name, accounts.FuncDeposit,
		accounts.ParamAgentID, contractAgentID2,
	).WithTransfer(
		balance.ColorIOTA, 42,
	)
	_, err = chain1.PostRequest(req, userWallet)
	require.NoError(t, err)

	accountsAgentID1 := coretypes.NewAgentIDFromContractID(accounts.Interface.ContractID(chain1.ChainID))
	accountsAgentID2 := coretypes.NewAgentIDFromContractID(accounts.Interface.ContractID(chain2.ChainID))

	env.AssertAddressBalance(userAddress, balance.ColorIOTA, 1337-43)
	chain1.AssertAccountBalance(userAgentID, balance.ColorIOTA, 1)
	chain2.AssertAccountBalance(userAgentID, balance.ColorIOTA, 0)
	chain1.AssertAccountBalance(contractAgentID1, balance.ColorIOTA, 0)
	chain1.AssertAccountBalance(contractAgentID2, balance.ColorIOTA, 42)
	chain2.AssertAccountBalance(contractAgentID1, balance.ColorIOTA, 0)
	chain2.AssertAccountBalance(contractAgentID2, balance.ColorIOTA, 0)

	chain1.AssertAccountBalance(accountsAgentID1, balance.ColorIOTA, 0)
	chain1.AssertAccountBalance(accountsAgentID2, balance.ColorIOTA, 0)
	chain2.AssertAccountBalance(accountsAgentID1, balance.ColorIOTA, 0)
	chain2.AssertAccountBalance(accountsAgentID2, balance.ColorIOTA, 0)

	req = solo.NewCall(test_sandbox.Name, test_sandbox.FuncWithdrawToChain,
		test_sandbox.ParamChainID, chain1.ChainID,
	).WithTransfer(
		balance.ColorIOTA, 3,
	)
	_, err = chain2.PostRequest(req, userWallet)
	require.NoError(t, err)

	chain1.WaitForEmptyBacklog()
	chain2.WaitForEmptyBacklog()

	env.AssertAddressBalance(userAddress, balance.ColorIOTA, 1337-47)
	chain1.AssertAccountBalance(userAgentID, balance.ColorIOTA, 1)
	chain2.AssertAccountBalance(userAgentID, balance.ColorIOTA, 1)
	chain1.AssertAccountBalance(contractAgentID1, balance.ColorIOTA, 0)
	chain1.AssertAccountBalance(contractAgentID2, balance.ColorIOTA, 0)
	chain2.AssertAccountBalance(contractAgentID1, balance.ColorIOTA, 0)
	chain2.AssertAccountBalance(contractAgentID2, balance.ColorIOTA, 43)

	chain1.AssertAccountBalance(accountsAgentID1, balance.ColorIOTA, 1) // !!!! TODO
	chain1.AssertAccountBalance(accountsAgentID2, balance.ColorIOTA, 0)
	chain2.AssertAccountBalance(accountsAgentID1, balance.ColorIOTA, 1) // !!!! TODO
	chain2.AssertAccountBalance(accountsAgentID2, balance.ColorIOTA, 0)
}
