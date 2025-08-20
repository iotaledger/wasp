// Package evmerrors defines error types and handling for EVM operations.
package evmerrors

import (
	"bytes"
	"encoding/hex"
	"errors"

	"github.com/ethereum/go-ethereum/accounts/abi"

	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/vm"
)

func IsExecutionReverted(err error) bool {
	if err == nil {
		return false
	}
	return isc.VMErrorIs(err, vm.ErrEVMExecutionReverted)
}

func ExtractRevertData(err error) ([]byte, error) {
	if err == nil {
		return nil, errors.New("expected err != nil")
	}
	if !IsExecutionReverted(err) {
		return nil, nil
	}
	var customError *isc.VMError
	ok := errors.As(err, &customError)
	if !ok {
		return nil, errors.New("could not extract VMError")
	}
	if len(customError.Params()) != 1 {
		return nil, errors.New("expected len(params) == 1")
	}
	revertDataHex, ok := customError.Params()[0].(string)
	if !ok {
		return nil, errors.New("expected params[0] to be string")
	}
	return hex.DecodeString(revertDataHex)
}

func UnpackCustomError(data []byte, abiError abi.Error) ([]any, error) {
	if len(data) < 4 {
		return nil, errors.New("invalid data for unpacking")
	}
	if !bytes.Equal(data[:4], abiError.ID[:4]) {
		return nil, errors.New("invalid error selector")
	}
	return abiError.Inputs.Unpack(data[4:])
}
