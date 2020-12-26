package root

import (
	"bytes"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/coretypes"
	"io"

	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/contract"
)

const (
	Name        = "root"
	Version     = "0.1"
	description = "Root Contract"
)

var (
	Interface = &contract.ContractInterface{
		Name:        Name,
		Description: description,
		ProgramHash: *hashing.HashStrings(Name),
	}
)

func init() {
	Interface.WithFunctions(initialize, []contract.ContractFunctionInterface{
		contract.Func(FuncDeployContract, deployContract),
		contract.ViewFunc(FuncFindContract, findContract),
		contract.Func(FuncClaimChainOwnership, claimChainOwnership),
		contract.Func(FuncDelegateChainOwnership, delegateChainOwnership),
		contract.ViewFunc(FuncGetChainInfo, getChainInfo),
		contract.ViewFunc(FuncGetFeeInfo, getFeeInfo),
		contract.Func(FuncSetDefaultFee, setDefaultFee),
		contract.Func(FuncSetContractFee, setContractFee),
		contract.Func(FuncGrantDeploy, grantDeployPermission),
		contract.Func(FuncRevokeDeploy, revokeDeployPermission),
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
	VarDeployAuthorisations  = "dep"
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

// ContractRecord is a structure which contains metadata for a deployed contract
type ContractRecord struct {
	ProgramHash  hashing.HashValue
	Description  string
	Name         string
	OwnerFee     int64 // owner part of the fee
	ValidatorFee int64 // validator part of the fee
	Creator      coretypes.AgentID
}

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

func NewContractRecord(itf *contract.ContractInterface, creator coretypes.AgentID) (ret ContractRecord) {
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
