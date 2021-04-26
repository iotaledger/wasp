package blocklog

import (
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/assert"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/kvdecoder"
	"github.com/iotaledger/wasp/packages/util"
	"golang.org/x/xerrors"
	"time"
)

func initialize(ctx coretypes.Sandbox) (dict.Dict, error) {
	blockIndex := SaveNextBlockInfo(ctx.State(), &BlockInfo{
		Timestamp:             time.Unix(0, ctx.GetTimestamp()),
		TotalRequests:         1,
		NumSuccessfulRequests: 1,
		NumOffLedgerRequests:  0,
	})
	assert.NewAssert(ctx.Log()).Require(blockIndex == 0, "blocklog.initialize.fail: unexpected block index")
	ctx.Log().Debugf("blocklog.initialize.success hname = %s", Interface.Hname().String())
	return nil, nil
}

func getBlockInfo(ctx coretypes.SandboxView) (dict.Dict, error) {
	params := kvdecoder.New(ctx.Params())
	blockIndex64 := params.MustGetUint64(ParamBlockIndex)
	if blockIndex64 > uint64(util.MaxUint32) {
		return nil, xerrors.New("blocklog::getBlockInfo: incorrect block index")
	}
	blockIndex := uint32(blockIndex64)
	data, found := getBlockInfoDataIntern(ctx.State(), blockIndex)
	if !found {
		return nil, xerrors.New("not found")
	}
	ret := dict.New()
	ret.Set(ParamBlockInfo, data)
	return ret, nil
}

func getLatestBlockInfo(ctx coretypes.SandboxView) (dict.Dict, error) {
	registry := collections.NewArray32ReadOnly(ctx.State(), StateVarBlockRegistry)
	l := registry.MustLen()
	if l == 0 {
		return nil, xerrors.New("blocklog::getLatestBlockInfo: empty log")
	}
	data := registry.MustGetAt(l - 1)
	ret := dict.New()
	ret.Set(ParamBlockIndex, codec.EncodeUint64(uint64(l-1)))
	ret.Set(ParamBlockInfo, data)
	return ret, nil
}

func isRequestProcessed(ctx coretypes.SandboxView) (dict.Dict, error) {
	params := kvdecoder.New(ctx.Params())
	requestID := params.MustGetRequestID(ParamRequestID)
	a := assert.NewAssert(ctx.Log())
	notSeen, err := RequestNotSeen(ctx.State(), &requestID)
	a.RequireNoError(err)
	ret := dict.New()
	if !notSeen {
		ret.Set(ParamRequestProcessed, codec.EncodeString("+"))
	}
	return ret, nil
}

func getRequestLogRecord(ctx coretypes.SandboxView) (dict.Dict, error) {
	params := kvdecoder.New(ctx.Params())
	requestID := params.MustGetRequestID(ParamRequestID)
	recBin, blockIndex, requestIndex, found := getRequestRecordDataByRequestID(ctx, requestID)
	ret := dict.New()
	if !found {
		return ret, nil
	}
	ret.Set(ParamRequestRecord, recBin)
	ret.Set(ParamBlockIndex, codec.EncodeUint64(uint64(blockIndex)))
	ret.Set(ParamRequestIndex, codec.EncodeUint64(uint64(requestIndex)))
	return ret, nil
}

func getRequestLogRecordsForBlock(ctx coretypes.SandboxView) (dict.Dict, error) {
	params := kvdecoder.New(ctx.Params())
	a := assert.NewAssert(ctx.Log())
	blockIndex64 := params.MustGetUint64(ParamBlockIndex)
	a.Require(int(blockIndex64) <= util.MaxUint32, "wrong block index parameter")
	blockIndex := uint32(blockIndex64)

	blockInfo, found := getBlockInfoIntern(ctx, blockIndex)
	a.Require(found, "not found")

	ret := dict.New()
	arr := collections.NewArray16(ret, ParamRequestRecord)
	for reqIdx := uint16(0); reqIdx < blockInfo.TotalRequests; reqIdx++ {
		data, found := getRequestRecordDataByRef(ctx.State(), blockIndex, reqIdx)
		a.Require(found, "inconsistency: request record not found")
		_ = arr.Push(data)
	}
	return ret, nil
}
