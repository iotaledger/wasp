package sbtestsc

import (
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/assert"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/kvdecoder"
)

func testCheckContextFromFullEP(ctx iscp.Sandbox) (dict.Dict, error) {
	par := kvdecoder.New(ctx.Params(), ctx.Log())
	a := assert.NewAssert(ctx.Log())

	a.Require(par.MustGetChainID(ParamChainID).Equals(ctx.ChainID()), "fail: chainID")
	a.Require(par.MustGetAgentID(ParamChainOwnerID).Equals(ctx.ChainOwnerID()), "fail: chainOwnerID")
	a.Require(par.MustGetAgentID(ParamCaller).Equals(ctx.Caller()), "fail: caller")
	myAgentID := iscp.NewAgentID(ctx.ChainID().AsAddress(), ctx.Contract())
	a.Require(par.MustGetAgentID(ParamAgentID).Equals(myAgentID), "fail: agentID")
	a.Require(par.MustGetAgentID(ParamContractCreator).Equals(ctx.ContractCreator()), "fail: creator")
	return nil, nil
}

func testCheckContextFromViewEP(ctx iscp.SandboxView) (dict.Dict, error) {
	par := kvdecoder.New(ctx.Params(), ctx.Log())
	a := assert.NewAssert(ctx.Log())

	a.Require(par.MustGetChainID(ParamChainID).Equals(ctx.ChainID()), "fail: chainID")
	a.Require(par.MustGetAgentID(ParamChainOwnerID).Equals(ctx.ChainOwnerID()), "fail: chainOwnerID")
	myAgentID := iscp.NewAgentID(ctx.ChainID().AsAddress(), ctx.Contract())
	a.Require(par.MustGetAgentID(ParamAgentID).Equals(myAgentID), "fail: agentID")
	a.Require(par.MustGetAgentID(ParamContractCreator).Equals(ctx.ContractCreator()), "fail: creator")
	return nil, nil
}

func passTypesFull(ctx iscp.Sandbox) (dict.Dict, error) {
	ret := dict.New()
	s, exists, err := codec.DecodeString(ctx.Params().MustGet("string"))
	checkFull(ctx, exists, err)
	if s != "string" {
		ctx.Log().Panicf("wrong string")
	}
	ret.Set("string", codec.EncodeString(s))

	i64, exists, err := codec.DecodeInt64(ctx.Params().MustGet("int64"))
	checkFull(ctx, exists, err)
	if i64 != 42 {
		ctx.Log().Panicf("wrong int64")
	}
	ret.Set("string", codec.EncodeInt64(42))

	i64_0, exists, err := codec.DecodeInt64(ctx.Params().MustGet("int64-0"))
	checkFull(ctx, exists, err)
	if i64_0 != 0 {
		ctx.Log().Panicf("wrong int64_0")
	}
	ret.Set("string", codec.EncodeInt64(0))

	hash, exists, err := codec.DecodeHashValue(ctx.Params().MustGet("Hash"))
	checkFull(ctx, exists, err)
	if hash != hashing.HashStrings("Hash") {
		ctx.Log().Panicf("wrong hash")
	}
	hname, exists, err := codec.DecodeHname(ctx.Params().MustGet("Hname"))
	checkFull(ctx, exists, err)
	if hname != iscp.Hn("Hname") {
		ctx.Log().Panicf("wrong hname")
	}
	hname0, exists, err := codec.DecodeHname(ctx.Params().MustGet("Hname-0"))
	checkFull(ctx, exists, err)
	if hname0 != 0 {
		ctx.Log().Panicf("wrong Hname-0")
	}
	_, exists, err = codec.DecodeHname(ctx.Params().MustGet(ParamContractID))
	checkFull(ctx, exists, err)

	_, exists, err = codec.DecodeHname(ctx.Params().MustGet(ParamChainID))
	checkFull(ctx, exists, err)

	_, exists, err = codec.DecodeHname(ctx.Params().MustGet(ParamAddress))
	checkFull(ctx, exists, err)

	_, exists, err = codec.DecodeHname(ctx.Params().MustGet(ParamAgentID))
	checkFull(ctx, exists, err)
	return nil, nil
}

func passTypesView(ctx iscp.SandboxView) (dict.Dict, error) {
	s, exists, err := codec.DecodeString(ctx.Params().MustGet("string"))
	checkView(ctx, exists, err)
	if s != "string" {
		ctx.Log().Panicf("wrong string")
	}
	i64, exists, err := codec.DecodeInt64(ctx.Params().MustGet("int64"))
	checkView(ctx, exists, err)
	if i64 != 42 {
		ctx.Log().Panicf("wrong int64")
	}
	i64_0, exists, err := codec.DecodeInt64(ctx.Params().MustGet("int64-0"))
	checkView(ctx, exists, err)
	if i64_0 != 0 {
		ctx.Log().Panicf("wrong int64_0")
	}
	hash, exists, err := codec.DecodeHashValue(ctx.Params().MustGet("Hash"))
	checkView(ctx, exists, err)
	if hash != hashing.HashStrings("Hash") {
		ctx.Log().Panicf("wrong hash")
	}
	hname, exists, err := codec.DecodeHname(ctx.Params().MustGet("Hname"))
	checkView(ctx, exists, err)
	if hname != iscp.Hn("Hname") {
		ctx.Log().Panicf("wrong hname")
	}
	hname0, exists, err := codec.DecodeHname(ctx.Params().MustGet("Hname-0"))
	checkView(ctx, exists, err)
	if hname0 != 0 {
		ctx.Log().Panicf("wrong hname-0")
	}
	_, exists, err = codec.DecodeHname(ctx.Params().MustGet(ParamContractID))
	checkView(ctx, exists, err)

	_, exists, err = codec.DecodeHname(ctx.Params().MustGet(ParamChainID))
	checkView(ctx, exists, err)

	_, exists, err = codec.DecodeHname(ctx.Params().MustGet(ParamAddress))
	checkView(ctx, exists, err)

	_, exists, err = codec.DecodeHname(ctx.Params().MustGet(ParamAgentID))
	checkView(ctx, exists, err)
	return nil, nil
}

func checkFull(ctx iscp.Sandbox, exists bool, err error) {
	if err != nil {
		ctx.Log().Panicf("Full sandbox: %v", err)
	}
	if !exists {
		ctx.Log().Panicf("Full sandbox: param not found")
	}
}

func checkView(ctx iscp.SandboxView, exists bool, err error) {
	if err != nil {
		ctx.Log().Panicf("View sandbox: %v", err)
	}
	if !exists {
		ctx.Log().Panicf("View sandbox: param not found")
	}
}
