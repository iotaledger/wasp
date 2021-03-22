// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package solo

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/vm/core/_default"
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

	_, err := ch.CallView(root.Interface.Name, root.FuncGetChainInfo)
	require.NoError(ch.Env.T, err)

	rootRec, err := ch.FindContract(root.Interface.Name)
	require.NoError(ch.Env.T, err)
	emptyRootRecord := root.NewContractRecord(root.Interface, &coretypes.AgentID{})
	require.EqualValues(ch.Env.T, root.EncodeContractRecord(emptyRootRecord), root.EncodeContractRecord(rootRec))

	defaultRec, err := ch.FindContract(_default.Interface.Name)
	require.NoError(ch.Env.T, err)
	require.EqualValues(ch.Env.T, _default.Interface.Name, defaultRec.Name)
	require.EqualValues(ch.Env.T, _default.Interface.Description, defaultRec.Description)
	require.EqualValues(ch.Env.T, _default.Interface.ProgramHash, defaultRec.ProgramHash)
	require.True(ch.Env.T, defaultRec.Creator.IsNil())

	accountsRec, err := ch.FindContract(accounts.Interface.Name)
	require.NoError(ch.Env.T, err)
	require.EqualValues(ch.Env.T, accounts.Interface.Name, accountsRec.Name)
	require.EqualValues(ch.Env.T, accounts.Interface.Description, accountsRec.Description)
	require.EqualValues(ch.Env.T, accounts.Interface.ProgramHash, accountsRec.ProgramHash)
	require.True(ch.Env.T, accountsRec.Creator.IsNil())

	blobRec, err := ch.FindContract(blob.Interface.Name)
	require.NoError(ch.Env.T, err)
	require.EqualValues(ch.Env.T, blob.Interface.Name, blobRec.Name)
	require.EqualValues(ch.Env.T, blob.Interface.Description, blobRec.Description)
	require.EqualValues(ch.Env.T, blob.Interface.ProgramHash, blobRec.ProgramHash)
	require.True(ch.Env.T, blobRec.Creator.IsNil())

	chainlogRec, err := ch.FindContract(eventlog.Interface.Name)
	require.NoError(ch.Env.T, err)
	require.EqualValues(ch.Env.T, eventlog.Interface.Name, chainlogRec.Name)
	require.EqualValues(ch.Env.T, eventlog.Interface.Description, chainlogRec.Description)
	require.EqualValues(ch.Env.T, eventlog.Interface.ProgramHash, chainlogRec.ProgramHash)
	require.True(ch.Env.T, chainlogRec.Creator.IsNil())

	ch.CheckAccountLedger()
}

// CheckAccountLedger check integrity of the on-chain ledger.
// Sum of all accounts must be equal to total assets
func (ch *Chain) CheckAccountLedger() {
	total := ch.GetTotalAssets()
	accs := ch.GetAccounts()
	sum := make(map[ledgerstate.Color]uint64)
	for _, acc := range accs {
		bals := ch.GetAccountBalance(&acc)
		bals.ForEach(func(col ledgerstate.Color, bal uint64) bool {
			s, _ := sum[col]
			sum[col] = s + bal
			return true
		})
	}
	require.True(ch.Env.T, coretypes.EqualColoredBalances(total, ledgerstate.NewColoredBalances(sum)))
	coreacc := coretypes.NewAgentID(ch.ChainID.AsAddress(), root.Interface.Hname())
	require.Zero(ch.Env.T, ch.GetAccountBalance(coreacc).Size())
	coreacc = coretypes.NewAgentID(ch.ChainID.AsAddress(), blob.Interface.Hname())
	require.Zero(ch.Env.T, ch.GetAccountBalance(coreacc).Size())
	coreacc = coretypes.NewAgentID(ch.ChainID.AsAddress(), accounts.Interface.Hname())
	require.Zero(ch.Env.T, ch.GetAccountBalance(coreacc).Size())
	coreacc = coretypes.NewAgentID(ch.ChainID.AsAddress(), eventlog.Interface.Hname())
	require.Zero(ch.Env.T, ch.GetAccountBalance(coreacc).Size())
}

// AssertAccountBalance asserts the on-chain account balance controlled by agentID for specific color
func (ch *Chain) AssertAccountBalance(agentID *coretypes.AgentID, col ledgerstate.Color, bal uint64) {
	bals := ch.GetAccountBalance(agentID)
	b, _ := bals.Get(col)
	require.EqualValues(ch.Env.T, int(bal), int(b))
}

func (ch *Chain) AssertIotas(agentID *coretypes.AgentID, bal uint64) {
	ch.AssertAccountBalance(agentID, ledgerstate.ColorIOTA, bal)
}

// AssertAccountBalance asserts the on-chain account balance controlled by agentID for specific color
func (ch *Chain) AssertOwnersBalance(col ledgerstate.Color, bal uint64) {
	bals := ch.GetOwnersBalance()
	b, _ := bals.Get(col)
	require.EqualValues(ch.Env.T, int(bal), int(b))
}

func (ch *Chain) AssertOwnersIotas(bal uint64) {
	require.EqualValues(ch.Env.T, int(bal), int(ch.GetOwnersIotas()))
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
