package vmcontext

import "golang.org/x/xerrors"

const MaxPostedOutputsInOneRequest = 4

var (
	ErrContractNotFound                   = xerrors.New("contract not found")
	ErrTargetEntryPointNotFound           = xerrors.New("entry point not found")
	ErrEntryPointCantBeAView              = xerrors.New("'init' entry point can't be a view")
	ErrTargetContractNotFound             = xerrors.New("target contract not found")
	ErrTransferTargetAccountDoesNotExists = xerrors.New("transfer target account does not exist")
	ErrRepeatingInitCall                  = xerrors.New("repeating init call")
	ErrInconsistentDustAssumptions        = xerrors.New("dust deposit requirements are not consistent with the chain assumptions")
	ErrTooManyEvents                      = xerrors.New("too many events issued for contract")
	ErrTooLargeEvent                      = xerrors.New("event data is too large")
	ErrPrivilegedCallFailed               = xerrors.New("privileged call failed")
	ErrExceededPostedOutputLimit          = xerrors.Errorf("exceeded maximum number of %d posted outputs in one request", MaxPostedOutputsInOneRequest)
	ErrGasBudgetExceeded                  = xerrors.New("gas budget exceeded")
)
