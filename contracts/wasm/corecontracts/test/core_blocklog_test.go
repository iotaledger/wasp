// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"testing"

	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/coreblocklog"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmtypes"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmsolo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupBlockLog(t *testing.T) *wasmsolo.SoloContext {
	ctx := setup(t)
	ctx = ctx.SoloContextForCore(t, coreblocklog.ScName, coreblocklog.OnDispatch)
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

	for i := uint32(0); i < 6; i++ {
		f := coreblocklog.ScFuncs.GetBlockInfo(ctx)
		f.Params.BlockIndex().SetValue(i)
		f.Func.Call()
		require.NoError(t, ctx.Err)
		b := f.Results.BlockInfo().Value()
		blockinfo, err := blocklog.BlockInfoFromBytes(i, b)
		require.NoError(t, err)

		expectBlockInfo, err := ctx.Chain.GetBlockInfo(i)
		require.NoError(t, err)

		assert.Equal(t, expectBlockInfo.BlockIndex, blockinfo.BlockIndex)
		assert.Equal(t, expectBlockInfo.Timestamp, blockinfo.Timestamp)
		assert.Equal(t, expectBlockInfo.TotalRequests, blockinfo.TotalRequests)
		assert.Equal(t, expectBlockInfo.NumSuccessfulRequests, blockinfo.NumSuccessfulRequests)
		assert.Equal(t, expectBlockInfo.NumOffLedgerRequests, blockinfo.NumOffLedgerRequests)
		assert.Equal(t, expectBlockInfo.PreviousL1Commitment, blockinfo.PreviousL1Commitment)
		assert.Equal(t, expectBlockInfo.L1Commitment, blockinfo.L1Commitment)
		assert.Equal(t, expectBlockInfo.AnchorTransactionID, blockinfo.AnchorTransactionID)
		assert.Equal(t, expectBlockInfo.TotalBaseTokensInL2Accounts, blockinfo.TotalBaseTokensInL2Accounts)
		assert.Equal(t, expectBlockInfo.TotalStorageDeposit, blockinfo.TotalStorageDeposit)
		assert.Equal(t, expectBlockInfo.GasBurned, blockinfo.GasBurned)
		assert.Equal(t, expectBlockInfo.GasFeeCharged, blockinfo.GasFeeCharged)
	}
}

func TestGetLatestBlockInfo(t *testing.T) {
	ctx := setupBlockLog(t)
	require.NoError(t, ctx.Err)

	expectBlockInfo, err := ctx.Chain.GetBlockInfo()
	require.NoError(t, err)
	f := coreblocklog.ScFuncs.GetBlockInfo(ctx)
	f.Func.Call()
	require.NoError(t, ctx.Err)
	index := f.Results.BlockIndex().Value()
	assert.Equal(t, expectBlockInfo.BlockIndex, index)

	blockinfo, err := blocklog.BlockInfoFromBytes(5, f.Results.BlockInfo().Value())
	require.NoError(t, err)
	assert.Equal(t, expectBlockInfo.BlockIndex, blockinfo.BlockIndex)
	assert.Equal(t, expectBlockInfo.Timestamp, blockinfo.Timestamp)
	assert.Equal(t, expectBlockInfo.TotalRequests, blockinfo.TotalRequests)
	assert.Equal(t, expectBlockInfo.NumSuccessfulRequests, blockinfo.NumSuccessfulRequests)
	assert.Equal(t, expectBlockInfo.NumOffLedgerRequests, blockinfo.NumOffLedgerRequests)
	assert.Equal(t, expectBlockInfo.PreviousL1Commitment, blockinfo.PreviousL1Commitment)
	assert.Equal(t, expectBlockInfo.L1Commitment, blockinfo.L1Commitment)
	assert.Equal(t, expectBlockInfo.AnchorTransactionID, blockinfo.AnchorTransactionID)
	assert.Equal(t, expectBlockInfo.TotalBaseTokensInL2Accounts, blockinfo.TotalBaseTokensInL2Accounts)
	assert.Equal(t, expectBlockInfo.TotalStorageDeposit, blockinfo.TotalStorageDeposit)
	assert.Equal(t, expectBlockInfo.GasBurned, blockinfo.GasBurned)
	assert.Equal(t, expectBlockInfo.GasFeeCharged, blockinfo.GasFeeCharged)
}

func TestGetRequestIDsForBlock(t *testing.T) {
	ctx := setupBlockLog(t)
	require.NoError(t, ctx.Err)

	f := coreblocklog.ScFuncs.GetRequestIDsForBlock(ctx)
	for blockNum := uint32(0); blockNum < 6; blockNum++ {
		f.Params.BlockIndex().SetValue(blockNum)
		f.Func.Call()
		require.NoError(t, ctx.Err)
		reqs := ctx.Chain.GetRequestIDsForBlock(blockNum)
		assert.Equal(t, uint32(len(reqs)), f.Results.RequestID().Length())
		for reqNum := uint32(0); reqNum < uint32(len(reqs)); reqNum++ {
			assert.Equal(t, reqs[reqNum].Bytes(), f.Results.RequestID().GetRequestID(reqNum).Value().Bytes())
		}
	}
}

func TestGetRequestReceipt(t *testing.T) {
	ctx := setupBlockLog(t)
	require.NoError(t, ctx.Err)

	blockIndex := uint32(3)
	reqIndex := uint16(0)
	reqs := ctx.Chain.GetRequestIDsForBlock(blockIndex)
	f := coreblocklog.ScFuncs.GetRequestReceipt(ctx)
	f.Params.RequestID().SetValue(ctx.Cvt.ScRequestID(reqs[0]))
	f.Func.Call()
	require.NoError(t, ctx.Err)
	assert.Equal(t, blockIndex, f.Results.BlockIndex().Value())
	assert.Equal(t, reqIndex, f.Results.RequestIndex().Value())

	receipt, err := blocklog.RequestReceiptFromBytes(f.Results.RequestRecord().Value())
	require.NoError(t, err)
	soloreceipt, err := ctx.Chain.GetRequestReceipt(reqs[0])
	assert.Nil(t, err)

	// note: this is what ctx.Chain.GetRequestReceipt() does as well,
	// so we better make sure they are equal before comparing
	receipt.BlockIndex = f.Results.BlockIndex().Value()
	receipt.RequestIndex = f.Results.RequestIndex().Value()
	assert.Equal(t, soloreceipt, receipt)
}

func TestGetRequestReceiptsForBlock(t *testing.T) {
	ctx := setupBlockLog(t)
	require.NoError(t, ctx.Err)

	for i := uint32(0); i < 5; i++ {
		f := coreblocklog.ScFuncs.GetRequestReceiptsForBlock(ctx)
		f.Params.BlockIndex().SetValue(i)
		f.Func.Call()
		require.NoError(t, ctx.Err)

		soloreceipts := ctx.Chain.GetRequestReceiptsForBlock(i)
		recNum := f.Results.RequestRecord().Length()
		for j := uint32(0); j < recNum; j++ {
			receipt, err := blocklog.RequestReceiptFromBytes(f.Results.RequestRecord().GetBytes(j).Value())
			require.NoError(t, err)
			receipt.BlockIndex = soloreceipts[j].BlockIndex
			assert.Equal(t, soloreceipts[j], receipt)
		}
	}
}

func TestIsRequestProcessed(t *testing.T) {
	ctx := setupBlockLog(t)
	require.NoError(t, ctx.Err)

	blockIndex := uint32(3)
	reqs := ctx.Chain.GetRequestIDsForBlock(blockIndex)

	f := coreblocklog.ScFuncs.IsRequestProcessed(ctx)
	f.Params.RequestID().SetValue(ctx.Cvt.ScRequestID(reqs[0]))
	f.Func.Call()
	require.NoError(t, ctx.Err)
	require.Equal(t, ctx.Chain.IsRequestProcessed(reqs[0]), f.Results.RequestProcessed().Value())

	notExistReqID := wasmtypes.RequestIDFromString("0xcc025a91fe7f071a7a53a1db5257d161d666d4aa1606422a3b3553c2b8b904e70000")
	f.Params.RequestID().SetValue(notExistReqID)
	f.Func.Call()
	require.NoError(t, ctx.Err)
	require.False(t, f.Results.RequestProcessed().Value())
}

func TestGetEventsForRequest(t *testing.T) {
	ctx := setupBlockLog(t)
	require.NoError(t, ctx.Err)

	blockIndex := uint32(3)
	reqs := ctx.Chain.GetRequestIDsForBlock(blockIndex)

	f := coreblocklog.ScFuncs.GetEventsForRequest(ctx)
	f.Params.RequestID().SetValue(ctx.Cvt.ScRequestID(reqs[0]))
	f.Func.Call()
	require.NoError(t, ctx.Err)

	events, err := ctx.Chain.GetEventsForRequest(reqs[0])
	require.NoError(t, err)
	assert.Equal(t, uint32(len(events)), f.Results.Event().Length())
	for i := uint32(0); i < uint32(len(events)); i++ {
		assert.Equal(t, []byte(events[i]), f.Results.Event().GetBytes(i).Value())
	}
}

func TestGetEventsForBlock(t *testing.T) {
	ctx := setupBlockLog(t)
	require.NoError(t, ctx.Err)

	for blockIndex := uint32(0); blockIndex < 5; blockIndex++ {
		f := coreblocklog.ScFuncs.GetEventsForBlock(ctx)
		f.Params.BlockIndex().SetValue(blockIndex)
		f.Func.Call()
		require.NoError(t, ctx.Err)
		events, err := ctx.Chain.GetEventsForBlock(blockIndex)
		require.NoError(t, err)
		assert.Equal(t, uint32(len(events)), f.Results.Event().Length())
		for i := uint32(0); i < uint32(len(events)); i++ {
			assert.Equal(t, []byte(events[i]), f.Results.Event().GetBytes(i).Value())
		}
	}
}

func TestGetEventsForContract(t *testing.T) {
	ctx := setupBlockLog(t)
	require.NoError(t, ctx.Err)

	f := coreblocklog.ScFuncs.GetEventsForContract(ctx)
	f.Params.ContractHname().SetValue(coreblocklog.HScName)
	f.Params.FromBlock().SetValue(0)
	f.Params.ToBlock().SetValue(5)
	f.Func.Call()
	require.NoError(t, ctx.Err)

	events, err := ctx.Chain.GetEventsForContract(coreblocklog.ScName)
	require.NoError(t, err)
	assert.Equal(t, uint32(len(events)), f.Results.Event().Length())
	for i := uint32(0); i < uint32(len(events)); i++ {
		assert.Equal(t, []byte(events[i]), f.Results.Event().GetBytes(i).Value())
	}
}
