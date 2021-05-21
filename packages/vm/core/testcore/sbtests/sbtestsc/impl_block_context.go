package sbtestsc

import (
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/assert"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/kvdecoder"
)

type blockCtx1 struct {
	numcalls int
}

func getBlockContext1(ctx coretypes.Sandbox) *blockCtx1 {
	construct := func(ctx coretypes.Sandbox) interface{} {
		return &blockCtx1{1}
	}
	log := ctx.Log()
	onClose := func(obj interface{}) {
		log.Infof("closing block context with numcalls = %d", obj.(*blockCtx1).numcalls)
	}
	bctxi := ctx.BlockContext(construct, onClose)
	bctx, ok := bctxi.(*blockCtx1)
	assert.NewAssert(ctx.Log()).Require(ok, "unexpected block context type")
	return bctx
}

func testBlockContext1(ctx coretypes.Sandbox) (dict.Dict, error) {
	bctx := getBlockContext1(ctx)
	bctx.numcalls++

	return nil, nil
}

type blockCtx2 struct {
	state kv.KVStore
	log   coretypes.LogInterface
}

func construct2(ctx coretypes.Sandbox) interface{} {
	return &blockCtx2{
		state: ctx.State(),
		log:   ctx.Log(),
	}
}

func onClose2(obj interface{}) {
	bctx := obj.(*blockCtx2)
	bctx.state.Set("atTheEndKey", []byte("atTheEndValue"))
	bctx.log.Infof("closing block context...")
}

func testBlockContext2(ctx coretypes.Sandbox) (dict.Dict, error) {
	ctx.BlockContext(construct2, onClose2) // just creating context, doing nothing, checking side effect
	return nil, nil
}

func getStringValue(ctx coretypes.SandboxView) (dict.Dict, error) {
	ctx.Log().Infof(FuncGetStringValue)
	deco := kvdecoder.New(ctx.Params(), ctx.Log())
	varName := deco.MustGetString(ParamVarName)
	value := string(ctx.State().MustGet(kv.Key(varName)))
	ret := dict.New()
	ret.Set(kv.Key(varName), codec.EncodeString(value))
	return ret, nil
}
