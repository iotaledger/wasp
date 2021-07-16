// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package solo

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/vm/core"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
	"github.com/iotaledger/wasp/packages/vm/core/eventlog"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/stretchr/testify/require"
)

// AssertAddressBalance asserts the UTXODB address balance of specific color in the address
func (env *Solo) AssertAddressBalance(addr ledgerstate.Address, col ledgerstate.Color, expected uint64) {
	require.EqualValues(env.T, int(expected), int(env.GetAddressBalance(addr, col)))
}

func (env *Solo) AssertAddressIotas(addr ledgerstate.Address, expected uint64) {
	env.AssertAddressBalance(addr, ledgerstate.ColorIOTA, expected)
}

// CheckChain checks fundamental integrity of the chain
func (ch *Chain) CheckChain() {
	_, err := ch.CallView(root.Interface.Name, root.FuncGetChainInfo.Name)
	require.NoError(ch.Env.T, err)

	for _, rec := range core.AllCoreContractsByHash {
		recFromState, err := ch.FindContract(rec.Interface.Name)
		require.NoError(ch.Env.T, err)
		require.EqualValues(ch.Env.T, rec.Interface.Name, recFromState.Name)
		require.EqualValues(ch.Env.T, rec.Interface.Description, recFromState.Description)
		require.EqualValues(ch.Env.T, rec.Interface.ProgramHash, recFromState.ProgramHash)
		require.True(ch.Env.T, recFromState.Creator.IsNil())
	}
	ch.CheckAccountLedger()
}

// CheckAccountLedger check integrity of the on-chain ledger.
// Sum of all accounts must be equal to total assets
func (ch *Chain) CheckAccountLedger() {
	total := ch.GetTotalAssets()
	accs := ch.GetAccounts()
	sum := make(map[ledgerstate.Color]uint64)
	for i := range accs {
		acc := accs[i]
		bals := ch.GetAccountBalance(&acc)
		bals.ForEach(func(col ledgerstate.Color, bal uint64) bool {
			s := sum[col]
			sum[col] = s + bal
			return true
		})
	}
	require.True(ch.Env.T, iscp.EqualColoredBalances(total, ledgerstate.NewColoredBalances(sum)))
	coreacc := iscp.NewAgentID(ch.ChainID.AsAddress(), root.Interface.Hname())
	require.Zero(ch.Env.T, ch.GetAccountBalance(coreacc).Size())
	coreacc = iscp.NewAgentID(ch.ChainID.AsAddress(), blob.Interface.Hname())
	require.Zero(ch.Env.T, ch.GetAccountBalance(coreacc).Size())
	coreacc = iscp.NewAgentID(ch.ChainID.AsAddress(), accounts.Interface.Hname())
	require.Zero(ch.Env.T, ch.GetAccountBalance(coreacc).Size())
	coreacc = iscp.NewAgentID(ch.ChainID.AsAddress(), eventlog.Interface.Hname())
	require.Zero(ch.Env.T, ch.GetAccountBalance(coreacc).Size())
}

// AssertAccountBalance asserts the on-chain account balance controlled by agentID for specific color
func (ch *Chain) AssertAccountBalance(agentID *iscp.AgentID, col ledgerstate.Color, bal uint64) {
	bals := ch.GetAccountBalance(agentID)
	b, _ := bals.Get(col)
	require.EqualValues(ch.Env.T, int(bal), int(b))
}

func (ch *Chain) AssertIotas(agentID *iscp.AgentID, bal uint64) {
	ch.AssertAccountBalance(agentID, ledgerstate.ColorIOTA, bal)
}

// AssertAccountBalance asserts the on-chain account balance controlled by agentID for specific color
func (ch *Chain) AssertOwnersBalance(col ledgerstate.Color, bal uint64) {
	bals := ch.GetCommonAccountBalance()
	b, _ := bals.Get(col)
	require.EqualValues(ch.Env.T, int(bal), int(b))
}

func (ch *Chain) AssertCommonAccountIotas(bal uint64) {
	require.EqualValues(ch.Env.T, int(bal), int(ch.GetCommonAccountIotas()))
}

// AssertAccountBalance asserts the on-chain account balance controlled by agentID for specific color
func (ch *Chain) AssertTotalAssets(col ledgerstate.Color, bal uint64) {
	bals := ch.GetTotalAssets()
	b, _ := bals.Get(col)
	require.EqualValues(ch.Env.T, int(bal), int(b))
}

func (ch *Chain) AssertTotalIotas(bal uint64) {
	iotas := ch.GetTotalIotas()
	require.EqualValues(ch.Env.T, int(bal), int(iotas))
}

func (ch *Chain) CheckControlAddresses() {
	rec := ch.GetControlAddresses()
	require.True(ch.Env.T, rec.StateAddress.Equals(ch.StateControllerAddress))
	require.True(ch.Env.T, rec.GoverningAddress.Equals(ch.StateControllerAddress))
	require.EqualValues(ch.Env.T, 0, rec.SinceBlockIndex)
}
