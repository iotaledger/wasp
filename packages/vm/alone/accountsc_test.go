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
	e := New(t, false, false)
	e.CheckAccountLedger()
}

func TestAccountsRepeatInit(t *testing.T) {
	e := New(t, false, false)
	req := NewCall(accountsc.Interface.Name, "init")
	_, err := e.PostRequest(req, nil)
	require.Error(t, err)
	e.CheckAccountLedger()
}

func TestAccountsBase1(t *testing.T) {
	e := New(t, false, false)
	e.CheckAccountLedger()

	newOwner := e.NewSigScheme()
	newOwnerAgentID := coretypes.NewAgentIDFromAddress(newOwner.Address())
	req := NewCall(root.Interface.Name, root.FuncAllowChangeChainOwner, root.ParamChainOwner, newOwnerAgentID)
	_, err := e.PostRequest(req, nil)
	require.NoError(t, err)

	e.CheckAccountBalance(e.OriginatorAgentID, balance.ColorIOTA, 2)
	e.CheckAccountBalance(newOwnerAgentID, balance.ColorIOTA, 0)
	e.CheckAccountLedger()

	req = NewCall(root.Interface.Name, root.FuncChangeChainOwner)
	_, err = e.PostRequest(req, newOwner)
	require.NoError(t, err)

	e.CheckAccountBalance(e.OriginatorAgentID, balance.ColorIOTA, 2)
	e.CheckAccountBalance(newOwnerAgentID, balance.ColorIOTA, 1)
	e.CheckAccountLedger()
}

func TestAccountsDepositWithdraw(t *testing.T) {
	e := New(t, false, false)
	e.CheckAccountLedger()

	newOwner := e.NewSigScheme()
	newOwnerAgentID := coretypes.NewAgentIDFromAddress(newOwner.Address())
	req := NewCall(accountsc.Interface.Name, accountsc.FuncDeposit).
		WithTransfer(map[balance.Color]int64{
			balance.ColorIOTA: 42,
		})
	_, err := e.PostRequest(req, newOwner)
	require.NoError(t, err)

	e.CheckAccountBalance(newOwnerAgentID, balance.ColorIOTA, 42+1)

	req = NewCall(accountsc.Interface.Name, accountsc.FuncWithdraw)
	_, err = e.PostRequest(req, newOwner)
	require.NoError(t, err)
	e.CheckAccountBalance(newOwnerAgentID, balance.ColorIOTA, 0)
	e.CheckUtxodbBalance(newOwner.Address(), balance.ColorIOTA, testutil.RequestFundsAmount)
}
