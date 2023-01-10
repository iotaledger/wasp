package transaction

import "errors"

var (
	ErrNotEnoughBaseTokens                  = errors.New("not enough base tokens")
	ErrNotEnoughBaseTokensForStorageDeposit = errors.New("not enough base tokens for storage deposit")
	ErrNotEnoughNativeTokens                = errors.New("not enough native tokens")
)
