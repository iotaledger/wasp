package transaction

import "golang.org/x/xerrors"

var (
	ErrNotEnoughBaseTokens                  = xerrors.New("not enough base tokens")
	ErrNotEnoughBaseTokensForStorageDeposit = xerrors.New("not enough base tokens for storage deposit")
	ErrNotEnoughNativeTokens                = xerrors.New("not enough native tokens")
)
