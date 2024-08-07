package blocklog

import (
	"time"

	"github.com/samber/lo"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/vm/core/errors/coreerrors"
)

var Processor = Contract.Processor(nil,
	ViewGetBlockInfo.WithHandler(viewGetBlockInfo),
	ViewGetEventsForBlock.WithHandler(viewGetEventsForBlock),
	ViewGetEventsForRequest.WithHandler(viewGetEventsForRequest),
	ViewGetRequestIDsForBlock.WithHandler(viewGetRequestIDsForBlock),
	ViewGetRequestReceipt.WithHandler(viewGetRequestReceipt),
	ViewGetRequestReceiptsForBlock.WithHandler(viewGetRequestReceiptsForBlock),
	ViewIsRequestProcessed.WithHandler(viewIsRequestProcessed),
)

var ErrBlockNotFound = coreerrors.Register("Block not found").Create()

func (s *StateWriter) SetInitialState() {
	s.SaveNextBlockInfo(&BlockInfo{
		SchemaVersion:         BlockInfoLatestSchemaVersion,
		Timestamp:             time.Time{},
		TotalRequests:         1,
		NumSuccessfulRequests: 1,
		NumOffLedgerRequests:  0,
	})
}

// viewGetBlockInfo returns blockInfo for a given block.
func viewGetBlockInfo(ctx isc.SandboxView, blockIndexOptional *uint32) (uint32, *BlockInfo) {
	blockIndex := getBlockIndexParams(ctx, blockIndexOptional)
	b, ok := NewStateReaderFromSandbox(ctx).GetBlockInfo(blockIndex)
	if !ok {
		panic(ErrBlockNotFound)
	}
	return blockIndex, b
}

var errNotFound = coreerrors.Register("not found").Create()

// viewGetRequestIDsForBlock returns a list of requestIDs for a given block.
func viewGetRequestIDsForBlock(ctx isc.SandboxView, blockIndexOptional *uint32) (uint32, []isc.RequestID) {
	blockIndex := getBlockIndexParams(ctx, blockIndexOptional)

	if blockIndex == 0 {
		// block 0 is an empty state
		return blockIndex, nil
	}

	receipts, found := NewStateReaderFromSandbox(ctx).getRequestLogRecordsForBlockBin(blockIndex)
	if !found {
		panic(errNotFound)
	}

	return blockIndex, lo.Map(receipts, func(b []byte, i int) isc.RequestID {
		receipt, err := RequestReceiptFromBytes(b, blockIndex, uint16(i))
		ctx.RequireNoError(err)
		return receipt.Request.ID()
	})
}

func viewGetRequestReceipt(ctx isc.SandboxView, reqID isc.RequestID) *RequestReceipt {
	rec, err := NewStateReaderFromSandbox(ctx).GetRequestRecordDataByRequestID(reqID)
	ctx.RequireNoError(err)
	return rec
}

// viewGetRequestReceiptsForBlock returns a list of receipts for a given block.
func viewGetRequestReceiptsForBlock(ctx isc.SandboxView, blockIndexOptional *uint32) (uint32, []*RequestReceipt) {
	blockIndex := getBlockIndexParams(ctx, blockIndexOptional)

	if blockIndex == 0 {
		// block 0 is an empty state
		return 0, nil
	}

	receipts, found := NewStateReaderFromSandbox(ctx).getRequestLogRecordsForBlockBin(blockIndex)
	if !found {
		panic(errNotFound)
	}

	return blockIndex, lo.Map(receipts, func(b []byte, i int) *RequestReceipt {
		receipt, err := RequestReceiptFromBytes(b, blockIndex, uint16(i))
		ctx.RequireNoError(err)
		return receipt
	})
}

func viewIsRequestProcessed(ctx isc.SandboxView, requestID isc.RequestID) bool {
	requestReceipt, err := NewStateReaderFromSandbox(ctx).GetRequestReceipt(requestID)
	ctx.RequireNoError(err)
	return requestReceipt != nil
}

// viewGetEventsForRequest returns a list of events for a given request.
func viewGetEventsForRequest(ctx isc.SandboxView, requestID isc.RequestID) []*isc.Event {
	events, err := NewStateReaderFromSandbox(ctx).getRequestEventsInternal(requestID)
	ctx.RequireNoError(err)
	return lo.Map(events, func(b []byte, _ int) *isc.Event {
		return lo.Must(isc.EventFromBytes(b))
	})
}

// viewGetEventsForBlock returns a list of events for a given block.
func viewGetEventsForBlock(ctx isc.SandboxView, blockIndexOptional *uint32) (uint32, []*isc.Event) {
	blockIndex := getBlockIndexParams(ctx, blockIndexOptional)

	if blockIndex == 0 {
		// block 0 is an empty state
		return 0, nil
	}

	state := NewStateReaderFromSandbox(ctx)
	blockInfo, ok := state.GetBlockInfo(blockIndex)
	ctx.Requiref(ok, "block not found: %d", blockIndex)
	events := state.GetEventsByBlockIndex(blockIndex, blockInfo.TotalRequests)

	return blockIndex, lo.Map(events, func(b []byte, _ int) *isc.Event {
		return lo.Must(isc.EventFromBytes(b))
	})
}
