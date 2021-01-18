package test_sandbox

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

func initialize(ctx vmtypes.Sandbox) (dict.Dict, error) {
	return nil, nil
}

func testCheckContextFromFullEP(ctx vmtypes.Sandbox) (dict.Dict, error) {
	par := ctx.Params()
	chainID, ok, err := codec.DecodeChainID(par.MustGet(ParamChainID))
	if err != nil || !ok || chainID != ctx.ContractID().ChainID() {
		return nil, fmt.Errorf("wrong '%s'", ParamChainID)
	}
	chainOwnerID, ok, err := codec.DecodeAgentID(par.MustGet(ParamChainOwnerID))
	if err != nil || !ok || chainOwnerID != ctx.ChainOwnerID() {
		return nil, fmt.Errorf("wrong '%s'", ParamChainOwnerID)
	}
	caller, ok, err := codec.DecodeAgentID(par.MustGet(ParamCaller))
	if err != nil || !ok || caller != ctx.Caller() {
		return nil, fmt.Errorf("wrong '%s'", ParamCaller)
	}
	contractID, ok, err := codec.DecodeContractID(par.MustGet(ParamContractID))
	if err != nil || !ok || contractID != ctx.ContractID() {
		return nil, fmt.Errorf("wrong '%s'", ParamContractID)
	}
	agentID, ok, err := codec.DecodeAgentID(par.MustGet(ParamAgentID))
	if err != nil || !ok || agentID != coretypes.NewAgentIDFromContractID(ctx.ContractID()) {
		return nil, fmt.Errorf("wrong '%s'", ParamAgentID)
	}
	contractCreator, ok, err := codec.DecodeAgentID(par.MustGet(ParamContractCreator))
	if err != nil || !ok || contractCreator != ctx.ContractCreator() {
		return nil, fmt.Errorf("wrong '%s'", ParamContractCreator)
	}
	return nil, nil
}

func testCheckContextFromViewEP(ctx vmtypes.SandboxView) (dict.Dict, error) {
	par := ctx.Params()
	chainID, ok, err := codec.DecodeChainID(par.MustGet(ParamChainID))
	if err != nil || !ok || chainID != ctx.ContractID().ChainID() {
		return nil, fmt.Errorf("wrong '%s'", ParamChainID)
	}
	contractID, ok, err := codec.DecodeContractID(par.MustGet(ParamContractID))
	if err != nil || !ok || contractID != ctx.ContractID() {
		return nil, fmt.Errorf("wrong '%s'", ParamContractID)
	}
	return nil, nil
}

// testEventLogGenericData is called several times in log_test.go
func testEventLogGenericData(ctx vmtypes.Sandbox) (dict.Dict, error) {
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

func testEventLogEventData(ctx vmtypes.Sandbox) (dict.Dict, error) {
	ctx.Event("[Event] - Testing Event...")
	return nil, nil
}

//The purpose of this function is to test Sandbox ChainOwnerID (It's not a ViewCall because ChainOwnerID is not in the SandboxView)
func testChainOwnerID(ctx vmtypes.Sandbox) (dict.Dict, error) {

	cOwnerID := ctx.ChainOwnerID()

	ret := dict.New()
	ret.Set(VarChainOwner, cOwnerID.Bytes())

	return ret, nil
}

//The purpose of this function is to test Sandbox ChainID
func testChainID(ctx vmtypes.SandboxView) (dict.Dict, error) {

	cCreator := ctx.ContractID().ChainID()

	ret := dict.New()
	ret.Set(VarChainID, cCreator.Bytes())

	return ret, nil
}

func testSandboxCall(ctx vmtypes.SandboxView) (dict.Dict, error) {
	ret, err := ctx.Call(root.Interface.Hname(), coretypes.Hn(root.FuncGetChainInfo), nil)
	if err != nil {
		return nil, err
	}
	desc := ret.MustGet(root.VarDescription)

	ret.Set(VarSandboxCall, desc)

	return ret, nil
}

func testEventLogDeploy(ctx vmtypes.Sandbox) (dict.Dict, error) {
	//Deploy the same contract with another name
	err := ctx.DeployContract(Interface.ProgramHash,
		VarContractNameDeployed, "test contract deploy log", nil)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func testPanicFullEP(ctx vmtypes.Sandbox) (dict.Dict, error) {
	ctx.Log().Panicf(MsgFullPanic)
	return nil, nil
}

func testPanicViewEP(ctx vmtypes.SandboxView) (dict.Dict, error) {
	ctx.Log().Panicf(MsgViewPanic)
	return nil, nil
}

func testJustView(ctx vmtypes.SandboxView) (dict.Dict, error) {
	ctx.Log().Infof("calling empty view entry point")
	return nil, nil
}

func testCallPanicFullEP(ctx vmtypes.Sandbox) (dict.Dict, error) {
	ctx.Log().Infof("will be calling entry point '%s' from full EP", FuncPanicFullEP)
	return ctx.Call(Interface.Hname(), coretypes.Hn(FuncPanicFullEP), nil, nil)
}

func testCallPanicViewEPFromFull(ctx vmtypes.Sandbox) (dict.Dict, error) {
	ctx.Log().Infof("will be calling entry point '%s' from full EP", FuncPanicViewEP)
	return ctx.Call(Interface.Hname(), coretypes.Hn(FuncPanicViewEP), nil, nil)
}

func testCallPanicViewEPFromView(ctx vmtypes.SandboxView) (dict.Dict, error) {
	ctx.Log().Infof("will be calling entry point '%s' from view EP", FuncPanicViewEP)
	return ctx.Call(Interface.Hname(), coretypes.Hn(FuncPanicViewEP), nil)
}

func doNothing(ctx vmtypes.Sandbox) (dict.Dict, error) {
	if ctx.IncomingTransfer().Len() == 0 {
		ctx.Log().Infof(MsgDoNothing)
	} else {
		ctx.Log().Infof(MsgDoNothing+" with transfer\n%s", ctx.IncomingTransfer().String())
	}
	return nil, nil
}

// sendToAddress send the whole account to ParamAddress if invoked by the creator
// Panics if wrong parameter or unauthorized access
func sendToAddress(ctx vmtypes.Sandbox) (dict.Dict, error) {
	ctx.Log().Infof(FuncSendToAddress)
	if ctx.Caller() != ctx.ContractCreator() {
		ctx.Log().Panicf("-------- panic due to unauthorized call")
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
