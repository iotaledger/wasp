// factory implement processor which is always present at the index 0
// it initializes and operates contract registry: creates contracts and provides search
package root

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

type rootProcessor struct{}

type rootEntryPoint func(ctx vmtypes.Sandbox) (codec.ImmutableCodec, error)

var (
	processor   = &rootProcessor{}
	ProgramHash = hashing.NilHash
)

func GetProcessor() vmtypes.Processor {
	return processor
}

var (
	EntryPointInitialize   = coretypes.NewEntryPointCodeFromFunctionName("initialize")
	EntryPointNewContract  = coretypes.NewEntryPointCodeFromFunctionName("deployContract")
	EntryPointFindContract = coretypes.NewEntryPointCodeFromFunctionName("findContract")
	EntryPointGetBinary    = coretypes.NewEntryPointCodeFromFunctionName("getBinary")
)

func (v *rootProcessor) GetEntryPoint(code coretypes.EntryPointCode) (vmtypes.EntryPoint, bool) {
	switch code {
	case EntryPointInitialize:
		return (rootEntryPoint)(initialize), true

	case EntryPointNewContract:
		return (rootEntryPoint)(deployContract), true

	case EntryPointFindContract:
		return (rootEntryPoint)(findContract), true

	case EntryPointGetBinary:
		return (rootEntryPoint)(getBinary), true

	}
	return nil, false
}

func (v *rootProcessor) GetDescription() string {
	return "Factory processor"
}

func (ep rootEntryPoint) Call(ctx vmtypes.Sandbox) (codec.ImmutableCodec, error) {
	ret, err := ep(ctx)
	if err != nil {
		ctx.Publishf("error occured: '%v'", err)
	}
	return ret, err
}

func (ep rootEntryPoint) WithGasLimit(_ int) vmtypes.EntryPoint {
	return ep
}

const (
	VarStateInitialized   = "i"
	VarChainID            = "c"
	VarRegistryOfBinaries = "b"
	VarContractRegistry   = "r"
	VarDescription        = "d"
)

const (
	ParamVMType        = "vmtype"
	ParamProgramBinary = "programBinary"
	ParamDescription   = "description"
	ParamIndex         = "index"
	ParamHash          = "hash"
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
	contractIndex, err := ctx.InstallProgram(vmtype, programBinary, description)
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
