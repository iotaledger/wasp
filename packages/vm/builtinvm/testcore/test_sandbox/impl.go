package test_sandbox

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/root"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

func initialize(ctx vmtypes.Sandbox) (dict.Dict, error) {
	return nil, nil
}

// testChainLogGenericData is called several times in log_test.go
func testChainLogGenericData(ctx vmtypes.Sandbox) (dict.Dict, error) {
	params := ctx.Params()
	inc, ok, err := codec.DecodeInt64(params.MustGet(VarCounter))
	if err != nil {
		return nil, err
	}
	if !ok {
		inc = 1
	}
	ctx.ChainLog([]byte(fmt.Sprintf("[GenericData] Counter Number: %d", inc)))
	return nil, nil
}

func testChainLogEventData(ctx vmtypes.Sandbox) (dict.Dict, error) {
	ctx.Event("[Event] - Testing Event...")
	return nil, nil
}

func testChainLogEventDataFormatted(ctx vmtypes.Sandbox) (dict.Dict, error) {
	params := ctx.Params()
	inc, ok, err := codec.DecodeInt64(params.MustGet(VarCounter))
	if err != nil {
		return nil, err
	}
	if !ok {
		inc = 1
	}
	ctx.Eventf("[Eventf] - (%d) - Testing Event...", inc)

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

	cCreator := ctx.ChainID()

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

func testChainlogDeploy(ctx vmtypes.Sandbox) (dict.Dict, error) {
	//Deploy the same contract with another name
	err := ctx.DeployContract(Interface.ProgramHash,
		VarContractNameDeployed, "test contract deploy log", nil)
	if err != nil {
		return nil, err
	}

	return nil, nil
}
