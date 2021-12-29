package vmcontext

import "golang.org/x/xerrors"

var (
	ErrTargetContractNotFound            = xerrors.New("target contract not found")
	ErrRepeatingInitCall                 = xerrors.New("repeating init call")
	ErrNonViewExpected                   = xerrors.New("non-view entry point expected")
	ErrNotEnoughTokensFor1GasNominalUnit = xerrors.New("not enough tokens for one nominal gas unit")
)
