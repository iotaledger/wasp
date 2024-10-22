package sbtestsc

import (
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/assert"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

func testCheckContextFromFullEP(ctx isc.Sandbox, chainID isc.ChainID, chainOwnerID isc.AgentID, caller isc.AgentID, agentID isc.AgentID) {
	ctx.Requiref(chainID.Equals(ctx.ChainID()), "fail: chainID")
	ctx.Requiref(chainOwnerID.Equals(ctx.ChainOwnerID()), "fail: chainOwnerID")
	ctx.Requiref(caller.Equals(ctx.Caller()), "fail: caller")
	myAgentID := isc.NewContractAgentID(ctx.ChainID(), ctx.Contract())
	ctx.Requiref(agentID.Equals(myAgentID), "fail: agentID")
}

func testCheckContextFromViewEP(ctx isc.SandboxView, chainID isc.ChainID, chainOwnerID isc.AgentID, agentID isc.AgentID) {
	a := assert.NewAssert(ctx.Log())

	a.Requiref(chainID.Equals(ctx.ChainID()), "fail: chainID")
	a.Requiref(chainOwnerID.Equals(ctx.ChainOwnerID()), "fail: chainOwnerID")
	myAgentID := isc.NewContractAgentID(ctx.ChainID(), ctx.Contract())
	a.Requiref(agentID.Equals(myAgentID), "fail: agentID")
}

func passTypesFull(ctx isc.Sandbox) isc.CallArguments {
	ret := dict.New()
	params := ctx.Params()
	s, err := isc.ArgAt[string](params, 0)
	checkFull(ctx, err)
	if s != "string" {
		ctx.Log().Panicf("wrong string")
	}
	ret.Set("string", codec.Encode[string](s))

	i64, err := isc.ArgAt[int64](params, 1)
	checkFull(ctx, err)
	if i64 != 42 {
		ctx.Log().Panicf("wrong int64")
	}
	ret.Set("string", codec.Encode[int64](42))

	i64_0, err := isc.ArgAt[int64](params, 2)
	checkFull(ctx, err)
	if i64_0 != 0 {
		ctx.Log().Panicf("wrong int64_0")
	}
	ret.Set("string", codec.Encode[int64](0))

	hash, err := isc.ArgAt[hashing.HashValue](params, 3)
	checkFull(ctx, err)
	if hash != hashing.HashStrings("Hash") {
		ctx.Log().Panicf("wrong hash")
	}
	hname, err := isc.ArgAt[isc.Hname](params, 4)
	checkFull(ctx, err)
	if hname != isc.Hn("Hname") {
		ctx.Log().Panicf("wrong hname")
	}
	hname0, err := isc.ArgAt[isc.Hname](params, 5)
	checkFull(ctx, err)
	if hname0 != 0 {
		ctx.Log().Panicf("wrong Hname-0")
	}
	_, err = isc.ArgAt[isc.AgentID](params, 6)
	checkFull(ctx, err)

	_, err = isc.ArgAt[isc.ChainID](params, 7)
	checkFull(ctx, err)

	_, err = isc.ArgAt[*cryptolib.Address](params, 8)
	checkFull(ctx, err)

	_, err = isc.ArgAt[isc.AgentID](params, 9)
	checkFull(ctx, err)
	return nil
}

func passTypesView(ctx isc.SandboxView) isc.CallArguments {
	params := ctx.Params()
	s, err := isc.ArgAt[string](params, 0)
	checkView(ctx, err)
	if s != "string" {
		ctx.Log().Panicf("wrong string")
	}
	i64, err := isc.ArgAt[int64](params, 1)
	checkView(ctx, err)
	if i64 != 42 {
		ctx.Log().Panicf("wrong int64")
	}
	i64_0, err := isc.ArgAt[int64](params, 2)
	checkView(ctx, err)
	if i64_0 != 0 {
		ctx.Log().Panicf("wrong int64_0")
	}
	hash, err := isc.ArgAt[hashing.HashValue](params, 3)
	checkView(ctx, err)
	if hash != hashing.HashStrings("Hash") {
		ctx.Log().Panicf("wrong hash")
	}
	hname, err := isc.ArgAt[isc.Hname](params, 4)
	checkView(ctx, err)
	if hname != isc.Hn("Hname") {
		ctx.Log().Panicf("wrong hname")
	}
	hname0, err := isc.ArgAt[isc.Hname](params, 5)
	checkView(ctx, err)
	if hname0 != 0 {
		ctx.Log().Panicf("wrong hname-0")
	}
	_, err = isc.ArgAt[isc.AgentID](params, 6)
	checkView(ctx, err)

	_, err = isc.ArgAt[isc.ChainID](params, 7)
	checkView(ctx, err)

	_, err = isc.ArgAt[*cryptolib.Address](params, 8)
	checkView(ctx, err)

	_, err = isc.ArgAt[isc.AgentID](params, 9)
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
