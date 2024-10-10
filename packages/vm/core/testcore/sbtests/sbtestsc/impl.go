package sbtestsc

import (
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/samber/lo"
)

var testError *isc.VMErrorTemplate

func initialize(ctx isc.Sandbox) isc.CallArguments {
	failOnPurpose := isc.MustArgAt[bool](ctx.Params(), 0)
	ctx.Requiref(!failOnPurpose, "failing on purpose")
	testError = ctx.RegisterError("ERROR_TEST")
	return nil
}

// testEventLogGenericData is called several times in log_test.go
func testEventLogGenericData(ctx isc.Sandbox, inc *uint64) {
	incV := lo.FromPtrOr(inc, 1)
	eventCounter(ctx, incV)
}

func testEventLogEventData(ctx isc.Sandbox) {
	eventTest(ctx)
}

func testChainOwnerIDView(ctx isc.SandboxView) isc.AgentID {
	return ctx.ChainOwnerID()
}

func testChainOwnerIDFull(ctx isc.Sandbox) isc.AgentID {
	return ctx.ChainOwnerID()
}

func testSandboxCall(ctx isc.SandboxView) isc.CallArguments {
	return ctx.CallView(governance.ViewGetChainInfo.Message())
}

func testEventLogDeploy(ctx isc.Sandbox) {
	// Deploy the same contract with another name
	panic("TODO: contract deployment")
	//ctx.DeployContract(Contract.ProgramHash, VarContractNameDeployed, nil)
}

func testPanicFullEP(ctx isc.Sandbox) {
	ctx.Log().Panicf(MsgFullPanic)
}

func testCustomError(_ isc.Sandbox) isc.CallArguments {
	panic(testError.Create("CUSTOM_ERROR"))
}

func testPanicViewEP(ctx isc.SandboxView) {
	ctx.Log().Panicf(MsgViewPanic)
}

func testJustView(ctx isc.SandboxView) {
	ctx.Log().Infof("calling empty view entry point")
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

func doNothing(ctx isc.Sandbox) {
	ctx.Log().Infof(MsgDoNothing)
}

func callViewFunc(ctx isc.SandboxView) func(isc.Message) (isc.CallArguments, error) {
	return func(m isc.Message) (isc.CallArguments, error) {
		m.Target.Contract = ctx.Contract()
		return ctx.CallView(m), nil
	}
}
