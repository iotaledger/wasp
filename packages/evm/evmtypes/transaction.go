// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmtypes

import (
	"bytes"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
)

func EncodeTransaction(tx *types.Transaction) []byte {
	var b bytes.Buffer
	err := tx.EncodeRLP(&b)
	if err != nil {
		panic(err)
	}
	return b.Bytes()
}

func DecodeTransaction(b []byte) (*types.Transaction, error) {
	tx := new(types.Transaction)
	err := tx.DecodeRLP(rlp.NewStream(bytes.NewReader(b), 0))
	return tx, err
}

func GetSender(tx *types.Transaction) common.Address {
	sender, _ := types.Sender(Signer(tx.ChainId()), tx)
	return sender
}
