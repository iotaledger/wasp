package vmconst

import "github.com/iotaledger/wasp/packages/sctransaction"

// built in request codes: the requests processed by any smart contract
// all of them are 'reserved' and 'protected'
const (
	RequestCodeNOP              = sctransaction.RequestCode(uint16(0) | sctransaction.RequestCodeProtectedReserved)
	RequestCodeInit             = sctransaction.RequestCode(uint16(1) | sctransaction.RequestCodeProtectedReserved)
	RequestCodeSetMinimumReward = sctransaction.RequestCode(uint16(2) | sctransaction.RequestCodeProtectedReserved)
	RequestCodeSetDescription   = sctransaction.RequestCode(uint16(3) | sctransaction.RequestCodeProtectedReserved)
)

const (
	VarNameOwnerAddress  = "$owneraddr$"
	VarNameProgramHash   = "$proghash$"
	VarNameMinimumReward = "$minreward$"
)
