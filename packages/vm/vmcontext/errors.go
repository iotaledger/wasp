package vmcontext

import (
	"github.com/iotaledger/wasp/packages/vm/core/errors/commonerrors"
)

const MaxPostedOutputsInOneRequest = 4

var (
	ErrContractNotFound                   = commonerrors.RegisterGlobalError("contract not found").CreateTyped()
	ErrTargetEntryPointNotFound           = commonerrors.RegisterGlobalError("entry point not found").CreateTyped()
	ErrEntryPointCantBeAView              = commonerrors.RegisterGlobalError("'init' entry point can't be a view").CreateTyped()
	ErrTargetContractNotFound             = commonerrors.RegisterGlobalError("target contract not found").CreateTyped()
	ErrTransferTargetAccountDoesNotExists = commonerrors.RegisterGlobalError("transfer target account does not exist").CreateTyped()
	ErrRepeatingInitCall                  = commonerrors.RegisterGlobalError("repeating init call").CreateTyped()
	ErrInconsistentDustAssumptions        = commonerrors.RegisterGlobalError("dust deposit requirements are not consistent with the chain assumptions").CreateTyped()
	ErrTooManyEvents                      = commonerrors.RegisterGlobalError("too many events issued for contract").CreateTyped()
	ErrTooLargeEvent                      = commonerrors.RegisterGlobalError("event data is too large").CreateTyped()
	ErrPrivilegedCallFailed               = commonerrors.RegisterGlobalError("privileged call failed").CreateTyped()
	ErrExceededPostedOutputLimit          = commonerrors.RegisterGlobalError("exceeded maximum number of %d posted outputs in one request").CreateTyped(MaxPostedOutputsInOneRequest)
	ErrGasBudgetExceeded                  = commonerrors.RegisterGlobalError("gas budget exceeded").CreateTyped()
	ErrSenderUnknown                      = commonerrors.RegisterGlobalError("sender unknown").CreateTyped()
	ErrGasBudgetDetail                    = commonerrors.RegisterGlobalError("%v: burned (budget) = %d (%d)")
)
