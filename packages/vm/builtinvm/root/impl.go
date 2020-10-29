// factory implement processor which is always present at the index 0
// it initializes and operates contract registry: creates contracts and provides search
package root

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
	"io"
)

type factoryProcessor struct{}

type factoryEntryPoint func(ctx vmtypes.Sandbox, params codec.ImmutableCodec) (codec.ImmutableCodec, error)

var (
	processor   = &factoryProcessor{}
	ProgramHash = hashing.AllFHash
)

func GetProcessor() vmtypes.Processor {
	return processor
}

var (
	entryPointInitialize  = coretypes.NewEntryPointCodeFromFunctionName("initialize")
	entryPointNewContract = coretypes.NewEntryPointCodeFromFunctionName("newContract")
)

func (v *factoryProcessor) GetEntryPoint(code coretypes.EntryPointCode) (vmtypes.EntryPoint, bool) {
	switch code {
	case entryPointInitialize:
		return (factoryEntryPoint)(initialize), true

	case entryPointNewContract:
		return (factoryEntryPoint)(newContract), true
	}
	return nil, false
}

func (v *factoryProcessor) GetDescription() string {
	return "Factory processor"
}

func (ep factoryEntryPoint) Call(ctx vmtypes.Sandbox, params codec.ImmutableCodec) (codec.ImmutableCodec, error) {
	ret, err := ep(ctx, params)
	if err != nil {
		ctx.Publishf("error occured: '%v'", err)
	}
	return ret, err
}

func (ep factoryEntryPoint) WithGasLimit(_ int) vmtypes.EntryPoint {
	return ep
}

const (
	VarStateInitialized = "i"
	VarChainID          = "c"
	VarContractRegistry = "r"
)

const (
	ParamVMType        = "vmtype"
	ParamProgramBinary = "programBinary"
)

type contractProgram struct {
	vmtype        string
	programBinary []byte
}

func initialize(ctx vmtypes.Sandbox, params codec.ImmutableCodec) (codec.ImmutableCodec, error) {
	ctx.Publishf("root.initialize.begin")
	state := ctx.AccessState()
	if state.Get(VarStateInitialized) != nil {
		return nil, fmt.Errorf("root.initialize.fail: already_initialized")
	}
	chainID, ok, err := params.GetChainID(VarChainID)
	if err != nil {
		return nil, fmt.Errorf("root.initialize.fail: can't read expected request argument '%s': %s", VarChainID, err.Error())
	}
	if !ok {
		return nil, fmt.Errorf("root.initialize.fail: 'chainID' not found")
	}
	registry := state.GetMap(VarContractRegistry)
	nextIndex := (uint16)(registry.Len())

	if nextIndex != 0 {
		return nil, fmt.Errorf("root.initialize.fail: registry_not_empty")
	}
	state.Set(VarStateInitialized, []byte{0xFF})
	state.SetChainID(VarChainID, chainID)
	// at index 0 always this contract
	registry.SetAt(util.Uint16To2Bytes(nextIndex), util.MustBytes(&contractProgram{
		vmtype:        "builtin",
		programBinary: ProgramHash[:],
	}))
	ctx.Publishf("root.initialize.success")
	return nil, nil
}

func newContract(ctx vmtypes.Sandbox, params codec.ImmutableCodec) (codec.ImmutableCodec, error) {
	ctx.Publishf("root.newContract.begin")

	var err error
	var ok bool
	rec := &contractProgram{}
	rec.vmtype, ok, err = params.GetString(ParamVMType)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("VMType undefined")
	}
	rec.programBinary, err = params.Get(ParamProgramBinary)
	if err != nil {
		return nil, err
	}
	contractIndex, err := ctx.InstallProgram(rec.vmtype, rec.programBinary)
	if err != nil {
		return nil, err
	}
	registry := ctx.AccessState().GetMap(VarContractRegistry)
	registry.SetAt(util.Uint16To2Bytes(contractIndex), util.MustBytes(rec))
	return nil, nil
}

// serde
func (p *contractProgram) Write(w io.Writer) error {
	if err := util.WriteString16(w, p.vmtype); err != nil {
		return err
	}
	if err := util.WriteBytes32(w, p.programBinary); err != nil {
		return err
	}
	return nil
}

func (p *contractProgram) Read(r io.Reader) error {
	var err error
	if p.vmtype, err = util.ReadString16(r); err != nil {
		return err
	}
	if p.programBinary, err = util.ReadBytes32(r); err != nil {
		return err
	}
	return nil
}
