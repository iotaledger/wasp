// Copyright 2022 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package state

import (
	"sync"
	"time"

	"github.com/iotaledger/hive.go/core/kvstore"
	"github.com/iotaledger/trie.go/common"
	"github.com/iotaledger/wasp/packages/kv/buffered"
)

// store is the implementation of the Store interface
type store struct {
	// db is the backing key-value store
	db *storeDB

	// mu protects all accesses by block index, since it is mutable information
	mu sync.RWMutex

	// trieRootByIndex is a cache of index -> trieRoot, since the only one
	// stored in the db is the latestTrieRoot and all others have to be discovered by
	// traversing the block chain backwards
	trieRootByIndex map[uint32]common.VCommitment
}

func NewStore(db kvstore.KVStore) Store {
	return &store{
		db:              &storeDB{db},
		trieRootByIndex: make(map[uint32]common.VCommitment),
	}
}

func (s *store) blockByTrieRoot(root common.VCommitment) (*block, error) {
	return s.db.readBlock(root)
}

func (s *store) HasTrieRoot(root common.VCommitment) bool {
	return s.db.hasBlock(root)
}

func (s *store) BlockByTrieRoot(root common.VCommitment) (Block, error) {
	return s.blockByTrieRoot(root)
}

func (s *store) stateByTrieRoot(root common.VCommitment) (*state, error) {
	return newState(s.db, root)
}

func (s *store) StateByTrieRoot(root common.VCommitment) (State, error) {
	return s.stateByTrieRoot(root)
}

func (s *store) NewOriginStateDraft() StateDraft {
	return newOriginStateDraft()
}

func (s *store) NewStateDraft(timestamp time.Time, prevL1Commitment *L1Commitment) (StateDraft, error) {
	prevState, err := s.stateByTrieRoot(prevL1Commitment.GetTrieRoot())
	if err != nil {
		return nil, err
	}
	return newStateDraft(timestamp, prevL1Commitment, prevState), nil
}

func (s *store) NewEmptyStateDraft(prevL1Commitment *L1Commitment) (StateDraft, error) {
	prevState, err := s.stateByTrieRoot(prevL1Commitment.GetTrieRoot())
	if err != nil {
		return nil, err
	}
	return newEmptyStateDraft(prevL1Commitment, prevState), nil
}

func (s *store) extractBlock(d StateDraft) (Block, *buffered.Mutations) {
	buf, bufDB := s.db.buffered()

	var baseTrieRoot common.VCommitment
	{
		baseL1Commitment := d.BaseL1Commitment()
		if baseL1Commitment != nil {
			if !s.db.hasBlock(baseL1Commitment.GetTrieRoot()) {
				panic("cannot commit state: base trie root not found")
			}
			baseTrieRoot = baseL1Commitment.GetTrieRoot()
		} else {
			baseTrieRoot = bufDB.initTrie()
		}
	}

	// compute state db mutations
	block := func() Block {
		trie, err := bufDB.trieUpdatable(baseTrieRoot)
		if err != nil {
			// should not happen
			panic(err)
		}
		for k, v := range d.Mutations().Sets {
			trie.Update([]byte(k), v)
		}
		for k := range d.Mutations().Dels {
			trie.Delete([]byte(k))
		}
		trieRoot := trie.Commit(bufDB.trieStore())
		block := &block{
			trieRoot:             trieRoot,
			mutations:            d.Mutations(),
			previousL1Commitment: d.BaseL1Commitment(),
		}
		bufDB.saveBlock(block)
		return block
	}()

	return block, buf.muts
}

func (s *store) ExtractBlock(d StateDraft) Block {
	block, _ := s.extractBlock(d)
	return block
}

func (s *store) Commit(d StateDraft) Block {
	block, muts := s.extractBlock(d)
	s.db.commitToDB(muts)
	return block
}

func (s *store) SetLatest(trieRoot common.VCommitment) error {
	block, err := s.BlockByTrieRoot(trieRoot)
	if err != nil {
		return err
	}
	state, err := s.StateByTrieRoot(trieRoot)
	if err != nil {
		return err
	}
	blockIndex := state.BlockIndex()

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.trieRootByIndex[blockIndex] != nil && EqualCommitments(s.trieRootByIndex[blockIndex], block.TrieRoot()) {
		// nothing to do
		return nil
	}

	isNext := (blockIndex > 0 &&
		s.trieRootByIndex[blockIndex] == nil &&
		s.trieRootByIndex[blockIndex-1] != nil &&
		EqualCommitments(s.trieRootByIndex[blockIndex-1], block.PreviousL1Commitment().GetTrieRoot()))
	if !isNext {
		// reorg
		s.trieRootByIndex = map[uint32]common.VCommitment{}
	}
	s.trieRootByIndex[blockIndex] = block.TrieRoot()
	s.db.setLatestTrieRoot(trieRoot)
	return nil
}

func (s *store) BlockByIndex(index uint32) (Block, error) {
	root, err := s.findTrieRootByIndex(index)
	if err != nil {
		return nil, err
	}
	return s.BlockByTrieRoot(root)
}

func (s *store) findTrieRootByIndex(index uint32) (common.VCommitment, error) {
	if trieRoot := func() common.VCommitment {
		s.mu.RLock()
		defer s.mu.RUnlock()
		return s.trieRootByIndex[index]
	}(); trieRoot != nil {
		return trieRoot, nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	latestTrieRoot, err := s.db.latestTrieRoot()
	if err != nil {
		return nil, err
	}
	state, err := s.StateByTrieRoot(latestTrieRoot)
	if err != nil {
		return nil, err
	}
	latestBlockIndex := state.BlockIndex()
	s.trieRootByIndex[latestBlockIndex] = latestTrieRoot

	for i := latestBlockIndex; i > 0 && i > index; i-- {
		block, err := s.BlockByTrieRoot(s.trieRootByIndex[i])
		if err != nil {
			return nil, err
		}
		s.trieRootByIndex[i-1] = block.PreviousL1Commitment().GetTrieRoot()
	}
	return s.trieRootByIndex[index], nil
}

func (s *store) LatestBlock() (Block, error) {
	index, err := s.LatestBlockIndex()
	if err != nil {
		return nil, err
	}
	return s.BlockByIndex(index)
}

func (s *store) LatestBlockIndex() (uint32, error) {
	latestTrieRoot, err := s.db.latestTrieRoot()
	if err != nil {
		return 0, err
	}
	state, err := s.StateByTrieRoot(latestTrieRoot)
	if err != nil {
		return 0, err
	}
	return state.BlockIndex(), nil
}

func (s *store) LatestState() (State, error) {
	index, err := s.LatestBlockIndex()
	if err != nil {
		return nil, err
	}
	return s.StateByIndex(index)
}

func (s *store) StateByIndex(index uint32) (State, error) {
	block, err := s.BlockByIndex(index)
	if err != nil {
		return nil, err
	}
	return s.StateByTrieRoot(block.TrieRoot())
}
