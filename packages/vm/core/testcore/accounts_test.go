// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package testcore

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"testing"

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
}

func TestAccountsRepeatInit(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")
	req := solo.NewCallParams(accounts.Interface.Name, "init")
	_, err := chain.PostRequestSync(req, nil)
	require.Error(t, err)
	chain.CheckAccountLedger()
}

func TestAccountsBase1(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")
	chain.CheckAccountLedger()

	newOwner, ownerAddr := env.NewKeyPairWithFunds()
	newOwnerAgentID := coretypes.NewAgentIDFromAddress(ownerAddr)
	req := solo.NewCallParams(root.Interface.Name, root.FuncDelegateChainOwnership, root.ParamChainOwner, newOwnerAgentID)
	_, err := chain.PostRequestSync(req, nil)
	require.NoError(t, err)

	chain.AssertAccountBalance(chain.OriginatorAgentID, ledgerstate.ColorIOTA, 2)
	chain.AssertAccountBalance(*newOwnerAgentID, ledgerstate.ColorIOTA, 0)
	chain.CheckAccountLedger()

	req = solo.NewCallParams(root.Interface.Name, root.FuncClaimChainOwnership)
	_, err = chain.PostRequestSync(req, newOwner)
	require.NoError(t, err)

	chain.AssertAccountBalance(chain.OriginatorAgentID, ledgerstate.ColorIOTA, 2)
	chain.AssertAccountBalance(*newOwnerAgentID, ledgerstate.ColorIOTA, 1)
	chain.CheckAccountLedger()
}

func TestAccountsDepositWithdrawToAddress(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")
	chain.CheckAccountLedger()

	newOwner, ownerAddr := env.NewKeyPairWithFunds()
	newOwnerAgentID := coretypes.NewAgentIDFromAddress(ownerAddr)
	req := solo.NewCallParams(accounts.Interface.Name, accounts.FuncDeposit).
		WithTransfer(ledgerstate.ColorIOTA, 42)
	_, err := chain.PostRequestSync(req, newOwner)
	require.NoError(t, err)

	chain.AssertAccountBalance(*newOwnerAgentID, ledgerstate.ColorIOTA, 42+1)

	req = solo.NewCallParams(accounts.Interface.Name, accounts.FuncWithdrawToAddress)
	_, err = chain.PostRequestSync(req, newOwner)
	require.NoError(t, err)
	chain.AssertAccountBalance(*newOwnerAgentID, ledgerstate.ColorIOTA, 0)
	env.AssertAddressBalance(ownerAddr, ledgerstate.ColorIOTA, solo.Saldo)
}

func TestAccountsDepositWithdrawToChainFail(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")
	chain.CheckAccountLedger()

	newOwner, ownerAddr := env.NewKeyPairWithFunds()
	newOwnerAgentID := coretypes.NewAgentIDFromAddress(ownerAddr)
	req := solo.NewCallParams(accounts.Interface.Name, accounts.FuncDeposit).
		WithTransfer(ledgerstate.ColorIOTA, 42)
	_, err := chain.PostRequestSync(req, newOwner)
	require.NoError(t, err)

	chain.AssertAccountBalance(*newOwnerAgentID, ledgerstate.ColorIOTA, 42+1)

	req = solo.NewCallParams(accounts.Interface.Name, accounts.FuncWithdrawToChain)
	_, err = chain.PostRequestSync(req, newOwner)
	require.Error(t, err)
	chain.AssertAccountBalance(*newOwnerAgentID, ledgerstate.ColorIOTA, 42+2)
	env.AssertAddressBalance(ownerAddr, ledgerstate.ColorIOTA, solo.Saldo-42-2)
}
