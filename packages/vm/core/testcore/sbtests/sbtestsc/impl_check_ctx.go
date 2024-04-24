package sbtestsc

import (
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/assert"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

func testCheckContextFromFullEP(ctx isc.Sandbox) dict.Dict {
	params := ctx.Params()

	ctx.Requiref(params.MustGetChainID(ParamChainID).Equals(ctx.ChainID()), "fail: chainID")
	ctx.Requiref(params.MustGetAgentID(ParamChainOwnerID).Equals(ctx.ChainOwnerID()), "fail: chainOwnerID")
	ctx.Requiref(params.MustGetAgentID(ParamCaller).Equals(ctx.Caller()), "fail: caller")
	myAgentID := isc.NewContractAgentID(ctx.ChainID(), ctx.Contract())
	ctx.Requiref(params.MustGetAgentID(ParamAgentID).Equals(myAgentID), "fail: agentID")
	return nil
}

func testCheckContextFromViewEP(ctx isc.SandboxView) dict.Dict {
	params := ctx.Params()
	a := assert.NewAssert(ctx.Log())

	a.Requiref(params.MustGetChainID(ParamChainID).Equals(ctx.ChainID()), "fail: chainID")
	a.Requiref(params.MustGetAgentID(ParamChainOwnerID).Equals(ctx.ChainOwnerID()), "fail: chainOwnerID")
	myAgentID := isc.NewContractAgentID(ctx.ChainID(), ctx.Contract())
	a.Requiref(params.MustGetAgentID(ParamAgentID).Equals(myAgentID), "fail: agentID")
	return nil
}

func passTypesFull(ctx isc.Sandbox) dict.Dict {
	ret := dict.New()
	params := ctx.Params()
	s, err := params.GetString("string")
	checkFull(ctx, err)
	if s != "string" {
		ctx.Log().Panicf("wrong string")
	}
	ret.Set("string", codec.String.Encode(s))

	i64, err := params.GetInt64("int64")
	checkFull(ctx, err)
	if i64 != 42 {
		ctx.Log().Panicf("wrong int64")
	}
	ret.Set("string", codec.Int64.Encode(42))

	i64_0, err := params.GetInt64("int64-0")
	checkFull(ctx, err)
	if i64_0 != 0 {
		ctx.Log().Panicf("wrong int64_0")
	}
	ret.Set("string", codec.Int64.Encode(0))

	hash, err := params.GetHashValue("Hash")
	checkFull(ctx, err)
	if hash != hashing.HashStrings("Hash") {
		ctx.Log().Panicf("wrong hash")
	}
	hname, err := params.GetHname("Hname")
	checkFull(ctx, err)
	if hname != isc.Hn("Hname") {
		ctx.Log().Panicf("wrong hname")
	}
	hname0, err := params.GetHname("Hname-0")
	checkFull(ctx, err)
	if hname0 != 0 {
		ctx.Log().Panicf("wrong Hname-0")
	}
	_, err = params.GetAgentID(ParamContractID)
	checkFull(ctx, err)

	_, err = params.GetChainID(ParamChainID)
	checkFull(ctx, err)

	_, err = params.GetAddress(ParamAddress)
	checkFull(ctx, err)

	_, err = params.GetAgentID(ParamAgentID)
	checkFull(ctx, err)
	return nil
}

func passTypesView(ctx isc.SandboxView) dict.Dict {
	params := ctx.Params()
	s, err := params.GetString("string")
	checkView(ctx, err)
	if s != "string" {
		ctx.Log().Panicf("wrong string")
	}
	i64, err := params.GetInt64("int64")
	checkView(ctx, err)
	if i64 != 42 {
		ctx.Log().Panicf("wrong int64")
	}
	i64_0, err := params.GetInt64("int64-0")
	checkView(ctx, err)
	if i64_0 != 0 {
		ctx.Log().Panicf("wrong int64_0")
	}
	hash, err := params.GetHashValue("Hash")
	checkView(ctx, err)
	if hash != hashing.HashStrings("Hash") {
		ctx.Log().Panicf("wrong hash")
	}
	hname, err := params.GetHname("Hname")
	checkView(ctx, err)
	if hname != isc.Hn("Hname") {
		ctx.Log().Panicf("wrong hname")
	}
	hname0, err := params.GetHname("Hname-0")
	checkView(ctx, err)
	if hname0 != 0 {
		ctx.Log().Panicf("wrong hname-0")
	}
	_, err = params.GetAgentID(ParamContractID)
	checkView(ctx, err)

	_, err = params.GetChainID(ParamChainID)
	checkView(ctx, err)

	_, err = params.GetAddress(ParamAddress)
	checkView(ctx, err)

	_, err = params.GetAgentID(ParamAgentID)
	checkView(ctx, err)
	return nil
}

func checkFull(ctx isc.Sandbox, err error) {
	if err != nil {
		ctx.Log().Panicf("Full sandbox: %v", err)
	}
}

func checkView(ctx isc.SandboxView, err error) {
	if err != nil {
		ctx.Log().Panicf("View sandbox: %v", err)
	}
}
