package vmtypes

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/table"
)

// Sandbox is an interface given to the processor to access the VMContext
// and virtual state, transaction builder and request parameters through it.
type Sandbox interface {
	// general function
	IsOriginState() bool
	GetOwnAddress() *address.Address
	GetTimestamp() int64
	GetEntropy() hashing.HashValue // 32 bytes of deterministic and unpredictably random data
	GetLog() *logger.Logger
	Rollback()
	// sub interfaces
	// access to the request block
	AccessRequest() RequestAccess
	// base level of virtual state access
	AccessState() StateAccess
	// AccessOwnAccount
	AccessOwnAccount() AccountAccess
	// Send request
	SendRequest(par NewRequestParams) bool
	// Send request to itself
	SendRequestToSelf(reqCode sctransaction.RequestCode, args table.MemTable) bool
	// Publish "vmmsg" message through Publisher
	Publish(msg string)
}

// access to request parameters (arguments)
type RequestAccess interface {
	ID() sctransaction.RequestId
	Code() sctransaction.RequestCode
	IsAuthorisedByAddress(addr *address.Address) bool
	Senders() []address.Address
	Args() table.RCodec
}

// access to the virtual state
type StateAccess interface {
	Variables() table.Codec
}

// access to token operations (txbuilder)
// mint (create new color) is not here on purpose: ColorNew is used for request tokens
type AccountAccess interface {
	// access to total available outputs/balances
	AvailableBalance(col *balance.Color) int64
	MoveTokens(targetAddr *address.Address, col *balance.Color, amount int64) bool
	EraseColor(targetAddr *address.Address, col *balance.Color, amount int64) bool
	// part of the outputs/balances which are coming from the current request transaction
	AvailableBalanceFromRequest(col *balance.Color) int64
	MoveTokensFromRequest(targetAddr *address.Address, col *balance.Color, amount int64) bool
	EraseColorFromRequest(targetAddr *address.Address, col *balance.Color, amount int64) bool
}

type NewRequestParams struct {
	TargetAddress *address.Address
	RequestCode   sctransaction.RequestCode
	Args          table.MemTable
	IncludeReward int64
}
