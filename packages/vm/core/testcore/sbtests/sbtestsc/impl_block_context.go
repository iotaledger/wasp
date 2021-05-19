package sbtestsc

import (
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/assert"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/kvdecoder"
)

type blockCtx struct {
	numcalls int
}

func testBlockContext1(ctx coretypes.Sandbox) (dict.Dict, error) {
	a := assert.NewAssert(ctx.Log())
	bctxi, exists := ctx.GetBlockContext()
	if !exists {
		bctx := &blockCtx{1}
		ctx.CreateBlockContext(bctx, func() {
			ctx.Log().Infof("closing block context with numcalls = %d", bctx.numcalls)
		})
		return nil, nil
	}
	bctx, exists := bctxi.(*blockCtx)
	a.Require(exists, "unexpected block context type")
	bctx.numcalls++

	return nil, nil
}

func testBlockContext2(ctx coretypes.Sandbox) (dict.Dict, error) {
	a := assert.NewAssert(ctx.Log())
	bctxi, exists := ctx.GetBlockContext()
	if !exists {
		bctx := ctx.State()
		ctx.CreateBlockContext(bctx, func() {
			bctx.Set("atTheEndKey", []byte("atTheEndValue"))
			ctx.Log().Infof("closing block context")
		})
		return nil, nil
	}
	bctx, exists := bctxi.(*blockCtx)
	a.Require(exists, "unexpected block context type")
	bctx.numcalls++

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
