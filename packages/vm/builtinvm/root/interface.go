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

// state variables
const (
	VarStateInitialized   = "i"
	VarChainID            = "c"
	VarRegistryOfBinaries = "b"
	VarContractRegistry   = "r"
	VarContractsByName    = "n"
	VarDescription        = "d"
)

// param variables
const (
	ParamChainID       = "chainid"
	ParamVMType        = "vmtype"
	ParamProgramBinary = "programBinary"
	ParamDescription   = "description"
	ParamIndex         = "index"
	ParamName          = "name"
	ParamHash          = "hash"
	ParamData          = "data"
)

// function names
const (
	FuncDeployContract      = "deployContract"
	FuncFindContractByIndex = "findContractByIndex"
	FuncFindContractByName  = "findContractRecordByName"
	FuncGetBinary           = "getBinary"
)

// entry point codes
var (
	EntryPointDeployContract      = coretypes.NewEntryPointCodeFromFunctionName(FuncDeployContract)
	EntryPointFindContractByIndex = coretypes.NewEntryPointCodeFromFunctionName(FuncFindContractByIndex)
	EntryPointFindContractByName  = coretypes.NewEntryPointCodeFromFunctionName(FuncFindContractByName)
	EntryPointGetBinary           = coretypes.NewEntryPointCodeFromFunctionName(FuncGetBinary)
)

// ContractRecord is a structure which contains metadata for a deployed contract
type ContractRecord struct {
	VMType         string
	DeploymentHash hashing.HashValue // hash(VMType, program binary)
	Description    string
	Name           string
	NodeFee        int64 // minimum node fee
}

type rootProcessor map[coretypes.EntryPointCode]rootEntryPoint

type rootEntryPoint func(ctx vmtypes.Sandbox) (codec.ImmutableCodec, error)

var (
	processor = rootProcessor{
		coretypes.EntryPointCodeInit:  initialize,
		EntryPointDeployContract:      deployContract,
		EntryPointFindContractByIndex: findContractByIndex,
		EntryPointFindContractByName:  findContractByName,
		EntryPointGetBinary:           getBinary,
	}
	ProgramHash = hashing.NilHash
)

func GetProcessor() vmtypes.Processor {
	return processor
}

func (v rootProcessor) GetEntryPoint(code coretypes.EntryPointCode) (vmtypes.EntryPoint, bool) {
	ret, ok := processor[code]
	return ret, ok
}

func (v rootProcessor) GetDescription() string {
	return "Root processor"
}

func (ep rootEntryPoint) Call(ctx vmtypes.Sandbox) (codec.ImmutableCodec, error) {
	ret, err := ep(ctx)
	if err != nil {
		ctx.Eventf("error occurred: '%v'", err)
	}
	return ret, err
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
