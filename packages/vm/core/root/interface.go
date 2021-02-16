package root

import (
	"bytes"
	"errors"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/coretypes"
	"io"

	"github.com/iotaledger/wasp/packages/coretypes/coreutil"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/util"
)

const (
	Name        = "root"
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
	VarChainColor            = "co"
	VarChainAddress          = "ad"
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
	ParamChainColor   = "$$color$$"
	ParamChainAddress = "$$address$$"
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
	OwnerFee int64
	// Validator part of the fee. If it is 0, it means chain-global default is in effect
	ValidatorFee int64 // validator part of the fee
	// The agentID of the entity which deployed the instance. It can be interpreted as
	// an priviledged user of the instance, however it is up to the smart contract.
	Creator coretypes.AgentID
}

// ChainInfo is an API structure which contains main properties of the chain in on place
type ChainInfo struct {
	ChainID             coretypes.ChainID
	ChainOwnerID        coretypes.AgentID
	ChainColor          balance.Color
	ChainAddress        address.Address
	Description         string
	FeeColor            balance.Color
	DefaultOwnerFee     int64
	DefaultValidatorFee int64
}

func (p *ContractRecord) Hname() coretypes.Hname {
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
	if err := util.WriteInt64(w, p.OwnerFee); err != nil {
		return err
	}
	if err := util.WriteInt64(w, p.ValidatorFee); err != nil {
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
	if err := util.ReadInt64(r, &p.OwnerFee); err != nil {
		return err
	}
	if err := util.ReadInt64(r, &p.ValidatorFee); err != nil {
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

func NewContractRecord(itf *coreutil.ContractInterface, creator coretypes.AgentID) (ret ContractRecord) {
	ret = ContractRecord{
		ProgramHash: itf.ProgramHash,
		Description: itf.Description,
		Name:        itf.Name,
		Creator:     creator,
	}
	return
}

func (p *ContractRecord) HasCreator() bool {
	return p.Creator != coretypes.AgentID{}
}
