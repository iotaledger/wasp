package evmutil

import (
	"slices"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

func IsFakeTransaction(tx *types.Transaction) bool {
	sender, err := GetSender(tx)

	// the error will fire when the transaction is invalid. This is most of the time a fake evm tx we use for internal calls, therefore it's fine to assume both.
	if slices.Equal(sender.Bytes(), common.Address{}.Bytes()) || err != nil {
		return true
	}

	return false
}
