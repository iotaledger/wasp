// in the blocklog core contract the VM keeps indices of blocks and requests in an optimized way
// for fast checking and timestamp access.
package governance

import (
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
)

// constants
const (
	MinEventSize               = uint16(200)
	MinEventsPerRequest        = uint16(10)
	DefaultMaxEventsPerRequest = uint16(50)
	DefaultMaxEventSize        = uint16(2000)    // 2Kb
	DefaultMaxBlobSize         = uint32(1000000) // 1Mb
)

var Contract = coreutil.NewContract(coreutil.CoreContractGovernance, "Governance contract")

var (
	// state controller (entity that owns the state output via AliasAddress)
	FuncRotateStateController               = coreutil.Func(coreutil.CoreEPRotateStateController)
	FuncAddAllowedStateControllerAddress    = coreutil.Func("addAllowedStateControllerAddress")
	FuncRemoveAllowedStateControllerAddress = coreutil.Func("removeAllowedStateControllerAddress")
	FuncGetAllowedStateControllerAddresses  = coreutil.ViewFunc("getAllowedStateControllerAddresses")

	// chain owner (L1 entity that is the "owner of the chain")
	FuncClaimChainOwnership    = coreutil.Func("claimChainOwnership")
	FuncDelegateChainOwnership = coreutil.Func("delegateChainOwnership")
	FuncGetChainOwner          = coreutil.ViewFunc("getChainOwner")

	// fees
	FuncSetContractFee = coreutil.Func("setContractFee")
	FuncGetFeeInfo     = coreutil.ViewFunc("getFeeInfo")

	// chain info
	FuncSetChainInfo   = coreutil.Func("setChainInfo")
	FuncGetChainInfo   = coreutil.ViewFunc("getChainInfo")
	FuncGetMaxBlobSize = coreutil.ViewFunc("getMaxBlobSize")
)

// state variables
const (
	// state controller
	StateVarAllowedStateControllerAddresses = "a"
	StateVarRotateToAddress                 = "r"

	// chain owner
	VarChainOwnerID          = "o"
	VarChainOwnerIDDelegated = "n"
	VarDefaultOwnerFee       = "do"
	VarOwnerFee              = "of"

	// fees
	VarDefaultValidatorFee  = "dv"
	VarValidatorFee         = "vf"
	VarFeeColor             = "f"
	VarContractFeesRegistry = "fr"

	// chain info
	VarChainID         = "c"
	VarDescription     = "d"
	VarMaxBlobSize     = "mb"
	VarMaxEventSize    = "me"
	VarMaxEventsPerReq = "mr"
)

// params
const (
	// state controller
	ParamStateControllerAddress          = coreutil.ParamStateControllerAddress
	ParamAllowedStateControllerAddresses = "a"

	// chain owner
	ParamChainOwner = "oi"
	ParamOwnerFee   = "of"

	// fees
	ParamFeeColor     = "fc"
	ParamValidatorFee = "vf"
	ParamHname        = "hn"

	// chain info
	ParamChainID             = "ci"
	ParamDescription         = "ds"
	ParamMaxBlobSize         = "bs"
	ParamMaxEventSize        = "es"
	ParamMaxEventsPerRequest = "ne"
)
