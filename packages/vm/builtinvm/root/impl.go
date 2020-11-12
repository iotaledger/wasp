// factory implement processor which is always present at the index 0
// it initializes and operates contract registry: creates contracts and provides search
package root

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

// initialize is a handler for the "initialize" request
// It stores chain ID in the state and creates record for root contract in the contract registry at 0 index
func initialize(ctx vmtypes.Sandbox) (codec.ImmutableCodec, error) {
	params := ctx.Params()
	ctx.Eventf("root.initialize.begin")
	state := ctx.AccessState()
	if state.Get(VarStateInitialized) != nil {
		return nil, fmt.Errorf("root.initialize.fail: already_initialized")
	}
	chainID, ok, err := params.GetChainID(ParamChainID)
	if err != nil {
		return nil, fmt.Errorf("root.initialize.fail: can't read expected request argument '%s': %s", ParamChainID, err.Error())
	}
	chainDescription, ok, err := params.GetString(ParamDescription)
	if err != nil {
		return nil, fmt.Errorf("root.initialize.fail: can't read expected request argument '%s': %s", ParamDescription, err.Error())
	}
	if !ok {
		chainDescription = "M/A"
	}
	contractRegistry := state.GetMap(VarContractRegistry)
	if contractRegistry.Len() != 0 {
		return nil, fmt.Errorf("root.initialize.fail: registry_not_empty")
	}
	state.Set(VarStateInitialized, []byte{0xFF})
	state.SetChainID(VarChainID, chainID)
	state.SetString(VarDescription, chainDescription)
	contractRegistry.SetAt(Hname.Bytes(), EncodeContractRecord(GetRootContractRecord()))
	ctx.Eventf("root.initialize.success hname = %s", Hname.String())
	return nil, nil
}

func deployContract(ctx vmtypes.Sandbox) (codec.ImmutableCodec, error) {
	ctx.Eventf("root.deployContract.begin")

	if ctx.AccessState().Get(VarStateInitialized) == nil {
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
	initParams := codec.NewCodec(dict.NewDict())
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
	contractRegistry := ctx.State().GetMap(VarContractRegistry)
	retBin := contractRegistry.GetAt(hname.Bytes())
	if retBin == nil {
		return nil, fmt.Errorf("contract '%s'  does not exist", hname.String())
	}
	ret := codec.NewCodec(dict.NewDict())
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
	contractRegistry := ctx.State().GetMap(VarRegistryOfBinaries)
	binary := contractRegistry.GetAt(deploymentHash[:])

	ret := codec.NewCodec(dict.NewDict())
	ret.Set(ParamData, binary)
	return ret, nil
}
