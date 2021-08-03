package root

import (
	"errors"

	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/colored"
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
)

var (
	Contract            = coreutil.NewContract(coreutil.CoreContractRoot, "Root Contract")
	ErrContractNotFound = errors.New("smart contract not found")
)

// constants
const (
	MinEventSize               = uint16(200)
	MinEventsPerRequest        = uint16(10)
	DefaultMaxEventsPerRequest = uint16(50)
	DefaultMaxEventSize        = uint16(2000)    // 2Kb
	DefaultMaxBlobSize         = uint32(1000000) // 1Mb
)

// state variables
const (
	VarChainID               = "c"
	VarChainOwnerID          = "o"
	VarChainOwnerIDDelegated = "n"
	VarContractRegistry      = "r"
	VarDefaultOwnerFee       = "do"
	VarDefaultValidatorFee   = "dv"
	VarDeployPermissions     = "dep"
	VarDescription           = "d"
	VarFeeColor              = "f"
	VarOwnerFee              = "of"
	VarStateInitialized      = "i"
	VarValidatorFee          = "vf"
	VarMaxBlobSize           = "mb"
	VarMaxEventSize          = "me"
	VarMaxEventsPerReq       = "mr"
)

// param variables
const (
	ParamChainID             = "ci"
	ParamChainOwner          = "oi"
	ParamDeployer            = "dp"
	ParamDescription         = "ds"
	ParamFeeColor            = "fc"
	ParamHname               = "hn"
	ParamName                = "nm"
	ParamOwnerFee            = "of"
	ParamProgramHash         = "ph"
	ParamValidatorFee        = "vf"
	ParamContractRecData     = "dt"
	ParamContractFound       = "cf"
	ParamMaxBlobSize         = "bs"
	ParamMaxEventSize        = "es"
	ParamMaxEventsPerRequest = "ne"
)

// function names
var (
	FuncClaimChainOwnership    = coreutil.Func("claimChainOwnership")
	FuncDelegateChainOwnership = coreutil.Func("delegateChainOwnership")
	FuncDeployContract         = coreutil.Func("deployContract")
	FuncGrantDeployPermission  = coreutil.Func("grantDeployPermission")
	FuncRevokeDeployPermission = coreutil.Func("revokeDeployPermission")
	FuncSetContractFee         = coreutil.Func("setContractFee")
	FuncSetChainInfo           = coreutil.Func("setChainInfo")
	FuncFindContract           = coreutil.ViewFunc("findContract")
	FuncGetChainInfo           = coreutil.ViewFunc("getChainInfo")
	FuncGetFeeInfo             = coreutil.ViewFunc("getFeeInfo")
	FuncGetMaxBlobSize         = coreutil.ViewFunc("getMaxBlobStyle")
)

// ChainInfo is an API structure which contains main properties of the chain in on place
type ChainInfo struct {
	ChainID             iscp.ChainID
	ChainOwnerID        iscp.AgentID
	Description         string
	FeeColor            colored.Color
	DefaultOwnerFee     int64
	DefaultValidatorFee int64
	MaxBlobSize         uint32
	MaxEventSize        uint16
	MaxEventsPerReq     uint16
}
