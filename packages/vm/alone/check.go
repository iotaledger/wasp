// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package alone

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/cbalances"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/accountsc"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/blob"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/root"
	"github.com/stretchr/testify/require"
)

func (glb *Glb) CheckUtxodbBalance(addr address.Address, col balance.Color, expected int64) {
	require.EqualValues(glb.T, expected, glb.GetUtxodbBalance(addr, col))
}

func (ch *Chain) CheckBase() {
	req := NewCall(root.Interface.Name, root.FuncGetInfo)
	res1, err := ch.PostRequest(req, nil)
	require.NoError(ch.Glb.T, err)

	res2, err := ch.CallViewFull(req)
	require.NoError(ch.Glb.T, err)

	require.EqualValues(ch.Glb.T, res1.Hash(), res2.Hash())

	rootRec, err := ch.FindContract(root.Interface.Name)
	require.NoError(ch.Glb.T, err)
	require.EqualValues(ch.Glb.T, root.EncodeContractRecord(&root.RootContractRecord), root.EncodeContractRecord(rootRec))

	accountsRec, err := ch.FindContract(accountsc.Interface.Name)
	require.NoError(ch.Glb.T, err)
	require.EqualValues(ch.Glb.T, accountsc.Interface.Name, accountsRec.Name)
	require.EqualValues(ch.Glb.T, accountsc.Interface.Description, accountsRec.Description)
	require.EqualValues(ch.Glb.T, accountsc.Interface.ProgramHash, accountsRec.ProgramHash)
	require.EqualValues(ch.Glb.T, ch.OriginatorAgentID, accountsRec.Creator)

	blobRec, err := ch.FindContract(blob.Interface.Name)
	require.NoError(ch.Glb.T, err)
	require.EqualValues(ch.Glb.T, blob.Interface.Name, blobRec.Name)
	require.EqualValues(ch.Glb.T, blob.Interface.Description, blobRec.Description)
	require.EqualValues(ch.Glb.T, blob.Interface.ProgramHash, blobRec.ProgramHash)
	require.EqualValues(ch.Glb.T, ch.OriginatorAgentID, blobRec.Creator)
}

func (ch *Chain) CheckAccountLedger() {
	total := ch.GetTotalAssets()
	accounts := ch.GetAccounts()
	sum := make(map[balance.Color]int64)
	for _, acc := range accounts {
		ch.GetAccountBalance(acc).AddToMap(sum)
	}
	require.True(ch.Glb.T, total.Equal(cbalances.NewFromMap(sum)))
}

func (ch *Chain) CheckAccountBalance(agentID coretypes.AgentID, col balance.Color, bal int64) {
	require.EqualValues(ch.Glb.T, bal, ch.GetAccountBalance(agentID).Balance(col))
}
