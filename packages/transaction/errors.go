package transaction

import "fmt"

var (
	ErrNotEnoughBaseTokens                  = fmt.Errorf("not enough base tokens")
	ErrNotEnoughBaseTokensForStorageDeposit = fmt.Errorf("not enough base tokens for storage deposit")
	ErrNotEnoughNativeTokens                = fmt.Errorf("not enough native tokens")
)
