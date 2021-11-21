// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"testing"

	"github.com/iotaledger/wasp/contracts/wasm/myworld/go/myworld"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/wasmlib/go/wasmlib"
	"github.com/iotaledger/wasp/packages/vm/wasmsolo"
	"github.com/stretchr/testify/require"
)

var (
	peer       *wasmsolo.SoloAgent
	tokenColor wasmlib.ScColor
)

func TestDeploy(t *testing.T) {
	ctx := wasmsolo.NewSoloContext(t, myworld.ScName, myworld.OnLoad)
	require.NoError(t, ctx.ContractExists(myworld.ScName))

	// set up account and mint some tokens
	peer = ctx.NewSoloAgent()
	tokenColor, ctx.Err = peer.Mint(10)
	require.NoError(t, ctx.Err)
	require.EqualValues(t, solo.Saldo-10, peer.Balance())
	require.EqualValues(t, 10, peer.Balance(tokenColor))

	var newTreasure myworld.Treasure
	newTreasure.Amount = 500
	newTreasure.Name = "Test"
	newTreasure.Owner = peer.ScAgentID()

	depositTreasure := myworld.ScFuncs.DepositTreasure(ctx.Sign(peer))
	depositTreasure.Params.Treasure().SetValue(&newTreasure)
	depositTreasure.Func.TransferIotas(500).Post()
	require.NoError(t, ctx.Err)

	getAllTreasures := myworld.ScFuncs.GetAllTreasures(ctx)
	getAllTreasures.Func.Call()
	require.NoError(t, ctx.Err)

	firstStoredTreasure := getAllTreasures.Results.Treasures().GetTreasure(0)
	t.Logf("Success on Treasure from %s", firstStoredTreasure.Value().Name)
}
