package vmconst

import "github.com/iotaledger/wasp/packages/sctransaction"

// built in request codes: the requests processed by any smart contract
// all of them are 'reserved' and 'protected'
const (
	RequestCodeNOP              = sctransaction.RequestCode(0 | sctransaction.RequestCodeProtectedReserved)
	RequestCodeInit             = sctransaction.RequestCode(1 | sctransaction.RequestCodeProtectedReserved)
	RequestCodeSetMinimumReward = sctransaction.RequestCode(2 | sctransaction.RequestCodeProtectedReserved)
)

const (
	VarNameOwnerAddress  = "$owneraddr$"
	VarNameProgramHash   = "$proghash$"
	VarNameDescription   = "$description$"
	VarNameMinimumReward = "$minreward$"
)
