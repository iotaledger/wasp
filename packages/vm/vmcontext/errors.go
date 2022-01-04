package vmcontext

import "golang.org/x/xerrors"

var (
	ErrTargetContractNotFound            = xerrors.New("target contract not found")
	ErrRepeatingInitCall                 = xerrors.New("repeating init call")
	ErrNonViewExpected                   = xerrors.New("non-view entry point expected")
	ErrNotEnoughTokensFor1GasNominalUnit = xerrors.New("not enough tokens for one nominal gas unit")
	ErrInconsistentDustAssumptions       = xerrors.New("dust deposit requirements are not consistent with the chain assumptions")
	ErrWrongParamsInSandboxCall          = xerrors.New("wrong parameter in a sandbox call")
)
