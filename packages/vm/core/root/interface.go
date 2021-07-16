package root

import (
	"bytes"
	"errors"
	"io"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/packages/util"
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

// TODO move ownership and fee-related methods to the governance contract

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
	FeeColor            ledgerstate.Color
	DefaultOwnerFee     int64
	DefaultValidatorFee int64
}

func (p *ContractRecord) Hname() iscp.Hname {
	if p.Name == "_default" {
		return 0
	}
	return iscp.Hn(p.Name)
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
		p.Creator = &iscp.AgentID{}
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

func NewContractRecord(itf *coreutil.ContractInfo, creator *iscp.AgentID) *ContractRecord {
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
