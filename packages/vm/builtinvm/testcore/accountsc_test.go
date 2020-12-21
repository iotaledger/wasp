// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package testcore

import (
	"testing"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/accounts"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/root"
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

func TestAccountsDepositWithdraw(t *testing.T) {
	glb := solo.New(t, false, false)
	chain := glb.NewChain(nil, "chain1")
	chain.CheckAccountLedger()

	newOwner := glb.NewSignatureSchemeWithFunds()
	newOwnerAgentID := coretypes.NewAgentIDFromAddress(newOwner.Address())
	req := solo.NewCall(accounts.Interface.Name, accounts.FuncDeposit).
		WithTransfer(map[balance.Color]int64{
			balance.ColorIOTA: 42,
		})
	_, err := chain.PostRequest(req, newOwner)
	require.NoError(t, err)

	chain.AssertAccountBalance(newOwnerAgentID, balance.ColorIOTA, 42+1)

	req = solo.NewCall(accounts.Interface.Name, accounts.FuncWithdraw)
	_, err = chain.PostRequest(req, newOwner)
	require.NoError(t, err)
	chain.AssertAccountBalance(newOwnerAgentID, balance.ColorIOTA, 0)
	glb.AssertUtxodbBalance(newOwner.Address(), balance.ColorIOTA, testutil.RequestFundsAmount)
}
