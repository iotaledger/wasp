package exceptions

import (
	"fmt"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

type ProtocolLimitException struct {
	msg string
}

// protocol limit exceptions causes skipping the request. Never appear in the receipt of the request
var (
	ErrInputLimitExceeded                   = &ProtocolLimitException{fmt.Sprintf("exceeded maximum number of inputs in transaction. iotago.MaxInputsCount = %d", iotago.MaxInputsCount)}
	ErrOutputLimitExceeded                  = &ProtocolLimitException{fmt.Sprintf("exceeded maximum number of outputs in transaction. iotago.MaxOutputsCount = %d", iotago.MaxOutputsCount)}
	ErrNumberOfNativeTokensLimitExceeded    = &ProtocolLimitException{fmt.Sprintf("exceeded maximum number of different native tokens in transaction. iotago.MaxNativeTokensCount = %d", iotago.MaxNativeTokensCount)}
	ErrNotEnoughFundsForInternalDustDeposit = &ProtocolLimitException{"not enough funds for internal dust deposit: common account must be topped up"}
	ErrBlockGasLimitExceeded                = &ProtocolLimitException{fmt.Sprintf("exceeded maximum gas allowed in a block. MaxGasPerBlock = %d", gas.MaxGasPerBlock)}
)

var All = []error{
	ErrInputLimitExceeded,
	ErrOutputLimitExceeded,
	ErrNumberOfNativeTokensLimitExceeded,
	ErrNotEnoughFundsForInternalDustDeposit,
	ErrBlockGasLimitExceeded,
}

func (m *ProtocolLimitException) Error() string {
	return m.msg
}

func IsProtocolLimitException(e interface{}) bool {
	_, ok := e.(*ProtocolLimitException)
	return ok
}
