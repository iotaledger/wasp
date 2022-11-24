// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package smGPAUtils

import (
	"github.com/iotaledger/wasp/packages/state"
)

func ContainsBlockHash(blockHash state.BlockHash, blockHashes []state.BlockHash) bool {
	for _, bh := range blockHashes {
		if bh.Equals(blockHash) {
			return true
		}
	}
	return false
}

func DeleteBlockHash(blockHash state.BlockHash, blockHashes []state.BlockHash) []state.BlockHash {
	for i := range blockHashes {
		if blockHashes[i].Equals(blockHash) {
			blockHashes[i] = blockHashes[len(blockHashes)-1]
			return blockHashes[:len(blockHashes)-1]
		}
	}
	return blockHashes
}

func RemoveAllBlockHashes(blockHashesToRemove []state.BlockHash, blockHashes []state.BlockHash) []state.BlockHash {
	result := blockHashes
	for i := range blockHashesToRemove {
		result = DeleteBlockHash(blockHashesToRemove[i], result)
	}
	return result
}

func AllDifferentBlockHashes(blockHashes []state.BlockHash) bool {
	for i := range blockHashes {
		for j := 0; j < i; j++ {
			if blockHashes[i].Equals(blockHashes[j]) {
				return false
			}
		}
	}
	return true
}
