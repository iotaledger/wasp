// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmutil

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/iotaledger/wasp/packages/util"
)

func Signer(chainID *big.Int) types.Signer {
	return types.NewEIP155Signer(chainID)
}

func GetSender(tx *types.Transaction) (common.Address, error) {
	return types.Sender(Signer(tx.ChainId()), tx)
}

func GetSenderIfTxSigned(tx *types.Transaction) (common.Address, error) {
	var sender common.Address
	v, r, s := tx.RawSignatureValues()
	if util.IsZeroBigInt(v) && util.IsZeroBigInt(r) && util.IsZeroBigInt(s) {
		return sender, nil // unsigned tx
	}
	return GetSender(tx)
}
