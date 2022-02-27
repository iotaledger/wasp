// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"testing"
	"time"

	"github.com/iotaledger/wasp/contracts/wasm/fairroulette/go/fairroulette"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmsolo"
	"github.com/stretchr/testify/require"
)

func setupTest(t *testing.T) *wasmsolo.SoloContext {
	return wasmsolo.NewSoloContext(t, fairroulette.ScName, fairroulette.OnLoad)
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
		placeBet.Func.TransferIotas(25).Post()
		require.NoError(t, ctx.Err)
	}
	// TODO this should be a simple 1 request to wait for, but sometimes
	// the payout will have already been triggered (bug), so instead of
	// waiting for that single payout request we will (erroneously) wait
	// for the inbuf and outbuf counts to equalize
	info := ctx.Chain.MempoolInfo()

	// wait for finalize_auction
	ctx.AdvanceClockBy(1201 * time.Second)
	require.True(t, ctx.WaitForPendingRequests(-info.InBufCounter))
}
