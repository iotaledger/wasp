// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"testing"
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/coreblocklog"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmsolo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupBlockLog(t *testing.T) *wasmsolo.SoloContext {
	ctx := setup(t)
	ctx = ctx.SoloContextForCore(t, coreblocklog.ScName, coreblocklog.OnLoad)
	require.NoError(t, ctx.Err)
	return ctx
}

func TestBlockLogXxx(t *testing.T) {
	ctx := setupBlockLog(t)
	require.NoError(t, ctx.Err)
}

func TestControlAddresses(t *testing.T) {
	ctx := setupBlockLog(t)
	require.NoError(t, ctx.Err)

	f := coreblocklog.ScFuncs.ControlAddresses(ctx)
	f.Func.Call()
	require.NoError(t, ctx.Err)
	// solo agent is the chain owner here, so it is both the state controller address and governing address
	assert.Equal(t, ctx.Cvt.ScAddress(ctx.Chain.StateControllerAddress), f.Results.StateControllerAddress().Value())
	assert.Equal(t, ctx.Cvt.ScAddress(ctx.Chain.StateControllerAddress), f.Results.GoverningAddress().Value())
	// solo agent is set as state controller address and governing address from the beginning of the chain
	assert.Equal(t, uint32(0), f.Results.BlockIndex().Value())
}

func TestGetBlockInfo(t *testing.T) {
	ctx := setupBlockLog(t)
	require.NoError(t, ctx.Err)

	f := coreblocklog.ScFuncs.GetBlockInfo(ctx)
	f.Params.BlockIndex().SetValue(0)
	f.Func.Call()
	require.NoError(t, ctx.Err)
	b := f.Results.BlockInfo().Value()
	blockinfo, err := blocklog.BlockInfoFromBytes(5, b)
	require.NoError(t, err)
	assert.Equal(t, uint32(0), blockinfo.BlockIndex)
	assert.Equal(t, time.Unix(1, 4000000), blockinfo.Timestamp)
	assert.Equal(t, uint16(1), blockinfo.TotalRequests)
	assert.Equal(t, uint16(1), blockinfo.NumSuccessfulRequests)
	assert.Equal(t, uint16(0), blockinfo.NumOffLedgerRequests)
	// assert.Equal(t, , blockinfo.PreviousL1Commitment) // FIXME: can't generate the expected object
	assert.Nil(t, blockinfo.L1Commitment)
	assert.Equal(t, iotago.TransactionID{}, blockinfo.AnchorTransactionID)
	assert.Equal(t, uint64(0), blockinfo.TotalIotasInL2Accounts)
	assert.Equal(t, uint64(0), blockinfo.TotalDustDeposit)
	assert.Equal(t, uint64(0), blockinfo.GasBurned)
	assert.Equal(t, uint64(0), blockinfo.GasFeeCharged)
}

func TestGetLatestBlockInfo(t *testing.T) {
	ctx := setupBlockLog(t)
	require.NoError(t, ctx.Err)

	f := coreblocklog.ScFuncs.GetLatestBlockInfo(ctx)
	f.Func.Call()
	require.NoError(t, ctx.Err)
	index := f.Results.BlockIndex().Value()
	assert.Equal(t, uint32(5), index)

	b := f.Results.BlockInfo().Value()
	blockinfo, err := blocklog.BlockInfoFromBytes(5, b)
	require.NoError(t, err)
	assert.Equal(t, uint32(5), blockinfo.BlockIndex)
	assert.Equal(t, time.Unix(1, 11000001), blockinfo.Timestamp)
	assert.Equal(t, uint16(1), blockinfo.TotalRequests)
	assert.Equal(t, uint16(1), blockinfo.NumSuccessfulRequests)
	assert.Equal(t, uint16(0), blockinfo.NumOffLedgerRequests)
	// assert.Equal(t, , blockinfo.PreviousL1Commitment) // FIXME: can't generate the expected object
	assert.Nil(t, blockinfo.L1Commitment)
	assert.Equal(t, iotago.TransactionID{}, blockinfo.AnchorTransactionID)
	assert.Equal(t, uint64(0x3d0acc), blockinfo.TotalIotasInL2Accounts)
	assert.Equal(t, uint64(0x11b), blockinfo.TotalDustDeposit)
	assert.Equal(t, uint64(0x2710), blockinfo.GasBurned)
	assert.Equal(t, uint64(0x64), blockinfo.GasFeeCharged)
}

func TestGetRequestIDsForBlock(t *testing.T) {
	ctx := setupBlockLog(t)
	require.NoError(t, ctx.Err)

	f := coreblocklog.ScFuncs.GetRequestIDsForBlock(ctx)
	f.Params.BlockIndex().SetValue(0)
	f.Func.Call()
	require.NoError(t, ctx.Err)

	// FIXME: check result
}

func TestGetRequestReceipt(t *testing.T) {
	ctx := setupBlockLog(t)
	require.NoError(t, ctx.Err)

	f := coreblocklog.ScFuncs.GetRequestReceipt(ctx)
	f.Func.Call()
	require.NoError(t, ctx.Err)

	// FIXME: check result
}

func TestGetRequestReceiptsForBlock(t *testing.T) {
	ctx := setupBlockLog(t)
	require.NoError(t, ctx.Err)

	f := coreblocklog.ScFuncs.GetRequestReceiptsForBlock(ctx)
	f.Func.Call()
	require.NoError(t, ctx.Err)

	// FIXME: check result
}

func TestIsRequestProcessed(t *testing.T) {
	ctx := setupBlockLog(t)
	require.NoError(t, ctx.Err)

	f := coreblocklog.ScFuncs.IsRequestProcessed(ctx)
	f.Func.Call()
	require.NoError(t, ctx.Err)

	// FIXME: check result
}

func TestGetEventsForRequest(t *testing.T) {
	ctx := setupBlockLog(t)
	require.NoError(t, ctx.Err)

	f := coreblocklog.ScFuncs.GetEventsForRequest(ctx)
	f.Func.Call()
	require.NoError(t, ctx.Err)

	// FIXME: check result
}

func TestGetEventsForBlock(t *testing.T) {
	ctx := setupBlockLog(t)
	require.NoError(t, ctx.Err)

	f := coreblocklog.ScFuncs.GetEventsForBlock(ctx)
	f.Func.Call()
	require.NoError(t, ctx.Err)

	// FIXME: check result
}

func TestGetEventsForContract(t *testing.T) {
	ctx := setupBlockLog(t)
	require.NoError(t, ctx.Err)

	f := coreblocklog.ScFuncs.GetEventsForContract(ctx)
	f.Func.Call()
	require.NoError(t, ctx.Err)

	// FIXME: check result
}
