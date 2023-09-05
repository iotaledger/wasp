package blocklog

import (
	"math"
	"time"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/errors/coreerrors"
)

var Processor = Contract.Processor(nil,
	ViewGetBlockInfo.WithHandler(viewGetBlockInfo),
	ViewGetEventsForBlock.WithHandler(viewGetEventsForBlock),
	ViewGetEventsForContract.WithHandler(viewGetEventsForContract),
	ViewGetEventsForRequest.WithHandler(viewGetEventsForRequest),
	ViewGetRequestIDsForBlock.WithHandler(viewGetRequestIDsForBlock),
	ViewGetRequestReceipt.WithHandler(viewGetRequestReceipt),
	ViewGetRequestReceiptsForBlock.WithHandler(viewGetRequestReceiptsForBlock),
	ViewIsRequestProcessed.WithHandler(viewIsRequestProcessed),
	ViewHasUnprocessable.WithHandler(viewHasUnprocessable),

	FuncRetryUnprocessable.WithHandler(retryUnprocessable),
)

var ErrBlockNotFound = coreerrors.Register("Block not found").Create()

func SetInitialState(s kv.KVStore) {
	SaveNextBlockInfo(s, &BlockInfo{
		SchemaVersion:         BlockInfoLatestSchemaVersion,
		Timestamp:             time.Time{},
		TotalRequests:         1,
		NumSuccessfulRequests: 1,
		NumOffLedgerRequests:  0,
	})
}

// viewGetBlockInfo returns blockInfo for a given block.
// params:
// ParamBlockIndex - index of the block (defaults to the latest block)
func viewGetBlockInfo(ctx isc.SandboxView) dict.Dict {
	blockIndex := getBlockIndexParams(ctx)
	b := getBlockInfoBytes(ctx.StateR(), blockIndex)
	if b == nil {
		panic(ErrBlockNotFound)
	}
	return dict.Dict{
		ParamBlockIndex: codec.EncodeUint32(blockIndex),
		ParamBlockInfo:  b,
	}
}

var errNotFound = coreerrors.Register("not found").Create()

// viewGetRequestIDsForBlock returns a list of requestIDs for a given block.
// params:
// ParamBlockIndex - index of the block (defaults to latest block)
func viewGetRequestIDsForBlock(ctx isc.SandboxView) dict.Dict {
	blockIndex := getBlockIndexParams(ctx)

	if blockIndex == 0 {
		// block 0 is an empty state
		return nil
	}

	receipts, found := getRequestLogRecordsForBlockBin(ctx.StateR(), blockIndex)
	if !found {
		panic(errNotFound)
	}

	ret := dict.New()
	requestIDs := collections.NewArray(ret, ParamRequestID)
	for i, receipt := range receipts {
		requestReceipt, err := RequestReceiptFromBytes(receipt, blockIndex, uint16(i))
		ctx.RequireNoError(err)
		requestIDs.Push(requestReceipt.Request.ID().Bytes())
	}
	ret.Set(ParamBlockIndex, codec.Encode(blockIndex))
	return ret
}

func viewGetRequestReceipt(ctx isc.SandboxView) dict.Dict {
	requestID := ctx.Params().MustGetRequestID(ParamRequestID)
	res, err := GetRequestRecordDataByRequestID(ctx.StateR(), requestID)
	ctx.RequireNoError(err)
	if res == nil {
		return nil
	}
	return dict.Dict{
		ParamRequestRecord: res.ReceiptBin,
		ParamBlockIndex:    codec.EncodeUint32(res.BlockIndex),
		ParamRequestIndex:  codec.EncodeUint16(res.RequestIndex),
	}
}

// viewGetRequestReceiptsForBlock returns a list of receipts for a given block.
// params:
// ParamBlockIndex - index of the block (defaults to latest block)
func viewGetRequestReceiptsForBlock(ctx isc.SandboxView) dict.Dict {
	blockIndex := getBlockIndexParams(ctx)

	if blockIndex == 0 {
		// block 0 is an empty state
		return nil
	}

	receipts, found := getRequestLogRecordsForBlockBin(ctx.StateR(), blockIndex)
	if !found {
		panic(errNotFound)
	}

	ret := dict.New()
	requestReceipts := collections.NewArray(ret, ParamRequestRecord)
	for _, receipt := range receipts {
		requestReceipts.Push(receipt)
	}
	ret.Set(ParamBlockIndex, codec.Encode(blockIndex))
	return ret
}

func viewIsRequestProcessed(ctx isc.SandboxView) dict.Dict {
	requestID := ctx.Params().MustGetRequestID(ParamRequestID)
	requestReceipt, err := getRequestReceipt(ctx.StateR(), requestID)
	ctx.RequireNoError(err)
	return dict.Dict{
		ParamRequestProcessed: codec.EncodeBool(requestReceipt != nil),
	}
}

// viewGetEventsForRequest returns a list of events for a given request.
// params:
// ParamRequestID - requestID
func viewGetEventsForRequest(ctx isc.SandboxView) dict.Dict {
	requestID := ctx.Params().MustGetRequestID(ParamRequestID)
	events, err := getRequestEventsInternal(ctx.StateR(), requestID)
	ctx.RequireNoError(err)
	return eventsToDict(events)
}

// viewGetEventsForBlock returns a list of events for a given block.
// params:
// ParamBlockIndex - index of the block (defaults to latest block)
func viewGetEventsForBlock(ctx isc.SandboxView) dict.Dict {
	blockIndex := getBlockIndexParams(ctx)

	if blockIndex == 0 {
		// block 0 is an empty state
		return nil
	}

	stateR := ctx.StateR()
	blockInfo, ok := GetBlockInfo(stateR, blockIndex)
	ctx.Requiref(ok, "block not found: %d", blockIndex)
	events := GetEventsByBlockIndex(stateR, blockIndex, blockInfo.TotalRequests)

	ret := eventsToDict(events)
	ret.Set(ParamBlockIndex, codec.Encode(blockIndex))
	return ret
}

// viewGetEventsForContract returns a list of events for a given smart contract.
// params:
// ParamContractHname - hname of the contract
// ParamFromBlock - defaults to 0
// ParamToBlock - defaults to latest block
func viewGetEventsForContract(ctx isc.SandboxView) dict.Dict {
	params := ctx.Params()
	contract := params.MustGetHname(ParamContractHname)
	fromBlock := params.MustGetUint32(ParamFromBlock, 0)
	toBlock := params.MustGetUint32(ParamToBlock, math.MaxUint32)
	events := getSmartContractEventsInternal(ctx.StateR(), contract, fromBlock, toBlock)

	return eventsToDict(events)
}
