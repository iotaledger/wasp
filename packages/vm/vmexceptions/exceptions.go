// Package vmexceptions contains exceptions thrown from the vm
package vmexceptions

import (
	"errors"
	"fmt"
)

type skipRequestException struct {
	msg string
}

// skipRequestException is a protocol limit vmexceptions. It causes skipping the request. Never appear in the receipt of the request
var (
	ErrBlockGasLimitExceeded      = &skipRequestException{"exceeded maximum gas allowed in a block"}
	ErrMaxTransactionSizeExceeded = &skipRequestException{"exceeded maximum size of the transaction"}
	ErrNotEnoughFundsForMinFee    = &skipRequestException{"user doesn't have enough on-chain funds to cover the minimum fee for processing this request"}
)

// not a protocol limit error, but something went wrong after request execution
var (
	ErrPostExecutionPanic = fmt.Errorf("post execution error")
)

var SkipRequestErrors = []error{
	ErrNotEnoughFundsForMinFee,
	ErrBlockGasLimitExceeded,
	ErrMaxTransactionSizeExceeded,
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
