// factory implement processor which is always present at the index 0
// it initializes and operates contract registry: creates contracts and provides search
package root

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/accountsc"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

// initialize is a handler for the "init" request
// It stores chain ID in the state and creates record for root contract in the contract registry
func initialize(ctx vmtypes.Sandbox) (codec.ImmutableCodec, error) {
	params := ctx.Params()
	ctx.Eventf("root.initialize.begin")
	state := ctx.State()
	if state.Get(VarStateInitialized) != nil {
		// can't be initialized twice
		return nil, fmt.Errorf("root.initialize.fail: already_initialized")
	}
	// retrieving init parameters
	// -- chain ID
	chainID, ok, err := params.GetChainID(ParamChainID)
	if err != nil {
		return nil, fmt.Errorf("root.initialize.fail: can't read expected request argument '%s': %s", ParamChainID, err.Error())
	}
	// -- description
	chainDescription, ok, err := params.GetString(ParamDescription)
	if err != nil {
		return nil, fmt.Errorf("root.initialize.fail: can't read expected request argument '%s': %s", ParamDescription, err.Error())
	}
	if !ok {
		chainDescription = "M/A"
	}
	sender := ctx.Caller()

	contractRegistry := state.GetMap(VarContractRegistry)
	if contractRegistry.Len() != 0 {
		return nil, fmt.Errorf("root.initialize.fail: registry_not_empty")
	}
	state.Set(VarStateInitialized, []byte{0xFF})
	state.SetChainID(VarChainID, chainID)
	state.SetAgentID(VarChainOwnerID, &sender) // chain owner is whoever sends init request
	state.SetString(VarDescription, chainDescription)
	contractRegistry.SetAt(Hname.Bytes(), EncodeContractRecord(&RootContractRecord))

	err = ctx.DeployContract(
		"builtinvm",
		accountsc.ProgramHash[:],
		accountsc.FullName,
		accountsc.Description,
		nil,
	)
	if err != nil {
		return nil, err
	}
	ctx.Eventf("root.initialize.success hname = %s", Hname.String())
	return nil, nil
}

func deployContract(ctx vmtypes.Sandbox) (codec.ImmutableCodec, error) {
	ctx.Eventf("root.deployContract.begin")

	if ctx.State().Get(VarStateInitialized) == nil {
		return nil, fmt.Errorf("root.initialize.fail: not_initialized")
	}
	params := ctx.Params()

	vmtype, ok, err := params.GetString(ParamVMType)
	if err != nil {
		ctx.Eventf("root.deployContract.error 1: %v", err)
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("root.deployContract.error: VMType undefined")
	}

	programBinary, err := params.Get(ParamProgramBinary)
	if err != nil {
		ctx.Eventf("root.deployContract.error 2: %v", err)
		return nil, err
	}
	if len(programBinary) == 0 {
		return nil, fmt.Errorf("root.deployContract.begin: programBinary undefined")
	}
	description, ok, err := params.GetString(ParamDescription)
	if err != nil {
		ctx.Eventf("root.deployContract.error 3: %v", err)
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("root.deployContract.begin: description undefined")
	}
	name, ok, err := params.GetString(ParamName)
	if err != nil {
		ctx.Eventf("root.deployContract.error 4: %v", err)
		return nil, err
	}
	if !ok || name == "" {
		return nil, fmt.Errorf("incorrect contract name")
	}
	// pass to init function all params not consumed so far
	initParams := codec.NewCodec(dict.New())
	err = params.Iterate("", func(key kv.Key, value []byte) bool {
		if key != ParamVMType && key != ParamProgramBinary && key != ParamDescription {
			initParams.Set(key, value)
		}
		return true
	})
	err = ctx.DeployContract(vmtype, programBinary, name, description, initParams)
	if err != nil {
		return nil, fmt.Errorf("root.deployContract: %v", err)
	}
	ctx.Eventf("root.deployContract.success. Deployed contract hname = %s, name = '%s'",
		coretypes.Hn(name).String(), name)
	return nil, nil
}

// findContract is a view
func findContract(ctx vmtypes.SandboxView) (codec.ImmutableCodec, error) {
	if ctx.State().Get(VarStateInitialized) == nil {
		return nil, fmt.Errorf("root.initialize.fail: not_initialized")
	}
	params := ctx.Params()
	hname, ok, err := params.GetHname(ParamHname)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("parameter 'hname' undefined")
	}
	rec, err := FindContract(ctx.State(), hname)
	if err != nil {
		return nil, err
	}
	retBin := EncodeContractRecord(rec)
	ret := codec.NewCodec(dict.New())
	ret.Set(ParamData, retBin)
	return ret, nil
}

// getBinary is
func getBinary(ctx vmtypes.SandboxView) (codec.ImmutableCodec, error) {
	if ctx.State().Get(VarStateInitialized) == nil {
		return nil, fmt.Errorf("root.initialize.fail: not_initialized")
	}
	params := ctx.Params()
	deploymentHash, ok, err := params.GetHashValue(ParamHash)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("parameter 'hash' undefined")
	}

	bin, err := GetBinary(ctx.State(), *deploymentHash)
	if err != nil {
		return nil, err
	}
	ret := codec.NewCodec(dict.New())
	ret.Set(ParamData, bin)
	return ret, nil
}

func changeChainOwner(ctx vmtypes.Sandbox) (codec.ImmutableCodec, error) {
	ctx.Eventf("root.changeChainOwner.begin")

	state := ctx.State()

	currentOwner, _ := state.GetAgentID(VarChainOwnerID)
	if *currentOwner != ctx.Caller() {
		ctx.Eventf("root.changeChainOwner.fail: not authorized")
		return nil, fmt.Errorf("not authorized")
	}
	newOwnerID, ok, err := ctx.Params().GetAgentID(ParamChainOwner)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("wrong parameter")
	}
	state.SetAgentID(VarChainOwnerID, newOwnerID)
	ctx.Eventf("root.changeChainOwner.success: owner changed from %s -> %s", currentOwner.String(), newOwnerID.String())
	return nil, nil
}

func getInfo(ctx vmtypes.SandboxView) (codec.ImmutableCodec, error) {
	d := dict.New()
	c := codec.NewMustCodec(d)

	chainID, _ := ctx.State().GetChainID(VarChainID)
	c.SetChainID(VarChainID, chainID)

	chainOwner, _ := ctx.State().GetAgentID(VarChainOwnerID)
	c.SetAgentID(VarChainOwnerID, chainOwner)

	cr := ctx.State().GetMap(VarContractRegistry)
	cr2 := c.GetMap(VarContractRegistry)
	cr.Iterate(func(elemKey []byte, value []byte) bool {
		cr2.SetAt(elemKey, value)
		return true
	})

	return codec.NewCodec(d), nil
}
