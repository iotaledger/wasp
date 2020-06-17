package builtin

import "github.com/iotaledger/wasp/packages/sctransaction"

// built in request codes: the requests processed by any smart contract
// all of them are 'reserved' and protected
const (
	RequestCodeNOP              = 0 + sctransaction.FirstBuiltInRequestCode
	RequestCodeSetMinimumReward = 1 + sctransaction.FirstBuiltInRequestCode
	RequestCodeSetDescription   = 2 + sctransaction.FirstBuiltInRequestCode
)
