// in the blocklog core contract the VM keeps indices of blocks and requests in an optimized way
// for fast checking and timestamp access.
package governance

import (
	"github.com/iotaledger/wasp/packages/coretypes/coreutil"
	"github.com/iotaledger/wasp/packages/hashing"
)

const (
	Name        = coreutil.CoreContractGovernance
	description = "Governance contract"
)

var (
	Interface = &coreutil.ContractInterface{
		Name:        Name,
		Description: description,
		ProgramHash: hashing.HashStrings(Name),
	}
)

func init() {
	Interface.WithFunctions(initialize, []coreutil.ContractFunctionInterface{
		coreutil.Func(coreutil.CoreEPRotateStateController, rotateStateController),
		coreutil.Func(FuncAddAllowedStateControllerAddress, addAllowedStateControllerAddress),
		coreutil.Func(FuncRemoveAllowedStateControllerAddress, removeAllowedStateControllerAddress),
		coreutil.ViewFunc(FuncGetAllowedStateControllerAddresses, getAllowedStateControllerAddresses),
	})
}

const (
	// functions
	FuncAddAllowedStateControllerAddress    = "addAllowedStateControllerAddress"
	FuncRemoveAllowedStateControllerAddress = "removeAllowedStateControllerAddress"
	FuncGetAllowedStateControllerAddresses  = "getAllowedStateControllerAddresses"

	// state variables
	StateVarAllowedStateControllerAddresses = "a"
	StateVarRotateToAddress                 = "r"

	// params
	ParamStateControllerAddress          = coreutil.ParamStateControllerAddress
	ParamAllowedStateControllerAddresses = "a"
)
