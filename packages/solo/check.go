// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package solo

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
	"github.com/iotaledger/wasp/packages/vm/core/eventlog"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/stretchr/testify/require"
)

// AssertAddressBalance asserts the UTXODB address balance of specific color in the address
func (env *Solo) AssertAddressBalance(addr ledgerstate.Address, col ledgerstate.Color, expected uint64) {
	require.EqualValues(env.T, expected, env.GetAddressBalance(addr, col))
}

// CheckChain checks fundamental integrity of the chain
func (ch *Chain) CheckChain() {

	_, err := ch.CallView(root.Interface.Name, root.FuncGetChainInfo)
	require.NoError(ch.Env.T, err)

	rootRec, err := ch.FindContract(root.Interface.Name)
	require.NoError(ch.Env.T, err)
	emptyRootRecord := root.NewContractRecord(root.Interface, &coretypes.AgentID{})
	require.EqualValues(ch.Env.T, root.EncodeContractRecord(&emptyRootRecord), root.EncodeContractRecord(rootRec))

	accountsRec, err := ch.FindContract(accounts.Interface.Name)
	require.NoError(ch.Env.T, err)
	require.EqualValues(ch.Env.T, accounts.Interface.Name, accountsRec.Name)
	require.EqualValues(ch.Env.T, accounts.Interface.Description, accountsRec.Description)
	require.EqualValues(ch.Env.T, accounts.Interface.ProgramHash, accountsRec.ProgramHash)
	require.EqualValues(ch.Env.T, ch.OriginatorAgentID, accountsRec.Creator)

	blobRec, err := ch.FindContract(blob.Interface.Name)
	require.NoError(ch.Env.T, err)
	require.EqualValues(ch.Env.T, blob.Interface.Name, blobRec.Name)
	require.EqualValues(ch.Env.T, blob.Interface.Description, blobRec.Description)
	require.EqualValues(ch.Env.T, blob.Interface.ProgramHash, blobRec.ProgramHash)
	require.EqualValues(ch.Env.T, ch.OriginatorAgentID, blobRec.Creator)

	chainlogRec, err := ch.FindContract(eventlog.Interface.Name)
	require.NoError(ch.Env.T, err)
	require.EqualValues(ch.Env.T, eventlog.Interface.Name, chainlogRec.Name)
	require.EqualValues(ch.Env.T, eventlog.Interface.Description, chainlogRec.Description)
	require.EqualValues(ch.Env.T, eventlog.Interface.ProgramHash, chainlogRec.ProgramHash)
	require.EqualValues(ch.Env.T, ch.OriginatorAgentID, chainlogRec.Creator)

	ch.CheckAccountLedger()
}

// CheckAccountLedger check integrity of the on-chain ledger.
// Sum of all accounts must be equal to total assets
func (ch *Chain) CheckAccountLedger() {
	total := ch.GetTotalAssets()
	accs := ch.GetAccounts()
	sum := make(map[ledgerstate.Color]uint64)
	for _, acc := range accs {
		bals := ch.GetAccountBalance(acc)
		bals.ForEach(func(col ledgerstate.Color, bal uint64) bool {
			s, _ := sum[col]
			sum[col] = s + bal
			return true
		})
	}
	require.True(ch.Env.T, coretypes.EqualColoredBalances(total, ledgerstate.NewColoredBalances(sum)))
}

// AssertAccountBalance asserts the on-chain account balance controlled by agentID for specific color
func (ch *Chain) AssertAccountBalance(agentID coretypes.AgentID, col ledgerstate.Color, bal int64) {
	bals := ch.GetAccountBalance(agentID)
	b, _ := bals.Get(col)
	require.EqualValues(ch.Env.T, bal, b)
}
