// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmtypes

import (
	"bytes"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
)

func EncodeBlock(block *types.Block) []byte {
	var b bytes.Buffer
	err := block.EncodeRLP(&b)
	if err != nil {
		panic(err)
	}
	return b.Bytes()
}

func DecodeBlock(b []byte) (*types.Block, error) {
	block := new(types.Block)
	err := block.DecodeRLP(rlp.NewStream(bytes.NewReader(b), 0))
	return block, err
}
