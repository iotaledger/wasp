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
		coreutil.Func(coreutil.CoreEPRotateCommittee, checkRotateCommitteeRequest),
		coreutil.Func(FuncAddAllowedCommitteeAddress, addAllowedCommitteeAddress),
		coreutil.Func(FuncRemoveAllowedCommitteeAddress, removeAllowedCommitteeAddress),
		coreutil.ViewFunc(FuncIsAllowedCommitteeAddress, isAllowedCommitteeAddress),
		coreutil.Func(FuncMoveToAddress, moveToAddress),
	})
}

const (
	// functions
	FuncMoveToAddress                 = "moveToAddress"
	FuncAddAllowedCommitteeAddress    = "addAllowedCommitteeAddress"
	FuncRemoveAllowedCommitteeAddress = "removeAllowedCommitteeAddress"
	FuncIsAllowedCommitteeAddress     = "isAllowedCommitteeAddress"

	// state variables
	StateVarAllowedCommitteeAddresses = "a"

	// params
	ParamStateAddress     = "s"
	ParamIsAllowedAddress = "i"
)
