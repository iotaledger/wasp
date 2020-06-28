package vmtypes

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/variables"
)

// Sandbox is an interface given to the processor to access the VMContext
// and virtual state, transaction builder and request parameters through it.
type Sandbox interface {
	// general function
	IsOriginState() bool
	GetAddress() address.Address
	GetTimestamp() int64
	Rollback()
	GetLog() *logger.Logger
	Entropy() hashing.HashValue
	// sub interfaces
	// access to the request block
	AccessRequest() RequestAccess
	// base level of virtual state access
	AccessState() StateAccess
	// AccessAccount
	AccessAccount() AccountAccess
	// Send request
	SendRequest(par NewRequestParams) bool
}

// access to request parameters (arguments)
type RequestAccess interface {
	ID() sctransaction.RequestId
	Code() sctransaction.RequestCode
	GetInt64(name string) (int64, bool)
	GetString(name string) (string, bool)
	GetAddressValue(name string) (address.Address, bool)
	GetHashValue(name string) (hashing.HashValue, bool)
	IsAuthorisedByAddress(addr *address.Address) bool
}

// access to the virtual state
type StateAccess interface {
	// getters
	Get(name string) ([]byte, bool)
	GetInt64(name string) (int64, bool, error)
	// setters
	Del(name string)
	Set(name string, value []byte)
	SetInt64(name string, value int64)
	SetString(name string, value string)
	SetAddressValue(name string, addr address.Address)
	SetHashValue(name string, h *hashing.HashValue)
}

// access to token operations (txbuilder)
// mint (create new color) is not here on purpose: ColorNew is used for request tokens
type AccountAccess interface {
	AvailableBalance(col *balance.Color) int64
	MoveTokens(targetAddr *address.Address, col *balance.Color, amount int64) bool
	EraseColor(targetAddr *address.Address, col *balance.Color, amount int64) bool
}

type NewRequestParams struct {
	TargetAddress *address.Address
	RequestCode   sctransaction.RequestCode
	Args          variables.Variables
	IncludeReward int64
}
