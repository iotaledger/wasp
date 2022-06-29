package sbtestsc

import (
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/assert"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/kvdecoder"
)

func testCheckContextFromFullEP(ctx iscp.Sandbox) dict.Dict {
	par := kvdecoder.New(ctx.Params(), ctx.Log())

	ctx.Requiref(par.MustGetChainID(ParamChainID).Equals(ctx.ChainID()), "fail: chainID")
	ctx.Requiref(par.MustGetAgentID(ParamChainOwnerID).Equals(ctx.ChainOwnerID()), "fail: chainOwnerID")
	ctx.Requiref(par.MustGetAgentID(ParamCaller).Equals(ctx.Caller()), "fail: caller")
	myAgentID := iscp.NewContractAgentID(ctx.ChainID(), ctx.Contract())
	ctx.Requiref(par.MustGetAgentID(ParamAgentID).Equals(myAgentID), "fail: agentID")
	return nil
}

func testCheckContextFromViewEP(ctx iscp.SandboxView) dict.Dict {
	par := kvdecoder.New(ctx.Params(), ctx.Log())
	a := assert.NewAssert(ctx.Log())

	a.Requiref(par.MustGetChainID(ParamChainID).Equals(ctx.ChainID()), "fail: chainID")
	a.Requiref(par.MustGetAgentID(ParamChainOwnerID).Equals(ctx.ChainOwnerID()), "fail: chainOwnerID")
	myAgentID := iscp.NewContractAgentID(ctx.ChainID(), ctx.Contract())
	a.Requiref(par.MustGetAgentID(ParamAgentID).Equals(myAgentID), "fail: agentID")
	return nil
}

func passTypesFull(ctx iscp.Sandbox) dict.Dict {
	ret := dict.New()
	s, err := codec.DecodeString(ctx.Params().MustGet("string"))
	checkFull(ctx, err)
	if s != "string" {
		ctx.Log().Panicf("wrong string")
	}
	ret.Set("string", codec.EncodeString(s))

	i64, err := codec.DecodeInt64(ctx.Params().MustGet("int64"))
	checkFull(ctx, err)
	if i64 != 42 {
		ctx.Log().Panicf("wrong int64")
	}
	ret.Set("string", codec.EncodeInt64(42))

	i64_0, err := codec.DecodeInt64(ctx.Params().MustGet("int64-0"))
	checkFull(ctx, err)
	if i64_0 != 0 {
		ctx.Log().Panicf("wrong int64_0")
	}
	ret.Set("string", codec.EncodeInt64(0))

	hash, err := codec.DecodeHashValue(ctx.Params().MustGet("Hash"))
	checkFull(ctx, err)
	if hash != hashing.HashStrings("Hash") {
		ctx.Log().Panicf("wrong hash")
	}
	hname, err := codec.DecodeHname(ctx.Params().MustGet("Hname"))
	checkFull(ctx, err)
	if hname != iscp.Hn("Hname") {
		ctx.Log().Panicf("wrong hname")
	}
	hname0, err := codec.DecodeHname(ctx.Params().MustGet("Hname-0"))
	checkFull(ctx, err)
	if hname0 != 0 {
		ctx.Log().Panicf("wrong Hname-0")
	}
	_, err = codec.DecodeHname(ctx.Params().MustGet(ParamContractID))
	checkFull(ctx, err)

	_, err = codec.DecodeHname(ctx.Params().MustGet(ParamChainID))
	checkFull(ctx, err)

	_, err = codec.DecodeHname(ctx.Params().MustGet(ParamAddress))
	checkFull(ctx, err)

	_, err = codec.DecodeHname(ctx.Params().MustGet(ParamAgentID))
	checkFull(ctx, err)
	return nil
}

func passTypesView(ctx iscp.SandboxView) dict.Dict {
	s, err := codec.DecodeString(ctx.Params().MustGet("string"))
	checkView(ctx, err)
	if s != "string" {
		ctx.Log().Panicf("wrong string")
	}
	i64, err := codec.DecodeInt64(ctx.Params().MustGet("int64"))
	checkView(ctx, err)
	if i64 != 42 {
		ctx.Log().Panicf("wrong int64")
	}
	i64_0, err := codec.DecodeInt64(ctx.Params().MustGet("int64-0"))
	checkView(ctx, err)
	if i64_0 != 0 {
		ctx.Log().Panicf("wrong int64_0")
	}
	hash, err := codec.DecodeHashValue(ctx.Params().MustGet("Hash"))
	checkView(ctx, err)
	if hash != hashing.HashStrings("Hash") {
		ctx.Log().Panicf("wrong hash")
	}
	hname, err := codec.DecodeHname(ctx.Params().MustGet("Hname"))
	checkView(ctx, err)
	if hname != iscp.Hn("Hname") {
		ctx.Log().Panicf("wrong hname")
	}
	hname0, err := codec.DecodeHname(ctx.Params().MustGet("Hname-0"))
	checkView(ctx, err)
	if hname0 != 0 {
		ctx.Log().Panicf("wrong hname-0")
	}
	_, err = codec.DecodeHname(ctx.Params().MustGet(ParamContractID))
	checkView(ctx, err)

	_, err = codec.DecodeHname(ctx.Params().MustGet(ParamChainID))
	checkView(ctx, err)

	_, err = codec.DecodeHname(ctx.Params().MustGet(ParamAddress))
	checkView(ctx, err)

	_, err = codec.DecodeHname(ctx.Params().MustGet(ParamAgentID))
	checkView(ctx, err)
	return nil
}

func checkFull(ctx iscp.Sandbox, err error) {
	if err != nil {
		ctx.Log().Panicf("Full sandbox: %v", err)
	}
}

func checkView(ctx iscp.SandboxView, err error) {
	if err != nil {
		ctx.Log().Panicf("View sandbox: %v", err)
	}
}
