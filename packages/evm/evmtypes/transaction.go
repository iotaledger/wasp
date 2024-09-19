// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmtypes

import (
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/iotaledger/wasp/packages/util/bcs"
)

func init() {
	bcs.AddCustomEncoder(func(e *bcs.Encoder, tx *types.Transaction) error {
		return tx.EncodeRLP(e)
	})

	bcs.AddCustomDecoder(func(d *bcs.Decoder, tx *types.Transaction) error {
		return tx.DecodeRLP(rlp.NewStream(d, 0))
	})
}

func EncodeTransaction(tx *types.Transaction) []byte {
	return bcs.MustMarshal(tx)
}

func DecodeTransaction(data []byte) (*types.Transaction, error) {
	return bcs.UnmarshalOver(data, &types.Transaction{})
}
