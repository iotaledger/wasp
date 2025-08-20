package vm

import (
	"github.com/iotaledger/wasp/v2/packages/vm/core/errors/coreerrors"
)

var (
	ErrNotEnoughFundsForAllowance = coreerrors.Register("not enough funds for allowance").Create()
	ErrInvalidAllowance           = coreerrors.Register("invalid allowance").Create()
	ErrContractNotFound           = coreerrors.Register("contract with hname %08x not found")
	ErrTargetEntryPointNotFound   = coreerrors.Register("entry point not found").Create()
	ErrTooManyEvents              = coreerrors.Register("too many events issued for contract").Create()
	ErrPrivilegedCallFailed       = coreerrors.Register("privileged call failed").Create()
	ErrGasBudgetExceeded          = coreerrors.Register("gas budget exceeded").Create()
	ErrSenderUnknown              = coreerrors.Register("sender unknown").Create()
	ErrNotEnoughTokensLeftForGas  = coreerrors.Register("not enough funds left to pay for gas").Create()
	ErrUnauthorized               = coreerrors.Register("unauthorized access").Create()
	ErrIllegalCall                = coreerrors.Register("illegal call - entrypoint cannot be called from contracts")
	ErrEVMExecutionReverted       = coreerrors.Register("execution reverted: %s") // hex-encoded revert data
)
