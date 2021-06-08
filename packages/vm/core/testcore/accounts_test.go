// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package testcore

import (
	"testing"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"

	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/stretchr/testify/require"
)

func TestAccountsBase(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")
	chain.CheckAccountLedger()
	chain.AssertTotalIotas(1)
	chain.AssertOwnersIotas(1)
}

func TestAccountsRepeatInit(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")
	req := solo.NewCallParams(accounts.Interface.Name, "init").WithIotas(1)
	_, err := chain.PostRequestSync(req, nil)
	require.Error(t, err)
	chain.CheckAccountLedger()
	chain.AssertTotalIotas(1)
	chain.AssertOwnersIotas(1)
}

func TestAccountsBase1(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")
	chain.CheckAccountLedger()

	newOwner, ownerAddr := env.NewKeyPairWithFunds()
	newOwnerAgentID := coretypes.NewAgentID(ownerAddr, 0)
	req := solo.NewCallParams(root.Interface.Name, root.FuncDelegateChainOwnership, root.ParamChainOwner, newOwnerAgentID)
	req.WithIotas(1)
	_, err := chain.PostRequestSync(req, nil)
	require.NoError(t, err)

	chain.AssertIotas(&chain.OriginatorAgentID, 0)
	chain.AssertIotas(newOwnerAgentID, 0)
	chain.AssertIotas(chain.ContractAgentID(root.Interface.Name), 0)
	chain.CheckAccountLedger()

	req = solo.NewCallParams(root.Interface.Name, root.FuncClaimChainOwnership).WithIotas(1)
	_, err = chain.PostRequestSync(req, newOwner)
	require.NoError(t, err)

	chain.AssertIotas(&chain.OriginatorAgentID, 0)
	chain.AssertIotas(newOwnerAgentID, 0)
	chain.AssertIotas(chain.ContractAgentID(root.Interface.Name), 0)
	chain.CheckAccountLedger()
	chain.AssertTotalIotas(3)
	chain.AssertOwnersIotas(3)
}

func TestAccountsDepositWithdrawToAddress(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")
	chain.CheckAccountLedger()

	newOwner, newOwnerAddr := env.NewKeyPairWithFunds()
	newOwnerAgentID := coretypes.NewAgentID(newOwnerAddr, 0)
	req := solo.NewCallParams(accounts.Interface.Name, accounts.FuncDeposit).
		WithIotas(42)
	_, err := chain.PostRequestSync(req, newOwner)
	require.NoError(t, err)

	chain.AssertAccountBalance(newOwnerAgentID, ledgerstate.ColorIOTA, 42)

	req = solo.NewCallParams(accounts.Interface.Name, accounts.FuncWithdraw).WithIotas(1)
	_, err = chain.PostRequestSync(req, newOwner)
	require.NoError(t, err)
	chain.CheckAccountLedger()
	chain.AssertAccountBalance(newOwnerAgentID, ledgerstate.ColorIOTA, 0)
	env.AssertAddressBalance(newOwnerAddr, ledgerstate.ColorIOTA, solo.Saldo-1)
	chain.AssertTotalIotas(2)
	chain.AssertOwnersIotas(2)

	// withdraw owner's iotas
	_, ownerFromChain, _ := chain.GetInfo()
	require.True(t, chain.OriginatorAgentID.Equals(&ownerFromChain))
	t.Logf("origintor/owner: %s", chain.OriginatorAgentID.String())

	req = solo.NewCallParams(accounts.Interface.Name, accounts.FuncWithdraw).WithIotas(1)
	_, err = chain.PostRequestSync(req, chain.OriginatorKeyPair)
	require.NoError(t, err)
	chain.AssertTotalIotas(0)
	chain.AssertOwnersIotas(0)

}

func TestAccountsDepositWithdrawToChainFail(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")
	chain.CheckAccountLedger()

	newOwner, ownerAddr := env.NewKeyPairWithFunds()
	newOwnerAgentID := coretypes.NewAgentID(ownerAddr, 0)
	req := solo.NewCallParams(accounts.Interface.Name, accounts.FuncDeposit).
		WithIotas(42)
	_, err := chain.PostRequestSync(req, newOwner)
	require.NoError(t, err)

	chain.AssertAccountBalance(newOwnerAgentID, ledgerstate.ColorIOTA, 42)

	req = solo.NewCallParams(accounts.Interface.Name, accounts.FuncWithdraw)
	_, err = chain.PostRequestSync(req, newOwner)
	require.Error(t, err)
	chain.AssertAccountBalance(newOwnerAgentID, ledgerstate.ColorIOTA, 42)
	env.AssertAddressBalance(ownerAddr, ledgerstate.ColorIOTA, solo.Saldo-42)
	chain.AssertTotalIotas(43)
	chain.AssertOwnersIotas(1)
}
