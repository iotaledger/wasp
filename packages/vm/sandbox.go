package vm

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/sctransaction"
)

// Sandbox is an interface given to the processor to access the VMContext.
type Sandbox interface {
	GetAddress() address.Address
	GetTimestamp() int64
	Rollback()
	Request() Request
	State() State
	GetLog() *logger.Logger
}

// access to request parameters
type Request interface {
	ID() sctransaction.RequestId
	Code() sctransaction.RequestCode
	GetInt64(name string) (int64, bool)
	GetString(name string) (string, bool)
}

type State interface {
	Index() uint32
	// getters
	// TODO
	// setters
	SetInt64(name string, value int64)
	SetString(name string, value string)
}
