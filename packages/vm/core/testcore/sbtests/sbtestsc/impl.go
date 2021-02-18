package sbtestsc

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/root"
)

func initialize(ctx coretypes.Sandbox) (dict.Dict, error) {
	if p, err := ctx.Params().Get(ParamFail); err == nil && p != nil {
		return nil, fmt.Errorf("failing on purpose")
	}
	return nil, nil
}

// testEventLogGenericData is called several times in log_test.go
func testEventLogGenericData(ctx coretypes.Sandbox) (dict.Dict, error) {
	params := ctx.Params()
	inc, ok, err := codec.DecodeInt64(params.MustGet(VarCounter))
	if err != nil {
		return nil, err
	}
	if !ok {
		inc = 1
	}
	ctx.Event(fmt.Sprintf("[GenericData] Counter Number: %d", inc))
	return nil, nil
}

func testEventLogEventData(ctx coretypes.Sandbox) (dict.Dict, error) {
	ctx.Event("[Event] - Testing Event...")
	return nil, nil
}

func testChainOwnerIDView(ctx coretypes.SandboxView) (dict.Dict, error) {
	cOwnerID := ctx.ChainOwnerID()
	ret := dict.New()
	ret.Set(ParamChainOwnerID, cOwnerID.Bytes())

	return ret, nil
}

func testChainOwnerIDFull(ctx coretypes.Sandbox) (dict.Dict, error) {
	cOwnerID := ctx.ChainOwnerID()
	ret := dict.New()
	ret.Set(ParamChainOwnerID, cOwnerID.Bytes())

	return ret, nil
}

func testContractIDView(ctx coretypes.SandboxView) (dict.Dict, error) {
	cID := ctx.ContractID()
	ret := dict.New()
	ret.Set(VarContractID, cID[:])
	return ret, nil
}

func testContractIDFull(ctx coretypes.Sandbox) (dict.Dict, error) {
	cID := ctx.ContractID()
	ret := dict.New()
	ret.Set(VarContractID, cID[:])
	return ret, nil
}

func testSandboxCall(ctx coretypes.SandboxView) (dict.Dict, error) {
	ret, err := ctx.Call(root.Interface.Hname(), coretypes.Hn(root.FuncGetChainInfo), nil)
	if err != nil {
		return nil, err
	}
	desc := ret.MustGet(root.VarDescription)

	ret.Set(VarSandboxCall, desc)

	return ret, nil
}

func testEventLogDeploy(ctx coretypes.Sandbox) (dict.Dict, error) {
	//Deploy the same contract with another name
	err := ctx.DeployContract(Interface.ProgramHash,
		VarContractNameDeployed, "test contract deploy log", nil)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func testPanicFullEP(ctx coretypes.Sandbox) (dict.Dict, error) {
	ctx.Log().Panicf(MsgFullPanic)
	return nil, nil
}

func testPanicViewEP(ctx coretypes.SandboxView) (dict.Dict, error) {
	ctx.Log().Panicf(MsgViewPanic)
	return nil, nil
}

func testJustView(ctx coretypes.SandboxView) (dict.Dict, error) {
	ctx.Log().Infof("calling empty view entry point")
	return nil, nil
}

func testCallPanicFullEP(ctx coretypes.Sandbox) (dict.Dict, error) {
	ctx.Log().Infof("will be calling entry point '%s' from full EP", FuncPanicFullEP)
	return ctx.Call(Interface.Hname(), coretypes.Hn(FuncPanicFullEP), nil, nil)
}

func testCallPanicViewEPFromFull(ctx coretypes.Sandbox) (dict.Dict, error) {
	ctx.Log().Infof("will be calling entry point '%s' from full EP", FuncPanicViewEP)
	return ctx.Call(Interface.Hname(), coretypes.Hn(FuncPanicViewEP), nil, nil)
}

func testCallPanicViewEPFromView(ctx coretypes.SandboxView) (dict.Dict, error) {
	ctx.Log().Infof("will be calling entry point '%s' from view EP", FuncPanicViewEP)
	return ctx.Call(Interface.Hname(), coretypes.Hn(FuncPanicViewEP), nil)
}

func doNothing(ctx coretypes.Sandbox) (dict.Dict, error) {
	if ctx.IncomingTransfer().Len() == 0 {
		ctx.Log().Infof(MsgDoNothing)
	} else {
		ctx.Log().Infof(MsgDoNothing+" with transfer\n%s", ctx.IncomingTransfer().String())
	}
	return nil, nil
}

// sendToAddress send the whole account to ParamAddress if invoked by the creator
// Panics if wrong parameter or unauthorized access
func sendToAddress(ctx coretypes.Sandbox) (dict.Dict, error) {
	ctx.Log().Infof(FuncSendToAddress)
	if ctx.Caller() != ctx.ContractCreator() {
		ctx.Log().Panicf(MsgPanicUnauthorized)
	}
	targetAddress, ok, err := codec.DecodeAddress(ctx.Params().MustGet(ParamAddress))
	if err != nil || !ok {
		ctx.Log().Panicf("wrong parameter '%s'", ParamAddress)
	}
	myTokens := ctx.Balances()
	succ := ctx.TransferToAddress(targetAddress, myTokens)
	if !succ {
		ctx.Log().Panicf("failed send to %s: tokens:\n%s", targetAddress, myTokens.String())
	}
	return nil, nil
}
