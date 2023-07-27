package vmexceptions

import (
	"errors"
	"fmt"

	iotago "github.com/iotaledger/iota.go/v3"
)

type skipRequestException struct {
	msg string
}

// skipRequestException is a protocol limit vmexceptions. It causes skipping the request. Never appear in the receipt of the request
var (
	ErrInputLimitExceeded                      = &skipRequestException{fmt.Sprintf("exceeded maximum number of inputs in transaction. iotago.MaxInputsCount = %d", iotago.MaxInputsCount)}
	ErrOutputLimitExceeded                     = &skipRequestException{fmt.Sprintf("exceeded maximum number of outputs in transaction. iotago.MaxOutputsCount = %d", iotago.MaxOutputsCount)}
	ErrTotalNativeTokensLimitExceeded          = &skipRequestException{fmt.Sprintf("exceeded maximum number of different native tokens in transaction. iotago.MaxNativeTokensCount = %d", iotago.MaxNativeTokensCount)}
	ErrNotEnoughFundsForInternalStorageDeposit = &skipRequestException{"not enough funds for internal storage deposit: common account must be topped up"}
	ErrBlockGasLimitExceeded                   = &skipRequestException{"exceeded maximum gas allowed in a block"}
	ErrMaxTransactionSizeExceeded              = &skipRequestException{"exceeded maximum size of the transaction"}
	ErrNotEnoughFundsForSD                     = &skipRequestException{"user doesn't have enough on-chain funds to cover the SD cost of processing this request"}
)

// not a protocol limit error, but something went wrong after request execution
var (
	ErrPostExecutionPanic = fmt.Errorf("post execution error")
)

var SkipRequestErrors = []error{
	ErrInputLimitExceeded,
	ErrOutputLimitExceeded,
	ErrTotalNativeTokensLimitExceeded,
	ErrNotEnoughFundsForInternalStorageDeposit,
	ErrBlockGasLimitExceeded,
	ErrMaxTransactionSizeExceeded,
	ErrNotEnoughFundsForSD,
	ErrPostExecutionPanic,
}

func (m *skipRequestException) Error() string {
	return m.msg
}

func IsSkipRequestException(e interface{}) error {
	s, ok := e.(*skipRequestException)
	if !ok {
		return nil
	}
	return errors.New(s.msg)
}
