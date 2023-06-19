// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmtypes

import (
	"bytes"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
)

// EncodeReceipt serializes the receipt in RLP format
func EncodeReceipt(receipt *types.Receipt) []byte {
	w := new(bytes.Buffer)
	_ = receipt.EncodeRLP(w)
	return w.Bytes()
}

func DecodeReceipt(b []byte) (*types.Receipt, error) {
	receipt := new(types.Receipt)
	err := receipt.DecodeRLP(rlp.NewStream(bytes.NewReader(b), 0))
	return receipt, err
}

func BloomFilter(bloom types.Bloom, addresses []common.Address, topics [][]common.Hash) bool {
	return bloomMatchesAddresses(bloom, addresses) && bloomMatchesAllEvents(bloom, topics)
}

func bloomMatchesAddresses(bloom types.Bloom, addresses []common.Address) bool {
	if len(addresses) == 0 {
		return true
	}
	for _, addr := range addresses {
		if types.BloomLookup(bloom, addr) {
			return true
		}
	}
	return false
}

func bloomMatchesAllEvents(bloom types.Bloom, events [][]common.Hash) bool {
	for _, topics := range events {
		if !bloomMatchesAnyTopic(bloom, topics) {
			return false
		}
	}
	return true
}

func bloomMatchesAnyTopic(bloom types.Bloom, topics []common.Hash) bool {
	if len(topics) == 0 {
		return true
	}
	for _, topic := range topics {
		if types.BloomLookup(bloom, topic) {
			return true
		}
	}
	return false
}
