package root

import (
	"bytes"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
	"io"
)

// Version of the root contract
const Version = "0.1"

// state variables
const (
	VarStateInitialized   = "i"
	VarChainID            = "c"
	VarChainOwnerID       = "o"
	VarRegistryOfBinaries = "b"
	VarContractRegistry   = "r"
	VarDescription        = "d"
)

// param variables
const (
	ParamChainID       = "chainid"
	ParamVMType        = "vmtype"
	ParamProgramBinary = "programBinary"
	ParamDescription   = "description"
	ParamHname         = "hname"
	ParamName          = "name"
	ParamHash          = "hash"
	ParamData          = "data"
)

// function names
const (
	FuncDeployContract = "deployContract"
	FuncFindContract   = "findContract"
	FuncGetBinary      = "getBinary"
)

const ContractName = "root " + Version

var (
	ProgramHash              = hashing.HashStrings(ContractName)
	Hname                    = coretypes.Hn(ContractName)
	EntryPointDeployContract = coretypes.Hn(FuncDeployContract)
	EntryPointFindContract   = coretypes.Hn(FuncFindContract)
	EntryPointGetBinary      = coretypes.Hn(FuncGetBinary)
)

// ContractRecord is a structure which contains metadata for a deployed contract
type ContractRecord struct {
	VMType         string
	DeploymentHash hashing.HashValue // hash(VMType, program binary)
	Description    string
	Name           string
	NodeFee        int64 // minimum node fee
}

type rootProcessor map[coretypes.Hname]rootEntryPoint

type epFunc func(ctx vmtypes.Sandbox) (codec.ImmutableCodec, error)
type epFuncView func(ctx vmtypes.SandboxView) (codec.ImmutableCodec, error)

type rootEntryPoint struct {
	fun interface{}
}

var (
	processor = rootProcessor{
		coretypes.EntryPointInit: {epFunc(initialize)},
		EntryPointDeployContract: {epFunc(deployContract)},
		EntryPointFindContract:   {epFuncView(findContract)},
		EntryPointGetBinary:      {epFuncView(getBinary)},
	}
)

func GetProcessor() vmtypes.Processor {
	return processor
}

func (v rootProcessor) GetEntryPoint(code coretypes.Hname) (vmtypes.EntryPoint, bool) {
	ret, ok := processor[code]
	return ret, ok
}

func (v rootProcessor) GetDescription() string {
	return "Root processor"
}

func (ep rootEntryPoint) Call(ctx vmtypes.Sandbox) (codec.ImmutableCodec, error) {
	fun, ok := ep.fun.(epFunc)
	if !ok {
		return nil, vmtypes.ErrWrongTypeEntryPoint
	}
	ret, err := fun(ctx)
	if err != nil {
		ctx.Eventf("error occurred: '%v'", err)
	}
	return ret, err
}

func (ep rootEntryPoint) IsView() bool {
	_, ok := ep.fun.(epFuncView)
	return ok
}

func (ep rootEntryPoint) CallView(ctx vmtypes.SandboxView) (codec.ImmutableCodec, error) {
	fun, ok := ep.fun.(epFuncView)
	if !ok {
		return nil, vmtypes.ErrWrongTypeEntryPoint
	}
	return fun(ctx)
}

func (ep rootEntryPoint) WithGasLimit(_ int) vmtypes.EntryPoint {
	return ep
}

// serde
func (p *ContractRecord) Write(w io.Writer) error {
	if err := util.WriteString16(w, p.VMType); err != nil {
		return err
	}
	if _, err := w.Write(p.DeploymentHash[:]); err != nil {
		return err
	}
	if err := util.WriteString16(w, p.Description); err != nil {
		return err
	}
	if err := util.WriteString16(w, p.Name); err != nil {
		return err
	}
	if err := util.WriteInt64(w, p.NodeFee); err != nil {
		return err
	}
	return nil
}

func (p *ContractRecord) Read(r io.Reader) error {
	var err error
	if p.VMType, err = util.ReadString16(r); err != nil {
		return err
	}
	if err := util.ReadHashValue(r, &p.DeploymentHash); err != nil {
		return err
	}
	if p.Description, err = util.ReadString16(r); err != nil {
		return err
	}
	if p.Name, err = util.ReadString16(r); err != nil {
		return err
	}
	if err := util.ReadInt64(r, &p.NodeFee); err != nil {
		return err
	}
	return nil
}

func EncodeContractRecord(p *ContractRecord) []byte {
	return util.MustBytes(p)
}

func DecodeContractRecord(data []byte) (*ContractRecord, error) {
	ret := new(ContractRecord)
	err := ret.Read(bytes.NewReader(data))
	return ret, err
}

// GetRootContractRecord creates predefined metadata for the root contract
func GetRootContractRecord() *ContractRecord {
	return &ContractRecord{
		VMType:         "builtin",
		DeploymentHash: *ProgramHash,
		Description:    "root contract",
		Name:           "root",
	}
}
