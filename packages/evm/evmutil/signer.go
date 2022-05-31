// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmutil

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

func Signer(chainID *big.Int) types.Signer {
	return types.NewEIP155Signer(chainID)
}

func GetSender(tx *types.Transaction) (common.Address, error) {
	return types.Sender(Signer(tx.ChainId()), tx)
}

func MustGetSender(tx *types.Transaction) common.Address {
	sender, err := GetSender(tx)
	if err != nil {
		panic(err)
	}
	return sender
}
