package blocklog

import (
	"math"
	"time"

	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

var Processor = Contract.Processor(initialize,
	FuncGetBlockInfo.WithHandler(viewGetBlockInfo),
	FuncGetLatestBlockInfo.WithHandler(viewGetLatestBlockInfo),
	FuncGetRequestReceipt.WithHandler(viewGetRequestReceipt),
	FuncGetRequestReceiptsForBlock.WithHandler(viewGetRequestReceiptsForBlock),
	FuncGetRequestIDsForBlock.WithHandler(viewGetRequestIDsForBlock),
	FuncIsRequestProcessed.WithHandler(viewIsRequestProcessed),
	FuncControlAddresses.WithHandler(viewControlAddresses),
	FuncGetEventsForRequest.WithHandler(viewGetEventsForRequest),
	FuncGetEventsForBlock.WithHandler(viewGetEventsForBlock),
	FuncGetEventsForContract.WithHandler(viewGetEventsForContract),
)

func initialize(ctx iscp.Sandbox) dict.Dict {
	blockIndex := SaveNextBlockInfo(ctx.State(), &BlockInfo{
		Timestamp:             time.Unix(0, ctx.Timestamp()),
		TotalRequests:         1,
		NumSuccessfulRequests: 1,
		NumOffLedgerRequests:  0,
	})
	ctx.Requiref(blockIndex == 0, "blocklog.initialize.fail: unexpected block index")
	ctx.Log().Debugf("blocklog.initialize.success hname = %s", Contract.Hname().String())
	return nil
}

// viewGetBlockInfo returns blockInfo for a given block.
// params:
// ParamBlockIndex - index of the block (defaults to latest block)
func viewGetBlockInfo(ctx iscp.SandboxView) dict.Dict {
	blockIndex := getBlockIndexParams(ctx)
	data, found, err := getBlockInfoDataInternal(ctx.State(), blockIndex)
	ctx.RequireNoError(err)
	ctx.Requiref(found, "not found")
	return dict.Dict{ParamBlockInfo: data}
}

func viewGetLatestBlockInfo(ctx iscp.SandboxView) dict.Dict {
	registry := collections.NewArray32ReadOnly(ctx.State(), prefixBlockRegistry)
	regLen := registry.MustLen()
	ctx.Requiref(regLen != 0, "blocklog::viewGetLatestBlockInfo: empty log")
	data := registry.MustGetAt(regLen - 1)
	return dict.Dict{
		ParamBlockIndex: codec.EncodeUint32(regLen - 1),
		ParamBlockInfo:  data,
	}
}

func viewIsRequestProcessed(ctx iscp.SandboxView) dict.Dict {
	requestID := ctx.ParamDecoder().MustGetRequestID(ParamRequestID)
	seen, err := isRequestProcessedInternal(ctx.State(), &requestID)
	ctx.RequireNoError(err)
	ret := dict.New()
	if seen {
		ret.Set(ParamRequestProcessed, codec.EncodeString("+"))
	}
	return ret
}

func viewGetRequestReceipt(ctx iscp.SandboxView) dict.Dict {
	requestID := ctx.ParamDecoder().MustGetRequestID(ParamRequestID)
	recBin, blockIndex, requestIndex, found := getRequestRecordDataByRequestID(ctx, requestID)
	if !found {
		return nil
	}
	return dict.Dict{
		ParamRequestRecord: recBin,
		ParamBlockIndex:    codec.EncodeUint32(blockIndex),
		ParamRequestIndex:  codec.EncodeUint16(requestIndex),
	}
}

// viewGetRequestIDsForBlock returns a list of requestIDs for a given block.
// params:
// ParamBlockIndex - index of the block (defaults to latest block)
func viewGetRequestIDsForBlock(ctx iscp.SandboxView) dict.Dict {
	blockIndex := getBlockIndexParams(ctx)

	if blockIndex == 0 {
		// block 0 is an empty state
		return nil
	}

	dataArr, found, err := getRequestLogRecordsForBlockBin(ctx.State(), blockIndex)
	ctx.RequireNoError(err)
	ctx.Requiref(found, "not found")

	ret := dict.New()
	arr := collections.NewArray16(ret, ParamRequestID)
	for _, d := range dataArr {
		rec, err := RequestReceiptFromBytes(d)
		ctx.RequireNoError(err)
		arr.MustPush(rec.Request.ID().Bytes())
	}
	return ret
}

// viewGetRequestReceiptsForBlock returns a list of receipts for a given block.
// params:
// ParamBlockIndex - index of the block (defaults to latest block)
func viewGetRequestReceiptsForBlock(ctx iscp.SandboxView) dict.Dict {
	blockIndex := getBlockIndexParams(ctx)

	if blockIndex == 0 {
		// block 0 is an empty state
		return nil
	}

	dataArr, found, err := getRequestLogRecordsForBlockBin(ctx.State(), blockIndex)
	ctx.RequireNoError(err)
	ctx.Requiref(found, "not found")

	ret := dict.New()
	arr := collections.NewArray16(ret, ParamRequestRecord)
	for _, d := range dataArr {
		arr.MustPush(d)
	}
	return ret
}

func viewControlAddresses(ctx iscp.SandboxView) dict.Dict {
	registry := collections.NewArray32ReadOnly(ctx.State(), prefixControlAddresses)
	l := registry.MustLen()
	ctx.Requiref(l > 0, "inconsistency: unknown control addresses")
	rec, err := ControlAddressesFromBytes(registry.MustGetAt(l - 1))
	ctx.RequireNoError(err)
	return dict.Dict{
		ParamStateControllerAddress: iscp.BytesFromAddress(rec.StateAddress),
		ParamGoverningAddress:       iscp.BytesFromAddress(rec.GoverningAddress),
		ParamBlockIndex:             codec.EncodeUint32(rec.SinceBlockIndex),
	}
}

// viewGetEventsForRequest returns a list of events for a given request.
// params:
// ParamRequestID - requestID
func viewGetEventsForRequest(ctx iscp.SandboxView) dict.Dict {
	requestID := ctx.ParamDecoder().MustGetRequestID(ParamRequestID)

	events, err := getRequestEventsInternal(ctx.State(), &requestID)
	ctx.RequireNoError(err)

	ret := dict.New()
	arr := collections.NewArray16(ret, ParamEvent)
	for _, event := range events {
		arr.MustPush([]byte(event))
	}
	return ret
}

// viewGetEventsForBlock returns a list of events for a given block.
// params:
// ParamBlockIndex - index of the block (defaults to latest block)
func viewGetEventsForBlock(ctx iscp.SandboxView) dict.Dict {
	blockIndex := getBlockIndexParams(ctx)

	if blockIndex == 0 {
		// block 0 is an empty state
		return nil
	}

	events, err := GetBlockEventsInternal(ctx.State(), blockIndex)
	ctx.RequireNoError(err)

	ret := dict.New()
	arr := collections.NewArray16(ret, ParamEvent)
	for _, event := range events {
		arr.MustPush([]byte(event))
	}
	return ret
}

// viewGetEventsForContract returns a list of events for a given smart contract.
// params:
// ParamContractHname - hname of the contract
// ParamFromBlock - defaults to 0
// ParamToBlock - defaults to latest block
func viewGetEventsForContract(ctx iscp.SandboxView) dict.Dict {
	contract := ctx.ParamDecoder().MustGetHname(ParamContractHname)
	fromBlock := ctx.ParamDecoder().MustGetUint32(ParamFromBlock, 0)
	toBlock := ctx.ParamDecoder().MustGetUint32(ParamToBlock, math.MaxUint32)
	events, err := getSmartContractEventsInternal(ctx.State(), contract, fromBlock, toBlock)
	ctx.RequireNoError(err)

	ret := dict.New()
	arr := collections.NewArray16(ret, ParamEvent)
	for _, event := range events {
		arr.MustPush([]byte(event))
	}
	return ret
}
