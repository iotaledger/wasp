package root

import (
	"bytes"
	"github.com/iotaledger/wasp/packages/coretypes"
	"io"

	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/contract"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

const (
	Name        = "root"
	Version     = "0.1"
	fullName    = Name + "-" + Version
	description = "Root Contract"
)

var (
	Interface = &contract.ContractInterface{
		Name:        fullName,
		Description: description,
		ProgramHash: *hashing.HashStrings(fullName),
	}
	RootContractRecord = ContractRecord{
		ProgramHash: Interface.ProgramHash,
		Name:        Interface.Name,
		Description: Interface.Description,
	}
)

func init() {
	Interface.WithFunctions(initialize, []contract.ContractFunctionInterface{
		contract.Func(FuncDeployContract, deployContract),
		contract.ViewFunc(FuncFindContract, findContract),
		contract.Func(FuncClaimChainOwnership, claimChainOwnership),
		contract.Func(FuncDelegateChainOwnership, delegateChainOwnership),
		contract.ViewFunc(FuncGetInfo, getInfo),
		contract.ViewFunc(FuncGetFeeInfo, getFeeInfo),
		contract.Func(FuncSetDefaultFee, setDefaultFee),
		contract.Func(FuncSetFee, setFee),
	})
}

// state variables
const (
	VarStateInitialized      = "i"
	VarChainID               = "c"
	VarChainOwnerID          = "o"
	VarFeeColor              = "f"
	VarDefaultFee            = "e"
	VarChainOwnerIDDelegated = "n"
	VarContractRegistry      = "r"
	VarDescription           = "d"
)

// param variables
const (
	ParamChainID     = "$$chainid$$"
	ParamChainOwner  = "$$owner$$"
	ParamProgramHash = "$$proghash$$"
	ParamDescription = "$$description$$"
	ParamHname       = "$$hname$$"
	ParamName        = "$$name$$"
	ParamData        = "$$data$$"
	ParamFeeColor    = "$$feecolor$$"
	ParamDefaultFee  = "$$defaultfee$$"
	ParamContractFee = "$$scFee$$"
)

// function names
const (
	FuncDeployContract         = "deployContract"
	FuncFindContract           = "findContract"
	FuncGetInfo                = "getInfo"
	FuncDelegateChainOwnership = "delegateChainOwnership"
	FuncClaimChainOwnership    = "claimChainOwnership"
	FuncGetFeeInfo             = "getFeeInfo"
	FuncSetDefaultFee          = "setDefaultFee"
	FuncSetFee                 = "setFee"
)

func GetProcessor() vmtypes.Processor {
	return Interface
}

// ContractRecord is a structure which contains metadata for a deployed contract
type ContractRecord struct {
	ProgramHash hashing.HashValue
	Description string
	Name        string
	Fee         int64 // minimum node fee
	Creator     coretypes.AgentID
}

// serde
func (p *ContractRecord) Write(w io.Writer) error {
	if _, err := w.Write(p.ProgramHash[:]); err != nil {
		return err
	}
	if err := util.WriteString16(w, p.Description); err != nil {
		return err
	}
	if err := util.WriteString16(w, p.Name); err != nil {
		return err
	}
	if err := util.WriteInt64(w, p.Fee); err != nil {
		return err
	}
	if _, err := w.Write(p.Creator[:]); err != nil {
		return err
	}
	return nil
}

func (p *ContractRecord) Read(r io.Reader) error {
	var err error
	if err := util.ReadHashValue(r, &p.ProgramHash); err != nil {
		return err
	}
	if p.Description, err = util.ReadString16(r); err != nil {
		return err
	}
	if p.Name, err = util.ReadString16(r); err != nil {
		return err
	}
	if err := util.ReadInt64(r, &p.Fee); err != nil {
		return err
	}
	if err := coretypes.ReadAgentID(r, &p.Creator); err != nil {
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

func NewBuiltinContractRecord(programHash hashing.HashValue, name string, description string) ContractRecord {
	return ContractRecord{
		ProgramHash: programHash,
		Description: description,
		Name:        name,
	}
}
