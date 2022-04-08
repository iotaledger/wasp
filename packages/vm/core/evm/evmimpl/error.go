package evmimpl

import (
	"bytes"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/iotaledger/wasp/packages/iscp"
	"golang.org/x/xerrors"
)

type RevertError struct {
	msg    string
	reason string // revert reason hex encoded
}

func (e *RevertError) Error() string {
	if e == nil {
		return ""
	}

	return e.msg
}

// ErrorCode returns the JSON error code for a revert.
// See: https://github.com/ethereum/wiki/wiki/JSON-RPC-Error-Codes-Improvement-Proposal
func (e *RevertError) ErrorCode() int {
	return 3
}

// ErrorData returns the hex encoded revert reason.
func (e *RevertError) ErrorData() interface{} {
	return e.reason
}

var VMErrorCode = crypto.Keccak256([]byte("VMError(uint16)"))[:4]

func UnpackVMError(result *core.ExecutionResult, contractID iscp.Hname) (*iscp.VMErrorCode, error) {
	data := result.Revert()

	if len(data) < 4 {
		return nil, xerrors.New("invalid data for unpacking")
	}

	if !bytes.Equal(data[:4], VMErrorCode) {
		return nil, xerrors.New("invalid data for unpacking")
	}

	abiUint16, _ := abi.NewType("uint16", "", nil)

	errorId, err := (abi.Arguments{{Type: abiUint16}}).Unpack(data[4:])

	if err != nil {
		return nil, err
	}

	errorCode := iscp.NewVMErrorCode(contractID, errorId[0].(uint16))

	return &errorCode, nil
}

func UnpackCommonRevertError(result *core.ExecutionResult) *RevertError {
	reason, _ := abi.UnpackRevert(result.Revert())

	return &RevertError{
		msg:    fmt.Sprintf("execution reverted: %s", reason),
		reason: hexutil.Encode(result.Revert()),
	}
}

func DecodeRevertError(result *core.ExecutionResult, contractID iscp.Hname) error {
	if !result.Failed() {
		return nil
	}

	reason := "(empty reason)"

	if len(result.Revert()) > 0 {
		// First try to decode a VMError
		// Secondly try to decode a normal error as string
		// Otherwise give up

		vmError, _ := UnpackVMError(result, contractID)

		if vmError != nil {
			reason = fmt.Sprintf("contractId: %v, errorId: %v", vmError.ContractID, vmError.ID)
		} else {
			reason, _ = abi.UnpackRevert(result.Revert())
		}
	}

	return &RevertError{
		msg:    fmt.Sprintf("execution reverted: %s", reason),
		reason: hexutil.Encode(result.Revert()),
	}
}

func GetRevertErrorMessage(result *core.ExecutionResult, contractID iscp.Hname) string {
	err := DecodeRevertError(result, contractID)

	if err == nil {
		return ""
	}

	return err.Error()
}
