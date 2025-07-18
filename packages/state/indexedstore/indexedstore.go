// Copyright 2022 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// Package indexedstore provides functionality to search blocks by index within the state store
package indexedstore

import (
	"fmt"

	"fortio.org/safecast"

	"github.com/iotaledger/wasp/v2/packages/state"
	"github.com/iotaledger/wasp/v2/packages/trie"
	"github.com/iotaledger/wasp/v2/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/v2/packages/vm/core/governance"
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

	// iterate until we find the next block (that contains the L1 commitment for the block we are looking for)
	targetBlockIndex := index + 1
	state := latestState

	blockKeepAmount := governance.NewStateReaderFromChainState(state).GetBlockKeepAmount() // block keep amount cannot be changed (its set in stone from origin)

	for blockKeepAmount != -1 { // no need to iterate if pruning is disabled
		blocklogState := blocklog.NewStateReaderFromChainState(state)
		earliestAvailableBlockIndex := uint32(0)
		blockKeepAmountUint32, err := safecast.Convert[uint32](blockKeepAmount)
		if err != nil {
			return trie.Hash{}, fmt.Errorf("iterating the chain: blockKeepAmount is not a uint32: %w", err)
		}
		if blockKeepAmountUint32 < state.BlockIndex() {
			earliestAvailableBlockIndex = state.BlockIndex() - blockKeepAmountUint32 + 1
		}
		if targetBlockIndex >= earliestAvailableBlockIndex {
			break // found it
		}
		bi, ok := blocklogState.GetBlockInfo(earliestAvailableBlockIndex + 1) // get +1 to make things easier and get the actual block (because we do previousL1Commitment)
		if !ok {
			return trie.Hash{}, fmt.Errorf("iterating the chain: blocklog missing block index %d on active state %d", earliestAvailableBlockIndex, state.BlockIndex())
		}
		state, err = s.StateByTrieRoot(bi.PreviousL1Commitment().TrieRoot())
		if err != nil {
			return trie.Hash{}, err
		}
	}
	nextBlockInfo, ok := blocklog.NewStateReaderFromChainState(state).GetBlockInfo(targetBlockIndex)
	if !ok {
		return trie.Hash{}, fmt.Errorf("blocklog missing block index %d on active state %d", targetBlockIndex, state.BlockIndex())
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
