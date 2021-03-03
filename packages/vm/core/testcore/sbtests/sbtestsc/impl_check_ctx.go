package sbtestsc

import (
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/assert"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/kvdecoder"
)

func testCheckContextFromFullEP(ctx coretypes.Sandbox) (dict.Dict, error) {
	par := kvdecoder.New(ctx.Params(), ctx.Log())
	a := assert.NewAssert(ctx.Log())

	a.Require(par.MustGetChainID(ParamChainID) == ctx.ContractID().ChainID(), "fail: chainID")
	a.Require(par.MustGetAgentID(ParamChainOwnerID) == ctx.ChainOwnerID(), "fail: chainOwnerID")
	a.Require(par.MustGetAgentID(ParamCaller) == ctx.Caller(), "fail: caller")
	a.Require(par.MustGetContractID(ParamContractID) == ctx.ContractID(), "fail: contractID")
	a.Require(par.MustGetAgentID(ParamAgentID) == coretypes.NewAgentIDFromContractID(ctx.ContractID()), "fail: agentID")
	a.Require(par.MustGetAgentID(ParamContractCreator) == ctx.ContractCreator(), "fail: creator")
	return nil, nil
}

func testCheckContextFromViewEP(ctx coretypes.SandboxView) (dict.Dict, error) {
	par := kvdecoder.New(ctx.Params(), ctx.Log())
	a := assert.NewAssert(ctx.Log())

	a.Require(par.MustGetChainID(ParamChainID) == ctx.ContractID().ChainID(), "fail: chainID")
	a.Require(par.MustGetAgentID(ParamChainOwnerID) == ctx.ChainOwnerID(), "fail: chainOwnerID")
	a.Require(par.MustGetContractID(ParamContractID) == ctx.ContractID(), "fail: contractID")
	a.Require(par.MustGetAgentID(ParamAgentID) == coretypes.NewAgentIDFromContractID(ctx.ContractID()), "fail: agentID")
	a.Require(par.MustGetAgentID(ParamContractCreator) == ctx.ContractCreator(), "fail: creator")
	return nil, nil
}

func passTypesFull(ctx coretypes.Sandbox) (dict.Dict, error) {
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
	if hname != coretypes.Hn("Hname") {
		ctx.Log().Panicf("wrong hname")
	}
	hname_0, exists, err := codec.DecodeHname(ctx.Params().MustGet("Hname-0"))
	checkFull(ctx, exists, err)
	if hname_0 != 0 {
		ctx.Log().Panicf("wrong Hname-0")
	}
	_, exists, err = codec.DecodeHname(ctx.Params().MustGet("ContractID"))
	checkFull(ctx, exists, err)

	_, exists, err = codec.DecodeHname(ctx.Params().MustGet("ChainID"))
	checkFull(ctx, exists, err)

	_, exists, err = codec.DecodeHname(ctx.Params().MustGet("Address"))
	checkFull(ctx, exists, err)

	_, exists, err = codec.DecodeHname(ctx.Params().MustGet("AgentID"))
	checkFull(ctx, exists, err)
	return nil, nil
}

func passTypesView(ctx coretypes.SandboxView) (dict.Dict, error) {
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
	if hname != coretypes.Hn("Hname") {
		ctx.Log().Panicf("wrong hname")
	}
	hname_0, exists, err := codec.DecodeHname(ctx.Params().MustGet("Hname-0"))
	checkView(ctx, exists, err)
	if hname_0 != 0 {
		ctx.Log().Panicf("wrong hname-0")
	}
	_, exists, err = codec.DecodeHname(ctx.Params().MustGet("ContractID"))
	checkView(ctx, exists, err)

	_, exists, err = codec.DecodeHname(ctx.Params().MustGet("ChainID"))
	checkView(ctx, exists, err)

	_, exists, err = codec.DecodeHname(ctx.Params().MustGet("Address"))
	checkView(ctx, exists, err)

	_, exists, err = codec.DecodeHname(ctx.Params().MustGet("AgentID"))
	checkView(ctx, exists, err)
	return nil, nil
}

func checkFull(ctx coretypes.Sandbox, exists bool, err error) {
	if err != nil {
		ctx.Log().Panicf("Full sandbox: %v", err)
	}
	if !exists {
		ctx.Log().Panicf("Full sandbox: param not found")
	}
}

func checkView(ctx coretypes.SandboxView, exists bool, err error) {
	if err != nil {
		ctx.Log().Panicf("View sandbox: %v", err)
	}
	if !exists {
		ctx.Log().Panicf("View sandbox: param not found")
	}
}
