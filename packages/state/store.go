// Copyright 2022 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package state

import (
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru"

	"github.com/iotaledger/hive.go/core/kvstore"
	"github.com/iotaledger/wasp/packages/kv/buffered"
	"github.com/iotaledger/wasp/packages/trie"
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
	trieRootByIndex map[uint32]trie.Hash

	// stateCache is a cache of immutable state readers by trie root. Reusing the
	// State instances allows to better take advantage of its internal caches.
	stateCache *lru.Cache // [trie.Hash]State
}

func NewStore(db kvstore.KVStore) Store {
	stateCache, err := lru.New(100)
	if err != nil {
		panic(err)
	}
	return &store{
		db:              &storeDB{db},
		trieRootByIndex: make(map[uint32]trie.Hash),
		stateCache:      stateCache,
	}
}

func (s *store) blockByTrieRoot(root trie.Hash) (*block, error) {
	return s.db.readBlock(root)
}

func (s *store) HasTrieRoot(root trie.Hash) bool {
	return s.db.hasBlock(root)
}

func (s *store) BlockByTrieRoot(root trie.Hash) (Block, error) {
	return s.blockByTrieRoot(root)
}

func (s *store) stateByTrieRoot(root trie.Hash) (*state, error) {
	if r, ok := s.stateCache.Get(root); ok {
		return r.(*state), nil
	}
	r, err := newState(s.db, root)
	if err != nil {
		return nil, err
	}
	s.stateCache.Add(root, r)
	return r, nil
}

func (s *store) StateByTrieRoot(root trie.Hash) (State, error) {
	return s.stateByTrieRoot(root)
}

func (s *store) NewOriginStateDraft() StateDraft {
	return newOriginStateDraft()
}

func (s *store) NewStateDraft(timestamp time.Time, prevL1Commitment *L1Commitment) (StateDraft, error) {
	prevState, err := s.stateByTrieRoot(prevL1Commitment.TrieRoot())
	if err != nil {
		return nil, err
	}
	return newStateDraft(timestamp, prevL1Commitment, prevState), nil
}

func (s *store) NewEmptyStateDraft(prevL1Commitment *L1Commitment) (StateDraft, error) {
	prevState, err := s.stateByTrieRoot(prevL1Commitment.TrieRoot())
	if err != nil {
		return nil, err
	}
	return newEmptyStateDraft(prevL1Commitment, prevState), nil
}

func (s *store) extractBlock(d StateDraft) (Block, *buffered.Mutations) {
	buf, bufDB := s.db.buffered()

	var baseTrieRoot trie.Hash
	{
		baseL1Commitment := d.BaseL1Commitment()
		if baseL1Commitment != nil {
			if !s.db.hasBlock(baseL1Commitment.TrieRoot()) {
				panic("cannot commit state: base trie root not found")
			}
			baseTrieRoot = baseL1Commitment.TrieRoot()
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

func (s *store) SetLatest(trieRoot trie.Hash) error {
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

	if s.trieRootByIndex[blockIndex] == block.TrieRoot() {
		// nothing to do
		return nil
	}

	isNextInSameBranch := func() bool {
		if blockIndex == 0 {
			return false
		}
		if _, ok := s.trieRootByIndex[blockIndex]; ok {
			return false
		}
		return s.trieRootByIndex[blockIndex-1] == block.PreviousL1Commitment().TrieRoot()
	}()
	if !isNextInSameBranch {
		// reorg
		s.trieRootByIndex = map[uint32]trie.Hash{}
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

func (s *store) findTrieRootByIndex(index uint32) (trie.Hash, error) {
	if cached, ok := func() (ret trie.Hash, ok bool) {
		s.mu.RLock()
		defer s.mu.RUnlock()
		ret, ok = s.trieRootByIndex[index]
		return
	}(); ok {
		return cached, nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	latestTrieRoot, err := s.db.latestTrieRoot()
	if err != nil {
		return trie.Hash{}, err
	}
	state, err := s.StateByTrieRoot(latestTrieRoot)
	if err != nil {
		return trie.Hash{}, err
	}
	latestBlockIndex := state.BlockIndex()
	s.trieRootByIndex[latestBlockIndex] = latestTrieRoot

	for i := latestBlockIndex; i > 0 && i > index; i-- {
		block, err := s.BlockByTrieRoot(s.trieRootByIndex[i])
		if err != nil {
			return trie.Hash{}, err
		}
		s.trieRootByIndex[i-1] = block.PreviousL1Commitment().TrieRoot()
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
