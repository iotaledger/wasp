// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package solo

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/cbalances"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
	"github.com/iotaledger/wasp/packages/vm/core/eventlog"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/stretchr/testify/require"
)

// AssertUtxodbBalance asserts the UTXODB balance of specific color in the address
func (glb *Solo) AssertUtxodbBalance(addr address.Address, col balance.Color, expected int64) {
	require.EqualValues(glb.T, expected, glb.GetUtxodbBalance(addr, col))
}

// CheckChain checks fundamental integrity of the chain
func (ch *Chain) CheckChain() {
	// post and call same request, must be equal results
	req := NewCall(root.Interface.Name, root.FuncGetChainInfo)
	res1, err := ch.PostRequest(req, nil)
	require.NoError(ch.Glb.T, err)

	res2, err := ch.CallView(root.Interface.Name, root.FuncGetChainInfo)
	require.NoError(ch.Glb.T, err)

	require.EqualValues(ch.Glb.T, res1.Hash(), res2.Hash())

	rootRec, err := ch.FindContract(root.Interface.Name)
	require.NoError(ch.Glb.T, err)
	emptyRootRecord := root.NewContractRecord(root.Interface, coretypes.AgentID{})
	require.EqualValues(ch.Glb.T, root.EncodeContractRecord(&emptyRootRecord), root.EncodeContractRecord(rootRec))

	accountsRec, err := ch.FindContract(accounts.Interface.Name)
	require.NoError(ch.Glb.T, err)
	require.EqualValues(ch.Glb.T, accounts.Interface.Name, accountsRec.Name)
	require.EqualValues(ch.Glb.T, accounts.Interface.Description, accountsRec.Description)
	require.EqualValues(ch.Glb.T, accounts.Interface.ProgramHash, accountsRec.ProgramHash)
	require.EqualValues(ch.Glb.T, ch.OriginatorAgentID, accountsRec.Creator)

	blobRec, err := ch.FindContract(blob.Interface.Name)
	require.NoError(ch.Glb.T, err)
	require.EqualValues(ch.Glb.T, blob.Interface.Name, blobRec.Name)
	require.EqualValues(ch.Glb.T, blob.Interface.Description, blobRec.Description)
	require.EqualValues(ch.Glb.T, blob.Interface.ProgramHash, blobRec.ProgramHash)
	require.EqualValues(ch.Glb.T, ch.OriginatorAgentID, blobRec.Creator)

	chainlogRec, err := ch.FindContract(eventlog.Interface.Name)
	require.NoError(ch.Glb.T, err)
	require.EqualValues(ch.Glb.T, eventlog.Interface.Name, chainlogRec.Name)
	require.EqualValues(ch.Glb.T, eventlog.Interface.Description, chainlogRec.Description)
	require.EqualValues(ch.Glb.T, eventlog.Interface.ProgramHash, chainlogRec.ProgramHash)
	require.EqualValues(ch.Glb.T, ch.OriginatorAgentID, chainlogRec.Creator)

	ch.CheckAccountLedger()
}

// CheckAccountLedger check integrity of the on-chain ledger.
// Sum of all accounts must be equal to total assets
func (ch *Chain) CheckAccountLedger() {
	total := ch.GetTotalAssets()
	accounts := ch.GetAccounts()
	sum := make(map[balance.Color]int64)
	for _, acc := range accounts {
		ch.GetAccountBalance(acc).AddToMap(sum)
	}
	require.True(ch.Glb.T, total.Equal(cbalances.NewFromMap(sum)))
}

// AssertAccountBalance asserts the on-chain account balance controlled by agentID for specific color
func (ch *Chain) AssertAccountBalance(agentID coretypes.AgentID, col balance.Color, bal int64) {
	require.EqualValues(ch.Glb.T, bal, ch.GetAccountBalance(agentID).Balance(col))
}
