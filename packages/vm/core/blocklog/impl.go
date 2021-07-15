package blocklog

import (
	"time"

	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/assert"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/kvdecoder"
	"golang.org/x/xerrors"
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
	ctx.Log().Debugf("blocklog.initialize.success hname = %s", Interface.Hname().String())
	return nil, nil
}

func viewGetBlockInfo(ctx iscp.SandboxView) (dict.Dict, error) {
	params := kvdecoder.New(ctx.Params())
	blockIndex := params.MustGetUint32(ParamBlockIndex)
	data, found, err := getBlockInfoDataIntern(ctx.State(), blockIndex)
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
	seen, err := isRequestProcessedIntern(ctx.State(), &requestID)
	a.RequireNoError(err)
	ret := dict.New()
	if seen {
		ret.Set(ParamRequestProcessed, codec.EncodeString("+"))
	}
	return ret, nil
}

func viewGetRequestLogRecord(ctx iscp.SandboxView) (dict.Dict, error) {
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

	dataArr, found, err := getRequestLogRecordsForBlockBin(ctx.State(), blockIndex)
	a.RequireNoError(err)
	a.Require(found, "not found")

	ret := dict.New()
	arr := collections.NewArray16(ret, ParamRequestID)
	for _, d := range dataArr {
		rec, err := RequestLogRecordFromBytes(d)
		a.RequireNoError(err)
		_ = arr.Push(rec.RequestID.Bytes())
	}
	return ret, nil
}

func viewGetRequestLogRecordsForBlock(ctx iscp.SandboxView) (dict.Dict, error) {
	params := kvdecoder.New(ctx.Params())
	a := assert.NewAssert(ctx.Log())
	blockIndex := params.MustGetUint32(ParamBlockIndex)

	dataArr, found, err := getRequestLogRecordsForBlockBin(ctx.State(), blockIndex)
	a.RequireNoError(err)
	a.Require(found, "not found")

	ret := dict.New()
	arr := collections.NewArray16(ret, ParamRequestRecord)
	for _, d := range dataArr {
		_ = arr.Push(d)
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
