package root

import (
	"bytes"
	"fmt"
	"io"

	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
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
	EntryPointDeployContract      = coretypes.Hn(FuncDeployContract)
	EntryPointFindContractByIndex = coretypes.Hn(FuncFindContractByIndex)
	EntryPointFindContractByName  = coretypes.Hn(FuncFindContractByName)
	EntryPointGetBinary           = coretypes.Hn(FuncGetBinary)
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
		coretypes.EntryPointCodeInit:  {epFunc(initialize)},
		EntryPointDeployContract:      {epFunc(deployContract)},
		EntryPointFindContractByIndex: {epFuncView(findContractByIndex)},
		EntryPointFindContractByName:  {epFuncView(findContractByName)},
		EntryPointGetBinary:           {epFuncView(getBinary)},
	}
	ProgramHash = hashing.NilHash
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
		return nil, fmt.Errorf("wrong type of entry point")
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
		return nil, fmt.Errorf("wrong type of entry point")
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

func FindContractByIndex(contractIndex uint16, call func(entryPointCode coretypes.Hname, params codec.ImmutableCodec) (codec.ImmutableCodec, error)) (*ContractRecord, bool) {
	if contractIndex == 0 {
		return GetRootContractRecord(), true
	}
	params := codec.NewCodec(dict.New())
	params.SetInt64(ParamIndex, int64(contractIndex))
	res, err := call(EntryPointFindContractByIndex, params)
	if err != nil {
		return nil, false
	}
	data, err := res.Get(ParamData)
	if err != nil {
		return nil, false
	}
	ret, err := DecodeContractRecord(data)
	if err != nil {
		return nil, false
	}
	return ret, true
}

func GetBinary(deploymentHash *hashing.HashValue, call func(entryPointCode coretypes.Hname, params codec.ImmutableCodec) (codec.ImmutableCodec, error)) ([]byte, bool) {
	params := codec.NewCodec(dict.New())
	params.SetHashValue(ParamHash, deploymentHash)
	res, err := call(EntryPointGetBinary, params)
	if err != nil {
		return nil, false
	}
	data, err := res.Get(ParamData)
	if err != nil {
		return nil, false
	}
	return data, true
}
