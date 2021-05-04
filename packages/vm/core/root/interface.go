package root

import (
	"bytes"
	"errors"
	"io"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/coreutil"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/util"
)

const (
	Name        = coreutil.CoreContractRoot
	description = "Root Contract"
)

var (
	Interface = &coreutil.ContractInterface{
		Name:        Name,
		Description: description,
		ProgramHash: hashing.HashStrings(Name),
	}
	ErrContractNotFound = errors.New("smart contract not found")
)

func init() {
	Interface.WithFunctions(initialize, []coreutil.ContractFunctionInterface{
		coreutil.Func(FuncDeployContract, deployContract),
		coreutil.ViewFunc(FuncFindContract, findContract),
		coreutil.Func(FuncClaimChainOwnership, claimChainOwnership),
		coreutil.Func(FuncDelegateChainOwnership, delegateChainOwnership),
		coreutil.ViewFunc(FuncGetChainInfo, getChainInfo),
		coreutil.ViewFunc(FuncGetFeeInfo, getFeeInfo),
		coreutil.Func(FuncSetDefaultFee, setDefaultFee),
		coreutil.Func(FuncSetContractFee, setContractFee),
		coreutil.Func(FuncGrantDeploy, grantDeployPermission),
		coreutil.Func(FuncRevokeDeploy, revokeDeployPermission),
	})
}

// state variables
const (
	VarStateInitialized      = "i"
	VarChainID               = "c"
	VarChainOwnerID          = "o"
	VarFeeColor              = "f"
	VarDefaultOwnerFee       = "do"
	VarDefaultValidatorFee   = "dv"
	VarChainOwnerIDDelegated = "n"
	VarContractRegistry      = "r"
	VarDescription           = "d"
	VarDeployPermissions     = "dep"
)

// param variables
const (
	ParamChainID      = "$$chainid$$"
	ParamChainOwner   = "$$owner$$"
	ParamProgramHash  = "$$proghash$$"
	ParamDescription  = "$$description$$"
	ParamHname        = "$$hname$$"
	ParamName         = "$$name$$"
	ParamData         = "$$data$$"
	ParamFeeColor     = "$$feecolor$$"
	ParamOwnerFee     = "$$ownerfee$$"
	ParamValidatorFee = "$$validatorfee$$"
	ParamDeployer     = "$$deployer$$"
)

// function names
const (
	FuncDeployContract         = "deployContract"
	FuncFindContract           = "findContract"
	FuncGetChainInfo           = "getChainInfo"
	FuncDelegateChainOwnership = "delegateChainOwnership"
	FuncClaimChainOwnership    = "claimChainOwnership"
	FuncGetFeeInfo             = "getFeeInfo"
	FuncSetDefaultFee          = "setDefaultFee"
	FuncSetContractFee         = "setContractFee"
	FuncGrantDeploy            = "grantDeployPermission"
	FuncRevokeDeploy           = "revokeDeployPermission"
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
	// is hname(name) =  coretypes.Hn(name)
	Name string
	// Chain owner part of the fee. If it is 0, it means chain-global default is in effect
	OwnerFee uint64
	// Validator part of the fee. If it is 0, it means chain-global default is in effect
	ValidatorFee uint64 // validator part of the fee
	// The agentID of the entity which deployed the instance. It can be interpreted as
	// an priviledged user of the instance, however it is up to the smart contract.
	Creator *coretypes.AgentID
}

// ChainInfo is an API structure which contains main properties of the chain in on place
type ChainInfo struct {
	ChainID             coretypes.ChainID
	ChainOwnerID        coretypes.AgentID
	Description         string
	FeeColor            ledgerstate.Color
	DefaultOwnerFee     int64
	DefaultValidatorFee int64
}

func (p *ContractRecord) Hname() coretypes.Hname {
	if p.Name == "_default" {
		return 0
	}
	return coretypes.Hn(p.Name)
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
	if err := util.WriteUint64(w, p.OwnerFee); err != nil {
		return err
	}
	if err := util.WriteUint64(w, p.ValidatorFee); err != nil {
		return err
	}
	if err := util.WriteBoolByte(w, p.Creator != nil); err != nil {
		return err
	}
	if p.Creator != nil {
		if err := p.Creator.Write(w); err != nil {
			return err
		}
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
	if err := util.ReadUint64(r, &p.OwnerFee); err != nil {
		return err
	}
	if err := util.ReadUint64(r, &p.ValidatorFee); err != nil {
		return err
	}
	var hasCreator bool
	if err := util.ReadBoolByte(r, &hasCreator); err != nil {
		return err
	}
	if hasCreator {
		p.Creator = &coretypes.AgentID{}
		if err := p.Creator.Read(r); err != nil {
			return err
		}
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

func NewContractRecord(itf *coreutil.ContractInterface, creator *coretypes.AgentID) *ContractRecord {
	return &ContractRecord{
		ProgramHash: itf.ProgramHash,
		Description: itf.Description,
		Name:        itf.Name,
		Creator:     creator,
	}
}

func (p *ContractRecord) HasCreator() bool {
	return p.Creator != nil
}
