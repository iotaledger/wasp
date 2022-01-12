package sbtestsc

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/assert"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/kvdecoder"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
)

func initialize(ctx iscp.Sandbox) (dict.Dict, error) {
	if p, err := ctx.Params().Get(ParamFail); err == nil && p != nil {
		return nil, fmt.Errorf("failing on purpose")
	}
	return nil, nil
}

// testEventLogGenericData is called several times in log_test.go
func testEventLogGenericData(ctx iscp.Sandbox) (dict.Dict, error) {
	params := ctx.Params()
	inc, err := codec.DecodeInt64(params.MustGet(VarCounter), 1)
	if err != nil {
		return nil, err
	}
	ctx.Event(fmt.Sprintf("[GenericData] Counter Number: %d", inc))
	return nil, nil
}

func testEventLogEventData(ctx iscp.Sandbox) (dict.Dict, error) {
	ctx.Event("[Event] - Testing Event...")
	return nil, nil
}

func testChainOwnerIDView(ctx iscp.SandboxView) (dict.Dict, error) {
	cOwnerID := ctx.ChainOwnerID()
	ret := dict.New()
	ret.Set(ParamChainOwnerID, cOwnerID.Bytes())

	return ret, nil
}

func testChainOwnerIDFull(ctx iscp.Sandbox) (dict.Dict, error) {
	cOwnerID := ctx.ChainOwnerID()
	ret := dict.New()
	ret.Set(ParamChainOwnerID, cOwnerID.Bytes())

	return ret, nil
}

func testSandboxCall(ctx iscp.SandboxView) (dict.Dict, error) {
	ret, err := ctx.Call(governance.Contract.Hname(), governance.FuncGetChainInfo.Hname(), nil)
	if err != nil {
		return nil, err
	}
	desc := ret.MustGet(governance.VarDescription)

	ret.Set(VarSandboxCall, desc)

	return ret, nil
}

func testEventLogDeploy(ctx iscp.Sandbox) (dict.Dict, error) {
	// Deploy the same contract with another name
	err := ctx.DeployContract(Contract.ProgramHash,
		VarContractNameDeployed, "test contract deploy log", nil)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func testPanicFullEP(ctx iscp.Sandbox) (dict.Dict, error) {
	ctx.Log().Panicf(MsgFullPanic)
	return nil, nil
}

func testPanicViewEP(ctx iscp.SandboxView) (dict.Dict, error) {
	ctx.Log().Panicf(MsgViewPanic)
	return nil, nil
}

func testJustView(ctx iscp.SandboxView) (dict.Dict, error) {
	ctx.Log().Infof("calling empty view entry point")
	return nil, nil
}

func testCallPanicFullEP(ctx iscp.Sandbox) (dict.Dict, error) {
	ctx.Log().Infof("will be calling entry point '%s' from full EP", FuncPanicFullEP)
	return ctx.Call(Contract.Hname(), FuncPanicFullEP.Hname(), nil, nil)
}

func testCallPanicViewEPFromFull(ctx iscp.Sandbox) (dict.Dict, error) {
	ctx.Log().Infof("will be calling entry point '%s' from full EP", FuncPanicViewEP)
	return ctx.Call(Contract.Hname(), FuncPanicViewEP.Hname(), nil, nil)
}

func testCallPanicViewEPFromView(ctx iscp.SandboxView) (dict.Dict, error) {
	ctx.Log().Infof("will be calling entry point '%s' from view EP", FuncPanicViewEP)
	return ctx.Call(Contract.Hname(), FuncPanicViewEP.Hname(), nil)
}

func doNothing(ctx iscp.Sandbox) (dict.Dict, error) {
	if len(ctx.Allowance()) == 0 {
		ctx.Log().Infof(MsgDoNothing)
	} else {
		ctx.Log().Infof(MsgDoNothing+" with transfer\n%s", ctx.Allowance().String())
	}
	return nil, nil
}

// sendToAddress send the whole account to ParamAddress if invoked by the creator
// Panics if wrong parameter or unauthorized access
func sendToAddress(ctx iscp.Sandbox) (dict.Dict, error) {
	ctx.Log().Infof(FuncSendToAddress.Name)
	a := assert.NewAssert(ctx.Log())
	par := kvdecoder.New(ctx.Params(), ctx.Log())
	a.Requiref(ctx.Caller().Equals(ctx.ContractCreator()), MsgPanicUnauthorized)
	targetAddress := par.MustGetAddress(ParamAddress)
	myTokens := ctx.Balances()
	a.Requiref(ctx.Send(targetAddress, myTokens, nil),
		fmt.Sprintf("failed send to %s: tokens:\n%s", targetAddress, myTokens.String()))

	ctx.Log().Infof("sent to %s: tokens:\n%s", targetAddress.Base58(), myTokens.String())
	return nil, nil
}
