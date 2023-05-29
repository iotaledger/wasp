// Copyright 2022 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package indexedstore

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/subrealm"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/trie"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
)

// IndexedStore augments a Store with functions to search blocks by index.
type IndexedStore interface {
	state.Store

	// BlockByIndex returns the block that corresponds to the given state index.
	BlockByIndex(uint32) (state.Block, error)
	// StateByIndex returns the chain state corresponding to the given state index
	StateByIndex(uint32) (state.State, error)
}

type istore struct {
	state.Store
}

// New returns an IndexedStore implemented by getting the blockinfo from the latest state.
func New(s state.Store) IndexedStore {
	return &istore{
		Store: s,
	}
}

func (s *istore) BlockByIndex(index uint32) (state.Block, error) {
	root, err := s.findTrieRootByIndex(index)
	if err != nil {
		return nil, err
	}
	return s.BlockByTrieRoot(root)
}

func (s *istore) StateByIndex(index uint32) (state.State, error) {
	block, err := s.BlockByIndex(index)
	if err != nil {
		return nil, err
	}
	return s.StateByTrieRoot(block.TrieRoot())
}

func (s *istore) findTrieRootByIndex(index uint32) (trie.Hash, error) {
	latestState, err := s.LatestState()
	if err != nil {
		return trie.Hash{}, err
	}

	latestIndex := latestState.BlockIndex()
	if index > latestIndex {
		return trie.Hash{}, fmt.Errorf(
			"block %d not found (latest index is %d)",
			index, latestIndex,
		)
	}
	if index == latestIndex {
		return latestState.TrieRoot(), nil
	}
	blocklogStatePartition := subrealm.NewReadOnly(latestState, kv.Key(blocklog.Contract.Hname().Bytes()))
	nextBlockIndex := index + 1
	nextBlockInfo, ok := blocklog.GetBlockInfo(blocklogStatePartition, nextBlockIndex)
	if !ok {
		return trie.Hash{}, fmt.Errorf("block not found: %d", nextBlockIndex)
	}
	return nextBlockInfo.PreviousL1Commitment().TrieRoot(), nil
}

type fakeistore struct {
	state.Store
}

// NewFake returns an implementation of IndexedStore that searches blocks by
// traversing the chain from the latest block.
func NewFake(s state.Store) IndexedStore {
	return &fakeistore{
		Store: s,
	}
}

func (s *fakeistore) BlockByIndex(index uint32) (state.Block, error) {
	latestBlock, err := s.LatestBlock()
	if err != nil {
		return nil, err
	}

	latestIndex := latestBlock.StateIndex()
	if index > latestIndex {
		return nil, fmt.Errorf(
			"block %d not found (latest index is %d)",
			index, latestIndex,
		)
	}
	if index == latestIndex {
		return latestBlock, nil
	}
	block := latestBlock
	for block.StateIndex() > index {
		block, err = s.BlockByTrieRoot(block.PreviousL1Commitment().TrieRoot())
		if err != nil {
			return nil, err
		}
	}
	return block, nil
}

func (s *fakeistore) StateByIndex(index uint32) (state.State, error) {
	block, err := s.BlockByIndex(index)
	if err != nil {
		return nil, err
	}
	return s.StateByTrieRoot(block.TrieRoot())
}
