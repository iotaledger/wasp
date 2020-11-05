// factory implement processor which is always present at the index 0
// it initializes and operates contract registry: creates contracts and provides search
package root

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

// initialize is a handler for the "initialize" request
// It stores chain ID in the state and creates record for root contract in the contract registry at 0 index
func initialize(ctx vmtypes.Sandbox) (codec.ImmutableCodec, error) {
	params := ctx.Params()
	ctx.Publishf("root.initialize.begin")
	state := ctx.AccessState()
	if state.Get(VarStateInitialized) != nil {
		return nil, fmt.Errorf("root.initialize.fail: already_initialized")
	}
	chainID, ok, err := params.GetChainID(VarChainID)
	if err != nil {
		return nil, fmt.Errorf("root.initialize.fail: can't read expected request argument '%s': %s", VarChainID, err.Error())
	}
	chainDescription, ok, err := params.GetString(VarDescription)
	if err != nil {
		return nil, fmt.Errorf("root.initialize.fail: can't read expected request argument '%s': %s", VarDescription, err.Error())
	}
	if !ok {
		chainDescription = "M/A"
	}
	contractRegistry := state.GetArray(VarContractRegistry)

	if contractRegistry.Len() != 0 {
		return nil, fmt.Errorf("root.initialize.fail: registry_not_empty")
	}
	state.Set(VarStateInitialized, []byte{0xFF})
	state.SetChainID(VarChainID, chainID)
	state.SetString(VarDescription, chainDescription)

	// at index 0 always this contract
	contractRegistry.Push(EncodeContractRecord(GetRootContractRecord()))
	ctx.Publishf("root.initialize.success")
	return nil, nil
}

func deployContract(ctx vmtypes.Sandbox) (codec.ImmutableCodec, error) {
	ctx.Publishf("root.deployContract.begin")

	if ctx.AccessState().Get(VarStateInitialized) == nil {
		return nil, fmt.Errorf("root.initialize.fail: not_initialized")
	}
	params := ctx.Params()

	vmtype, ok, err := params.GetString(ParamVMType)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("VMType undefined")
	}
	programBinary, err := params.Get(ParamProgramBinary)
	if err != nil {
		return nil, err
	}
	if len(programBinary) == 0 {
		return nil, fmt.Errorf("programBinary undefined")
	}
	description, ok, err := params.GetString(ParamDescription)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("description undefined")
	}

	// pass to init function all params not consumed so far
	initParams := codec.NewCodec(dict.NewDict())
	err = params.Iterate("", func(key kv.Key, value []byte) bool {
		if key != ParamVMType && key != ParamProgramBinary && key != ParamDescription {
			initParams.Set(key, value)
		}
		return true
	})
	contractIndex, err := ctx.DeployContract(vmtype, programBinary, description, initParams)
	ret := codec.NewCodec(dict.NewDict())
	ret.SetInt64("index", int64(contractIndex))
	return ret, nil
}

func findContract(ctx vmtypes.Sandbox) (codec.ImmutableCodec, error) {
	ctx.Publishf("root.findContract.begin")
	if ctx.AccessState().Get(VarStateInitialized) == nil {
		return nil, fmt.Errorf("root.initialize.fail: not_initialized")
	}

	params := ctx.Params()

	contractIndex, ok, err := params.GetInt64(ParamIndex)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("parameter 'index' undefined")
	}
	contractRegistry := ctx.AccessState().GetArray(VarContractRegistry)
	if contractIndex >= int64(contractRegistry.Len()) {
		return nil, fmt.Errorf("wrong index")
	}
	ret := codec.NewCodec(dict.NewDict())
	ret.Set("data", contractRegistry.GetAt(uint16(contractIndex)))
	return ret, nil
}

func getBinary(ctx vmtypes.Sandbox) (codec.ImmutableCodec, error) {
	ctx.Publishf("root.getBinary.begin")
	if ctx.AccessState().Get(VarStateInitialized) == nil {
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
	contractRegistry := ctx.AccessState().GetMap(VarRegistryOfBinaries)
	binary := contractRegistry.GetAt(deploymentHash[:])

	ret := codec.NewCodec(dict.NewDict())
	ret.Set("data", binary)
	return ret, nil
}
