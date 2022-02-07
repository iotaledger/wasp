package sbtestsc

import (
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/assert"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/kvdecoder"
)

type blockCtx1 struct {
	numcalls int
}

func getBlockContext1(ctx iscp.Sandbox) *blockCtx1 {
	construct := func(ctx iscp.Sandbox) interface{} {
		return &blockCtx1{1}
	}
	log := ctx.Log()
	onClose := func(obj interface{}) {
		log.Infof("closing block context with numcalls = %d", obj.(*blockCtx1).numcalls)
	}
	bctxi := ctx.Privileged().BlockContext(construct, onClose)
	bctx, ok := bctxi.(*blockCtx1)
	assert.NewAssert(ctx.Log()).Requiref(ok, "unexpected block context type")
	return bctx
}

func testBlockContext1(ctx iscp.Sandbox) dict.Dict {
	bctx := getBlockContext1(ctx)
	bctx.numcalls++

	return nil
}

type blockCtx2 struct {
	state kv.KVStore
	log   iscp.LogInterface
}

func construct2(ctx iscp.Sandbox) interface{} {
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

func testBlockContext2(ctx iscp.Sandbox) dict.Dict {
	ctx.Privileged().BlockContext(construct2, onClose2) // just creating context, doing nothing, checking side effect
	return nil
}

func getStringValue(ctx iscp.SandboxView) dict.Dict {
	ctx.Log().Infof(FuncGetStringValue.Name)
	deco := kvdecoder.New(ctx.Params(), ctx.Log())
	varName := deco.MustGetString(ParamVarName)
	value := string(ctx.State().MustGet(kv.Key(varName)))
	ret := dict.New()
	ret.Set(kv.Key(varName), codec.EncodeString(value))
	return ret
}
