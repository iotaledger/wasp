package transaction

import "golang.org/x/xerrors"

var (
	ErrNotEnoughBaseTokens               = xerrors.New("not enough base tokens")
	ErrNotEnoughBaseTokensForDustDeposit = xerrors.New("not enough base tokens for dust deposit")
	ErrNotEnoughNativeTokens             = xerrors.New("not enough native tokens")
)
