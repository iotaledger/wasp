// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/coreblocklog"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmtypes"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmsolo"
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

func TestGetBlockInfo(t *testing.T) {
	ctx := setupBlockLog(t)
	require.NoError(t, ctx.Err)

	for i := uint32(0); i < 5; i++ {
		f := coreblocklog.ScFuncs.GetBlockInfo(ctx)
		f.Params.BlockIndex().SetValue(i)
		f.Func.Call()
		require.NoError(t, ctx.Err)
		b := f.Results.BlockInfo().Value()
		blockinfo, err := blocklog.BlockInfoFromBytes(b)
		require.NoError(t, err)

		expectBlockInfo, err := ctx.Chain.GetBlockInfo(i)
		require.NoError(t, err)

		require.Equal(t, expectBlockInfo.SchemaVersion, blockinfo.SchemaVersion)
		require.Equal(t, expectBlockInfo.BlockIndex(), blockinfo.BlockIndex())
		require.Equal(t, expectBlockInfo.Timestamp, blockinfo.Timestamp)
		require.Equal(t, expectBlockInfo.TotalRequests, blockinfo.TotalRequests)
		require.Equal(t, expectBlockInfo.NumSuccessfulRequests, blockinfo.NumSuccessfulRequests)
		require.Equal(t, expectBlockInfo.NumOffLedgerRequests, blockinfo.NumOffLedgerRequests)
		require.Equal(t, expectBlockInfo.PreviousAliasOutput, blockinfo.PreviousAliasOutput)
		require.Equal(t, expectBlockInfo.GasBurned, blockinfo.GasBurned)
		require.Equal(t, expectBlockInfo.GasFeeCharged, blockinfo.GasFeeCharged)
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
	require.Equal(t, expectBlockInfo.BlockIndex(), index)

	blockinfo, err := blocklog.BlockInfoFromBytes(f.Results.BlockInfo().Value())
	require.NoError(t, err)
	require.Equal(t, expectBlockInfo.SchemaVersion, blockinfo.SchemaVersion)
	require.Equal(t, expectBlockInfo.BlockIndex(), blockinfo.BlockIndex())
	require.Equal(t, expectBlockInfo.Timestamp, blockinfo.Timestamp)
	require.Equal(t, expectBlockInfo.TotalRequests, blockinfo.TotalRequests)
	require.Equal(t, expectBlockInfo.NumSuccessfulRequests, blockinfo.NumSuccessfulRequests)
	require.Equal(t, expectBlockInfo.NumOffLedgerRequests, blockinfo.NumOffLedgerRequests)
	require.Equal(t, expectBlockInfo.PreviousAliasOutput, blockinfo.PreviousAliasOutput)
	require.Equal(t, expectBlockInfo.GasBurned, blockinfo.GasBurned)
	require.Equal(t, expectBlockInfo.GasFeeCharged, blockinfo.GasFeeCharged)
}

func TestGetRequestIDsForBlock(t *testing.T) {
	ctx := setupBlockLog(t)
	require.NoError(t, ctx.Err)

	f := coreblocklog.ScFuncs.GetRequestIDsForBlock(ctx)
	for blockNum := uint32(0); blockNum < 5; blockNum++ {
		f.Params.BlockIndex().SetValue(blockNum)
		f.Func.Call()
		require.NoError(t, ctx.Err)
		reqs := ctx.Chain.GetRequestIDsForBlock(blockNum)
		require.Equal(t, uint32(len(reqs)), f.Results.RequestID().Length())
		for reqNum := uint32(0); reqNum < uint32(len(reqs)); reqNum++ {
			require.Equal(t, reqs[reqNum].Bytes(), f.Results.RequestID().GetRequestID(reqNum).Value().Bytes())
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
	require.Equal(t, blockIndex, f.Results.BlockIndex().Value())
	require.Equal(t, reqIndex, f.Results.RequestIndex().Value())

	receipt, err := blocklog.RequestReceiptFromBytes(
		f.Results.RequestReceipt().Value(),
		f.Results.BlockIndex().Value(),
		f.Results.RequestIndex().Value(),
	)
	require.NoError(t, err)
	soloreceipt, err := ctx.Chain.GetRequestReceipt(reqs[0])
	require.NoError(t, err)

	require.Equal(t, soloreceipt, receipt)
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
		receipts := f.Results.RequestReceipts()
		recNum := receipts.Length()
		for j := uint32(0); j < recNum; j++ {
			receipt, err := blocklog.RequestReceiptFromBytes(
				receipts.GetBytes(j).Value(),
				soloreceipts[j].BlockIndex,
				soloreceipts[j].RequestIndex,
			)
			require.NoError(t, err)
			require.Equal(t, soloreceipts[j], receipt)
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
	require.Equal(t, uint32(len(events)), f.Results.Event().Length())
	for i := uint32(0); i < uint32(len(events)); i++ {
		require.Equal(t, events[i].Bytes(), f.Results.Event().GetBytes(i).Value())
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
		require.Equal(t, uint32(len(events)), f.Results.Event().Length())
		for i := uint32(0); i < uint32(len(events)); i++ {
			require.Equal(t, events[i].Bytes(), f.Results.Event().GetBytes(i).Value())
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
	require.Equal(t, uint32(len(events)), f.Results.Event().Length())
	for i := uint32(0); i < uint32(len(events)); i++ {
		require.Equal(t, events[i].Bytes(), f.Results.Event().GetBytes(i).Value())
	}
}
