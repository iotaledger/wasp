// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"testing"
	"time"

	"github.com/iotaledger/wasp/contracts/wasm/fairroulette/go/fairrouletteimpl"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/contracts/wasm/fairroulette/go/fairroulette"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmsolo"
)

func setupTest(t *testing.T) *wasmsolo.SoloContext {
	return wasmsolo.NewSoloContext(t, fairroulette.ScName, fairrouletteimpl.OnDispatch)
}

func TestDeploy(t *testing.T) {
	ctx := setupTest(t)
	require.NoError(t, ctx.ContractExists(fairroulette.ScName))
}

func TestBets(t *testing.T) {
	ctx := setupTest(t)
	var better [10]*wasmsolo.SoloAgent
	for i := 0; i < 10; i++ {
		better[i] = ctx.NewSoloAgent()
		placeBet := fairroulette.ScFuncs.PlaceBet(ctx.Sign(better[i]))
		placeBet.Params.Number().SetValue(3)
		placeBet.Func.TransferBaseTokens(1234).Post()
		require.NoError(t, ctx.Err)
	}

	// wait for finalize_auction
	ctx.AdvanceClockBy(1201 * time.Second)
	require.True(t, ctx.WaitForPendingRequests(1))
}
