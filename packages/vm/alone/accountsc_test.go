// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0
package alone

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/accountsc"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/root"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestAccountsBase(t *testing.T) {
	glb := New(t, true, false)
	chain := glb.NewChain(nil, "chain1")
	chain.CheckAccountLedger()
}

func TestAccountsRepeatInit(t *testing.T) {
	glb := New(t, false, false)
	chain := glb.NewChain(nil, "chain1")
	req := NewCall(accountsc.Interface.Name, "init")
	_, err := chain.PostRequest(req, nil)
	require.Error(t, err)
	chain.CheckAccountLedger()
}

func TestAccountsBase1(t *testing.T) {
	glb := New(t, false, false)
	chain := glb.NewChain(nil, "chain1")
	chain.CheckAccountLedger()

	newOwner := glb.NewSigSchemeWithFunds()
	newOwnerAgentID := coretypes.NewAgentIDFromAddress(newOwner.Address())
	req := NewCall(root.Interface.Name, root.FuncDelegateChainOwnership, root.ParamChainOwner, newOwnerAgentID)
	_, err := chain.PostRequest(req, nil)
	require.NoError(t, err)

	chain.CheckAccountBalance(chain.OriginatorAgentID, balance.ColorIOTA, 2)
	chain.CheckAccountBalance(newOwnerAgentID, balance.ColorIOTA, 0)
	chain.CheckAccountLedger()

	req = NewCall(root.Interface.Name, root.FuncClaimChainOwnership)
	_, err = chain.PostRequest(req, newOwner)
	require.NoError(t, err)

	chain.CheckAccountBalance(chain.OriginatorAgentID, balance.ColorIOTA, 2)
	chain.CheckAccountBalance(newOwnerAgentID, balance.ColorIOTA, 1)
	chain.CheckAccountLedger()
}

func TestAccountsDepositWithdraw(t *testing.T) {
	glb := New(t, false, false)
	chain := glb.NewChain(nil, "chain1")
	chain.CheckAccountLedger()

	newOwner := glb.NewSigSchemeWithFunds()
	newOwnerAgentID := coretypes.NewAgentIDFromAddress(newOwner.Address())
	req := NewCall(accountsc.Interface.Name, accountsc.FuncDeposit).
		WithTransfer(map[balance.Color]int64{
			balance.ColorIOTA: 42,
		})
	_, err := chain.PostRequest(req, newOwner)
	require.NoError(t, err)

	chain.CheckAccountBalance(newOwnerAgentID, balance.ColorIOTA, 42+1)

	req = NewCall(accountsc.Interface.Name, accountsc.FuncWithdraw)
	_, err = chain.PostRequest(req, newOwner)
	require.NoError(t, err)
	chain.CheckAccountBalance(newOwnerAgentID, balance.ColorIOTA, 0)
	glb.CheckUtxodbBalance(newOwner.Address(), balance.ColorIOTA, testutil.RequestFundsAmount)
}
