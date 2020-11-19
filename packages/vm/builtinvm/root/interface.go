package root

import (
	"bytes"
	"io"

	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

const (
	Name        = "root"
	Version     = "0.1"
	Description = "Root Contract"
)

var (
	RootContractRecord = NewBuiltinContractRecord(Name, Version, Description)

	Interface = ContractInterface{
		Name:        Name,
		Version:     Version,
		Description: Description,
		VMType:      RootContractRecord.VMType,
		Functions: Funcs(initialize, []ContractFunctionInterface{
			Func(FuncDeployContract, deployContract),
			ViewFunc(FuncFindContract, findContract),
			ViewFunc(FuncGetBinary, getBinary),
		}),
	}

	ProgramHash = BuiltinProgramHash(Name, Version)
	Hname       = BuiltinHname(Name, Version)
)

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

func GetProcessor() vmtypes.Processor {
	return &Interface
}

// ContractRecord is a structure which contains metadata for a deployed contract
type ContractRecord struct {
	VMType         string
	DeploymentHash hashing.HashValue // hash(VMType, program binary)
	Description    string
	Name           string
	NodeFee        int64 // minimum node fee
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

func NewBuiltinContractRecord(name string, version string, description string) ContractRecord {
	return ContractRecord{
		VMType:         "builtin",
		DeploymentHash: BuiltinProgramHash(name, version),
		Description:    description,
		Name:           name,
	}
}

func BuiltinProgramHash(name string, version string) hashing.HashValue {
	return *hashing.HashStrings(name + "-" + version)
}

func BuiltinHname(name string, version string) coretypes.Hname {
	return coretypes.Hn(name + "-" + version)
}
