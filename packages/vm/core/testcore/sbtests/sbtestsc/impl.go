package sbtestsc

import (
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
)

var testError *isc.VMErrorTemplate

func initialize(ctx isc.Sandbox) isc.CallArguments {
	p := ctx.Params().Get(ParamFail)
	ctx.Requiref(p == nil, "failing on purpose")
	testError = ctx.RegisterError("ERROR_TEST")
	return nil
}

// testEventLogGenericData is called several times in log_test.go
func testEventLogGenericData(ctx isc.Sandbox) isc.CallArguments {
	params := ctx.Params()
	inc := codec.MustDecode[uint64](params.Get(VarCounter), 1)
	eventCounter(ctx, inc)
	return nil
}

func testEventLogEventData(ctx isc.Sandbox) isc.CallArguments {
	eventTest(ctx)
	return nil
}

func testChainOwnerIDView(ctx isc.SandboxView) isc.CallArguments {
	cOwnerID := ctx.ChainOwnerID()
	return dict.Dict{ParamChainOwnerID: cOwnerID.Bytes()}
}

func testChainOwnerIDFull(ctx isc.Sandbox) isc.CallArguments {
	cOwnerID := ctx.ChainOwnerID()
	return dict.Dict{ParamChainOwnerID: cOwnerID.Bytes()}
}

func testSandboxCall(ctx isc.SandboxView) isc.CallArguments {
	return ctx.CallView(governance.ViewGetChainInfo.Message())
}

func testEventLogDeploy(ctx isc.Sandbox) isc.CallArguments {
	// Deploy the same contract with another name
	ctx.DeployContract(Contract.ProgramHash, VarContractNameDeployed, nil)
	return nil
}

func testPanicFullEP(ctx isc.Sandbox) isc.CallArguments {
	ctx.Log().Panicf(MsgFullPanic)
	return nil
}

func testCustomError(_ isc.Sandbox) dict.Dict {
	panic(testError.Create("CUSTOM_ERROR"))
}

func testPanicViewEP(ctx isc.SandboxView) isc.CallArguments {
	ctx.Log().Panicf(MsgViewPanic)
	return nil
}

func testJustView(ctx isc.SandboxView) isc.CallArguments {
	ctx.Log().Infof("calling empty view entry point")
	return nil
}

func testCallPanicFullEP(ctx isc.Sandbox) isc.CallArguments {
	ctx.Log().Infof("will be calling entry point '%s' from full EP", FuncPanicFullEP)
	return ctx.Call(isc.NewMessage(Contract.Hname(), FuncPanicFullEP.Hname(), nil), nil)
}

func testCallPanicViewEPFromFull(ctx isc.Sandbox) isc.CallArguments {
	ctx.Log().Infof("will be calling entry point '%s' from full EP", FuncPanicViewEP)
	return ctx.Call(isc.NewMessage(Contract.Hname(), FuncPanicViewEP.Hname(), nil), nil)
}

func testCallPanicViewEPFromView(ctx isc.SandboxView) isc.CallArguments {
	ctx.Log().Infof("will be calling entry point '%s' from view EP", FuncPanicViewEP)
	return ctx.CallView(isc.NewMessage(Contract.Hname(), FuncPanicViewEP.Hname(), nil))
}

func doNothing(ctx isc.Sandbox) isc.CallArguments {
	ctx.Log().Infof(MsgDoNothing)
	return nil
}
