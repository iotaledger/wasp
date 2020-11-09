package vmconst

import (
	"github.com/iotaledger/wasp/packages/coretypes"
)

// built in request codes: the requests processed by any smart contract
// all of them are 'reserved' and 'protected'
const (
	RequestCodeNOP  = coretypes.Hname(10000)
	RequestCodeInit = coretypes.Hname(10001)
)

const (
	VarNameProgramData   = "$progdata$"
	VarNameOwnerAddress  = "$owneraddr$"
	VarNameDescription   = "$description$"
	VarNameMinimumReward = "$minreward$"
)
