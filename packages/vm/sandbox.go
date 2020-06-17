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
	GetIntRequest(name string) (int, bool)
	GetStringRequest(name string) (string, bool)
	// state Getters
	// state setters
	SetInt(name string, value int)
	SetString(name string, value string)
}
