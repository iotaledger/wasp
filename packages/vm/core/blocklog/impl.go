package blocklog

import (
	"math"
	"time"

	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/assert"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/kvdecoder"
	"golang.org/x/xerrors"
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

func initialize(ctx iscp.Sandbox) (dict.Dict, error) {
	blockIndex := SaveNextBlockInfo(ctx.State(), &BlockInfo{
		Timestamp:             time.Unix(0, ctx.GetTimestamp()),
		TotalRequests:         1,
		NumSuccessfulRequests: 1,
		NumOffLedgerRequests:  0,
	})
	a := assert.NewAssert(ctx.Log())
	a.Require(blockIndex == 0, "blocklog.initialize.fail: unexpected block index")
	ctx.Log().Debugf("blocklog.initialize.success hname = %s", Contract.Hname().String())
	return nil, nil
}

func viewGetBlockInfo(ctx iscp.SandboxView) (dict.Dict, error) {
	params := kvdecoder.New(ctx.Params())
	blockIndex := params.MustGetUint32(ParamBlockIndex)
	data, found, err := getBlockInfoDataInternal(ctx.State(), blockIndex)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, xerrors.New("not found")
	}
	ret := dict.New()
	ret.Set(ParamBlockInfo, data)
	return ret, nil
}

func viewGetLatestBlockInfo(ctx iscp.SandboxView) (dict.Dict, error) {
	registry := collections.NewArray32ReadOnly(ctx.State(), StateVarBlockRegistry)
	regLen := registry.MustLen()
	if regLen == 0 {
		return nil, xerrors.New("blocklog::viewGetLatestBlockInfo: empty log")
	}
	data := registry.MustGetAt(regLen - 1)
	ret := dict.New()
	ret.Set(ParamBlockIndex, codec.EncodeUint32(regLen-1))
	ret.Set(ParamBlockInfo, data)
	return ret, nil
}

func viewIsRequestProcessed(ctx iscp.SandboxView) (dict.Dict, error) {
	params := kvdecoder.New(ctx.Params())
	requestID := params.MustGetRequestID(ParamRequestID)
	a := assert.NewAssert(ctx.Log())
	seen, err := isRequestProcessedInternal(ctx.State(), &requestID)
	a.RequireNoError(err)
	ret := dict.New()
	if seen {
		ret.Set(ParamRequestProcessed, codec.EncodeString("+"))
	}
	return ret, nil
}

func viewGetRequestReceipt(ctx iscp.SandboxView) (dict.Dict, error) {
	params := kvdecoder.New(ctx.Params())
	requestID := params.MustGetRequestID(ParamRequestID)
	recBin, blockIndex, requestIndex, found := getRequestRecordDataByRequestID(ctx, requestID)
	ret := dict.New()
	if !found {
		return ret, nil
	}
	ret.Set(ParamRequestRecord, recBin)
	ret.Set(ParamBlockIndex, codec.EncodeUint32(blockIndex))
	ret.Set(ParamRequestIndex, codec.EncodeUint16(requestIndex))
	return ret, nil
}

func viewGetRequestIDsForBlock(ctx iscp.SandboxView) (dict.Dict, error) {
	params := kvdecoder.New(ctx.Params())
	a := assert.NewAssert(ctx.Log())
	blockIndex := params.MustGetUint32(ParamBlockIndex)

	if blockIndex == 0 {
		// block 0 is an empty state
		return dict.Dict{}, nil
	}

	dataArr, found, err := getRequestLogRecordsForBlockBin(ctx.State(), blockIndex)
	a.RequireNoError(err)
	a.Require(found, "not found")

	ret := dict.New()
	arr := collections.NewArray16(ret, ParamRequestID)
	for _, d := range dataArr {
		rec, err := RequestReceiptFromBytes(d)
		a.RequireNoError(err)
		arr.MustPush(rec.Request.ID().Bytes())
	}
	return ret, nil
}

func viewGetRequestReceiptsForBlock(ctx iscp.SandboxView) (dict.Dict, error) {
	params := kvdecoder.New(ctx.Params())
	a := assert.NewAssert(ctx.Log())
	blockIndex := params.MustGetUint32(ParamBlockIndex)

	if blockIndex == 0 {
		// block 0 is an empty state
		return dict.Dict{}, nil
	}

	dataArr, found, err := getRequestLogRecordsForBlockBin(ctx.State(), blockIndex)
	a.RequireNoError(err)
	a.Require(found, "not found")

	ret := dict.New()
	arr := collections.NewArray16(ret, ParamRequestRecord)
	for _, d := range dataArr {
		arr.MustPush(d)
	}
	return ret, nil
}

func viewControlAddresses(ctx iscp.SandboxView) (dict.Dict, error) {
	a := assert.NewAssert(ctx.Log())
	registry := collections.NewArray32ReadOnly(ctx.State(), StateVarControlAddresses)
	l := registry.MustLen()
	a.Require(l > 0, "inconsistency: unknown control addresses")
	rec, err := ControlAddressesFromBytes(registry.MustGetAt(l - 1))
	a.RequireNoError(err)
	ret := dict.New()
	ret.Set(ParamStateControllerAddress, codec.EncodeAddress(rec.StateAddress))
	ret.Set(ParamGoverningAddress, codec.EncodeAddress(rec.GoverningAddress))
	ret.Set(ParamBlockIndex, codec.EncodeUint32(rec.SinceBlockIndex))
	return ret, nil
}

// viewGetEventsForRequest returns a list of events for a given request.
// params:
// ParamRequestID - requestID
func viewGetEventsForRequest(ctx iscp.SandboxView) (dict.Dict, error) {
	params := kvdecoder.New(ctx.Params())
	requestID := params.MustGetRequestID(ParamRequestID)

	events, err := getRequestEventsInternal(ctx.State(), &requestID)
	if err != nil {
		return nil, err
	}

	ret := dict.New()
	arr := collections.NewArray16(ret, ParamEvent)
	for _, event := range events {
		arr.MustPush([]byte(event))
	}
	return ret, nil
}

// viewGetEventsForBlock returns a list of events for a given block.
// params:
// ParamBlockIndex - index of the block (defaults to latest block)
func viewGetEventsForBlock(ctx iscp.SandboxView) (dict.Dict, error) {
	params := kvdecoder.New(ctx.Params())

	var blockIndex uint32
	if ctx.Params().MustHas(ParamBlockIndex) {
		blockIndex = params.MustGetUint32(ParamBlockIndex)
	} else {
		registry := collections.NewArray32ReadOnly(ctx.State(), StateVarBlockRegistry)
		blockIndex = registry.MustLen() - 1
	}

	if blockIndex == 0 {
		// block 0 is an empty state
		return dict.Dict{}, nil
	}

	events, err := GetBlockEventsInternal(ctx.State(), blockIndex)
	if err != nil {
		return nil, err
	}

	ret := dict.New()
	arr := collections.NewArray16(ret, ParamEvent)
	for _, event := range events {
		arr.MustPush([]byte(event))
	}
	return ret, nil
}

// viewGetEventsForContract returns a list of events for a given smart contract.
// params:
// ParamContractHname - hname of the contract
// ParamFromBlock - defaults to 0
// ParamToBlock - defaults to latest block
func viewGetEventsForContract(ctx iscp.SandboxView) (dict.Dict, error) {
	params := kvdecoder.New(ctx.Params())
	contract := params.MustGetHname(ParamContractHname)
	fromBlock, err := params.GetUint32(ParamFromBlock, 0)
	if err != nil {
		return nil, err
	}
	toBlock, _ := params.GetUint32(ParamToBlock, math.MaxUint32)
	if err != nil {
		return nil, err
	}
	events, err := getSmartContractEventsInternal(ctx.State(), contract, fromBlock, toBlock)
	if err != nil {
		return nil, err
	}

	ret := dict.New()
	arr := collections.NewArray16(ret, ParamEvent)
	for _, event := range events {
		arr.MustPush([]byte(event))
	}
	return ret, nil
}
