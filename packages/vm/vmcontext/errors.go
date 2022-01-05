package vmcontext

import "golang.org/x/xerrors"

var (
	ErrContractNotFound                   = xerrors.New("contract not found")
	ErrTargetEntryPointNotFound           = xerrors.New("entry point not found")
	ErrEntryPointCantBeAView              = xerrors.New("'init' entry point can't be a view")
	ErrTargetContractNotFound             = xerrors.New("target contract not found")
	ErrTransferTargetAccountDoesNotExists = xerrors.New("transfer target account does not exist")
	ErrRepeatingInitCall                  = xerrors.New("repeating init call")
	ErrNotEnoughTokensFor1GasNominalUnit  = xerrors.New("not enough tokens for one nominal gas unit")
	ErrInconsistentDustAssumptions        = xerrors.New("dust deposit requirements are not consistent with the chain assumptions")
	ErrNotEnoughIotasForDustDeposit       = xerrors.New("not enough iotas for dust deposit")
)
