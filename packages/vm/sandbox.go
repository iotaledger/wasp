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
	GetStateIndex() uint32
	GetRequestID() sctransaction.RequestId
	GetRequestCode() sctransaction.RequestCode
	GetLog() *logger.Logger
	// request Getters
	GetInt64RequestParam(name string) (int64, bool)
	GetStringRequestParam(name string) (string, bool)
	// state Getters
	// state setters
	SetInt(name string, value int)
	SetString(name string, value string)
}
