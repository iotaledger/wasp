// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmutil

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/iotaledger/wasp/v2/packages/bigint"
)

func Signer(chainID *big.Int) types.Signer {
	return types.NewPragueSigner(chainID)
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

func MustGetSenderIfTxSigned(tx *types.Transaction) common.Address {
	var sender common.Address
	v, r, s := tx.RawSignatureValues()
	if bigint.IsZero(v) && bigint.IsZero(r) && bigint.IsZero(s) {
		return sender // unsigned tx
	}
	return MustGetSender(tx)
}
