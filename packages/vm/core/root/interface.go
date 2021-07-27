package root

import (
	"errors"

	"github.com/iotaledger/hive.go/marshalutil"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/colored"
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
)

var (
	Contract            = coreutil.NewContract(coreutil.CoreContractRoot, "Root Contract")
	ErrContractNotFound = errors.New("smart contract not found")
)

// state variables
const (
	VarChainID               = "c"
	VarChainOwnerID          = "o"
	VarChainOwnerIDDelegated = "n"
	VarContractRegistry      = "r"
	VarData                  = "dt"
	VarDefaultOwnerFee       = "do"
	VarDefaultValidatorFee   = "dv"
	VarDeployPermissions     = "dep"
	VarDescription           = "d"
	VarFeeColor              = "f"
	VarOwnerFee              = "of"
	VarStateInitialized      = "i"
	VarValidatorFee          = "vf"
)

// param variables
const (
	ParamChainID      = "$$chainid$$"
	ParamChainOwner   = "$$owner$$"
	ParamDeployer     = "$$deployer$$"
	ParamDescription  = "$$description$$"
	ParamFeeColor     = "$$feecolor$$"
	ParamHname        = "$$hname$$"
	ParamName         = "$$name$$"
	ParamOwnerFee     = "$$ownerfee$$"
	ParamProgramHash  = "$$proghash$$"
	ParamValidatorFee = "$$validatorfee$$"
)

// function names
var (
	FuncClaimChainOwnership    = coreutil.Func("claimChainOwnership")
	FuncDelegateChainOwnership = coreutil.Func("delegateChainOwnership")
	FuncDeployContract         = coreutil.Func("deployContract")
	FuncGrantDeployPermission  = coreutil.Func("grantDeployPermission")
	FuncRevokeDeployPermission = coreutil.Func("revokeDeployPermission")
	FuncSetContractFee         = coreutil.Func("setContractFee")
	FuncSetDefaultFee          = coreutil.Func("setDefaultFee")
	FuncFindContract           = coreutil.ViewFunc("findContract")
	FuncGetChainInfo           = coreutil.ViewFunc("getChainInfo")
	FuncGetFeeInfo             = coreutil.ViewFunc("getFeeInfo")
)

// ContractRecord is a structure which contains metadata of the deployed contract instance
type ContractRecord struct {
	// The ProgramHash uniquely defines the program of the smart contract
	// It is interpreted either as one of builtin contracts (including examples)
	// or a hash (reference) to the of the blob in 'blob' contract in the 'program binary' format,
	// i.e. with at least 2 pre-defined fields:
	//  - VarFieldVType
	//  - VarFieldProgramBinary
	ProgramHash hashing.HashValue
	// Description of the instance
	Description string
	// Unique name of the contract on the chain. The real identity of the instance on the chain
	// is hname(name) =  iscp.Hn(name)
	Name string
	// Chain owner part of the fee. If it is 0, it means chain-global default is in effect
	OwnerFee uint64
	// Validator part of the fee. If it is 0, it means chain-global default is in effect
	ValidatorFee uint64 // validator part of the fee
	// The agentID of the entity which deployed the instance. It can be interpreted as
	// an priviledged user of the instance, however it is up to the smart contract.
	Creator *iscp.AgentID
}

// ChainInfo is an API structure which contains main properties of the chain in on place
type ChainInfo struct {
	ChainID             iscp.ChainID
	ChainOwnerID        iscp.AgentID
	Description         string
	FeeColor            colored.Color
	DefaultOwnerFee     int64
	DefaultValidatorFee int64
}

func NewContractRecord(itf *coreutil.ContractInfo, creator *iscp.AgentID) *ContractRecord {
	// enforce correct creator agentID --  begin
	if creator == nil {
		panic("NewContractRecord: creator can't be nil")
	}
	creator.Bytes() // panics if wrong address
	// enforce correct creator agentID --  end

	return &ContractRecord{
		ProgramHash: itf.ProgramHash,
		Description: itf.Description,
		Name:        itf.Name,
		Creator:     creator,
	}
}

func ContractRecordFromMarshalUtil(mu *marshalutil.MarshalUtil) (*ContractRecord, error) {
	ret := &ContractRecord{}
	if buf, err := mu.ReadBytes(len(ret.ProgramHash)); err != nil {
		return nil, err
	} else {
		copy(ret.ProgramHash[:], buf)
	}
	var err error
	if ret.Description, err = readString(mu); err != nil {
		return nil, err
	}
	if ret.Name, err = readString(mu); err != nil {
		return nil, err
	}
	if ret.OwnerFee, err = mu.ReadUint64(); err != nil {
		return nil, err
	}
	if ret.ValidatorFee, err = mu.ReadUint64(); err != nil {
		return nil, err
	}
	creatorNotNil, err := mu.ReadBool()
	if err != nil {
		return nil, err
	}
	if creatorNotNil {
		if ret.Creator, err = iscp.AgentIDFromMarshalUtil(mu); err != nil {
			return nil, err
		}
	}
	return ret, nil
}

func (p *ContractRecord) Bytes() []byte {
	mu := marshalutil.New()
	mu.WriteBytes(p.ProgramHash[:])
	writeString(mu, p.Description)
	writeString(mu, p.Name)
	mu.WriteUint64(p.OwnerFee)
	mu.WriteUint64(p.ValidatorFee)
	mu.WriteBool(p.Creator != nil)
	if p.Creator != nil {
		mu.Write(p.Creator)
	}
	return mu.Bytes()
}

func (p *ContractRecord) Hname() iscp.Hname {
	if p.Name == "_default" {
		return 0
	}
	return iscp.Hn(p.Name)
}

func ContractRecordFromBytes(data []byte) (*ContractRecord, error) {
	return ContractRecordFromMarshalUtil(marshalutil.New(data))
}

func writeString(mu *marshalutil.MarshalUtil, str string) {
	mu.WriteUint16(uint16(len(str))).WriteBytes([]byte(str))
}

func readString(mu *marshalutil.MarshalUtil) (string, error) {
	sz, err := mu.ReadUint16()
	if err != nil {
		return "", err
	}
	ret, err := mu.ReadBytes(int(sz))
	if err != nil {
		return "", err
	}
	return string(ret), nil
}

func (p *ContractRecord) HasCreator() bool {
	return p.Creator != nil
}
