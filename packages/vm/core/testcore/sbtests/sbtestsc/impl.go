package sbtestsc

import (
	"github.com/samber/lo"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
)

// testEventLogGenericData is called several times in log_test.go
func testEventLogGenericData(ctx isc.Sandbox, inc *uint64) {
	incV := lo.FromPtrOr(inc, 1)
	eventCounter(ctx, incV)
}

func testEventLogEventData(ctx isc.Sandbox) {
	eventTest(ctx)
}

func testChainAdminView(ctx isc.SandboxView) isc.AgentID {
	return ctx.ChainAdmin()
}

func testChainAdminFull(ctx isc.Sandbox) isc.AgentID {
	return ctx.ChainAdmin()
}

func testSandboxCall(ctx isc.SandboxView) isc.CallArguments {
	return ctx.CallView(governance.ViewGetChainInfo.Message())
}

func testEventLogDeploy(ctx isc.Sandbox) {
	// Deploy the same contract with another name
	panic("TODO: contract deployment")
	// ctx.DeployContract(Contract.ProgramHash, VarContractNameDeployed, nil)
}

func testPanicFullEP(ctx isc.Sandbox) {
	ctx.Log().Panicf(MsgFullPanic)
}

func testPanicViewEP(ctx isc.SandboxView) {
	ctx.Log().Panicf(MsgViewPanic)
}

func testJustView(ctx isc.SandboxView) {
	ctx.Log().Infof("calling empty view entry point")
}

func testCallPanicFullEP(ctx isc.Sandbox) isc.CallArguments {
	ctx.Log().Infof("will be calling entry point '%s' from full EP", FuncPanicFullEP)
	return ctx.Call(isc.NewMessage(Contract.Hname(), FuncPanicFullEP.Hname(), nil), isc.NewEmptyAssets())
}

func testCallPanicViewEPFromFull(ctx isc.Sandbox) isc.CallArguments {
	ctx.Log().Infof("will be calling entry point '%s' from full EP", FuncPanicViewEP)
	return ctx.Call(isc.NewMessage(Contract.Hname(), FuncPanicViewEP.Hname(), nil), isc.NewEmptyAssets())
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

func stackOverflow(ctx isc.Sandbox) {
	ctx.Call(FuncStackOverflow.Message(), isc.NewEmptyAssets())
}
