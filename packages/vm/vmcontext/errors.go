package vmcontext

import (
	"github.com/iotaledger/wasp/packages/vm/core/errors/commonerrors"
)

const MaxPostedOutputsInOneRequest = 4

var (
	ErrContractNotFound                   = commonerrors.RegisterGlobalError("contract not found").Create()
	ErrTargetEntryPointNotFound           = commonerrors.RegisterGlobalError("entry point not found").Create()
	ErrEntryPointCantBeAView              = commonerrors.RegisterGlobalError("'init' entry point can't be a view").Create()
	ErrTargetContractNotFound             = commonerrors.RegisterGlobalError("target contract not found").Create()
	ErrTransferTargetAccountDoesNotExists = commonerrors.RegisterGlobalError("transfer target account does not exist").Create()
	ErrRepeatingInitCall                  = commonerrors.RegisterGlobalError("repeating init call").Create()
	ErrInconsistentDustAssumptions        = commonerrors.RegisterGlobalError("dust deposit requirements are not consistent with the chain assumptions").Create()
	ErrTooManyEvents                      = commonerrors.RegisterGlobalError("too many events issued for contract").Create()
	ErrTooLargeEvent                      = commonerrors.RegisterGlobalError("event data is too large").Create()
	ErrPrivilegedCallFailed               = commonerrors.RegisterGlobalError("privileged call failed").Create()
	ErrExceededPostedOutputLimit          = commonerrors.RegisterGlobalError("exceeded maximum number of %d posted outputs in one request").Create(MaxPostedOutputsInOneRequest)
	ErrGasBudgetExceeded                  = commonerrors.RegisterGlobalError("gas budget exceeded").Create()
	ErrSenderUnknown                      = commonerrors.RegisterGlobalError("sender unknown").Create()
	ErrGasBudgetDetail                    = commonerrors.RegisterGlobalError("%v: burned (budget) = %d (%d)")
)
