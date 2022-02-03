package sbtestsc

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
)

func initialize(ctx iscp.Sandbox) dict.Dict {
	p := ctx.Params().MustGet(ParamFail)
	ctx.Requiref(p == nil, "failing on purpose")
	return nil
}

// testEventLogGenericData is called several times in log_test.go
func testEventLogGenericData(ctx iscp.Sandbox) dict.Dict {
	params := ctx.Params()
	inc := codec.MustDecodeInt64(params.MustGet(VarCounter), 1)
	ctx.Event(fmt.Sprintf("[GenericData] Counter Number: %d", inc))
	return nil
}

func testEventLogEventData(ctx iscp.Sandbox) dict.Dict {
	ctx.Event("[Event] - Testing Event...")
	return nil
}

func testChainOwnerIDView(ctx iscp.SandboxView) dict.Dict {
	cOwnerID := ctx.ChainOwnerID()
	return dict.Dict{ParamChainOwnerID: cOwnerID.Bytes()}
}

func testChainOwnerIDFull(ctx iscp.Sandbox) dict.Dict {
	cOwnerID := ctx.ChainOwnerID()
	return dict.Dict{ParamChainOwnerID: cOwnerID.Bytes()}
}

func testSandboxCall(ctx iscp.SandboxView) dict.Dict {
	ret := ctx.Call(governance.Contract.Hname(), governance.FuncGetChainInfo.Hname(), nil)
	desc := ret.MustGet(governance.VarDescription)
	ret.Set(VarSandboxCall, desc)
	return ret
}

func testEventLogDeploy(ctx iscp.Sandbox) dict.Dict {
	// Deploy the same contract with another name
	ctx.DeployContract(Contract.ProgramHash,
		VarContractNameDeployed, "test contract deploy log", nil)
	return nil
}

func testPanicFullEP(ctx iscp.Sandbox) dict.Dict {
	ctx.Log().Panicf(MsgFullPanic)
	return nil
}

func testPanicViewEP(ctx iscp.SandboxView) dict.Dict {
	ctx.Log().Panicf(MsgViewPanic)
	return nil
}

func testJustView(ctx iscp.SandboxView) dict.Dict {
	ctx.Log().Infof("calling empty view entry point")
	return nil
}

func testCallPanicFullEP(ctx iscp.Sandbox) dict.Dict {
	ctx.Log().Infof("will be calling entry point '%s' from full EP", FuncPanicFullEP)
	return ctx.Call(Contract.Hname(), FuncPanicFullEP.Hname(), nil, nil)
}

func testCallPanicViewEPFromFull(ctx iscp.Sandbox) dict.Dict {
	ctx.Log().Infof("will be calling entry point '%s' from full EP", FuncPanicViewEP)
	return ctx.Call(Contract.Hname(), FuncPanicViewEP.Hname(), nil, nil)
}

func testCallPanicViewEPFromView(ctx iscp.SandboxView) dict.Dict {
	ctx.Log().Infof("will be calling entry point '%s' from view EP", FuncPanicViewEP)
	return ctx.Call(Contract.Hname(), FuncPanicViewEP.Hname(), nil)
}

func doNothing(ctx iscp.Sandbox) dict.Dict {
	ctx.Log().Infof(MsgDoNothing)
	return nil
}

// sendToAddress send the whole account to ParamAddress if invoked by the creator
// Panics if wrong parameter or unauthorized access
//func sendToAddress(ctx iscp.Sandbox) (dict.Dict) {
//	ctx.Log().Infof(FuncSendToAddress.Name)
//	a := assert.NewAssert(ctx.Log())
//	par := kvdecoder.New(ctx.Params(), ctx.Log())
//	a.Requiref(ctx.Caller().Equals(ctx.ContractCreator()), MsgPanicUnauthorized)
//	targetAddress := par.MustGetAddress(ParamAddress)
//	myTokens := ctx.Balances()
//	a.Requiref(ctx.Send(targetAddress, myTokens, nil),
//		fmt.Sprintf("failed send to %s: tokens:\n%s", targetAddress, myTokens.String()))
//
//	ctx.Log().Infof("sent to %s: tokens:\n%s", targetAddress.Base58(), myTokens.String())
//	return nil
//}
