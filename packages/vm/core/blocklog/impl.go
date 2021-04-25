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
	blockIndex64, err := params.GetUint64(ParamBlockIndex)
	if err != nil {
		return nil, err
	}
	if blockIndex64 > uint64(util.MaxUint32) {
		return nil, xerrors.New("blocklog::getBlockInfo: incorrect block index")
	}
	blockIndex := uint32(blockIndex64)
	data, err := collections.NewArray32ReadOnly(ctx.State(), BlockRegistry).GetAt(blockIndex)
	if err != nil {
		return nil, xerrors.Errorf("blocklog::getBlockInfo at index #%d: %w", blockIndex, err)
	}
	if data == nil {
		return nil, xerrors.Errorf("blocklog::getBlockInfo at index #%d: not found", blockIndex)
	}
	ret := dict.New()
	ret.Set(ParamBlockInfo, data)
	return ret, nil
}

func getLatestBlockInfo(ctx coretypes.SandboxView) (dict.Dict, error) {
	registry := collections.NewArray32ReadOnly(ctx.State(), BlockRegistry)
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
