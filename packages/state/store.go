// Copyright 2022 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package state

import (
	"sync"
	"time"

	"github.com/iotaledger/hive.go/core/kvstore"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/trie.go/common"
	"github.com/iotaledger/wasp/packages/kv/buffered"
)

// store is the implementation of the Store interface
type store struct {
	db              *storeDB
	mu              sync.RWMutex
	trieRootByIndex map[uint32]common.VCommitment // TODO: store in db? just latest trie root
}

func NewStore(db kvstore.KVStore) Store {
	return &store{
		db:              &storeDB{db},
		trieRootByIndex: make(map[uint32]common.VCommitment),
	}
}

func (s *store) blockByTrieRoot(root common.VCommitment) *block {
	b, err := s.db.readBlock(root)
	mustNoErr(err)
	return b
}

func (s *store) BlockByTrieRoot(root common.VCommitment) Block {
	return s.blockByTrieRoot(root)
}

func (s *store) stateByTrieRoot(root common.VCommitment) *state {
	return newState(s.db, root)
}

func (s *store) StateByTrieRoot(root common.VCommitment) State {
	return s.stateByTrieRoot(root)
}

func (s *store) NewOriginStateDraft() StateDraft {
	return newOriginStateDraft()
}

func (s *store) NewStateDraft(timestamp time.Time, prevL1Commitment *L1Commitment) StateDraft {
	return newStateDraft(s.stateByTrieRoot(prevL1Commitment.StateCommitment), timestamp, prevL1Commitment)
}

func (s *store) extractBlock(d StateDraft) (Block, *buffered.Mutations) {
	buf, bufDB := s.db.buffered()

	// compute trie db mutations
	baseTrieRoot := d.BaseTrieRoot()
	if baseTrieRoot != nil {
		if !s.db.hasBlock(d.BaseTrieRoot()) {
			panic("cannot commit state: base trie root not found")
		}
	} else {
		baseTrieRoot = bufDB.initTrie()
	}

	// compute state db mutations
	block := func() Block {
		trie := bufDB.trieUpdatable(baseTrieRoot)
		for k, v := range d.Mutations().Sets {
			trie.Update([]byte(k), v)
		}
		for k := range d.Mutations().Dels {
			trie.Delete([]byte(k))
		}
		trieRoot := trie.Commit(bufDB.trieStore())
		block := &block{
			mutations:        d.Mutations(),
			trieRoot:         trieRoot,
			previousTrieRoot: d.BaseTrieRoot(),
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

func (s *store) SetApprovingOutputID(trieRoot common.VCommitment, oid *iotago.UTXOInput) {
	block := s.BlockByTrieRoot(trieRoot)
	block.setApprovingOutputID(oid)
	s.db.saveBlock(block)
}

func (s *store) SetLatest(trieRoot common.VCommitment) {
	block := s.BlockByTrieRoot(trieRoot)
	blockIndex := s.StateByTrieRoot(trieRoot).BlockIndex()

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.trieRootByIndex[blockIndex] != nil && EqualCommitments(s.trieRootByIndex[blockIndex], block.TrieRoot()) {
		// nothing to do
		return
	}

	isNext := (blockIndex > 0 &&
		s.trieRootByIndex[blockIndex] == nil &&
		s.trieRootByIndex[blockIndex-1] != nil &&
		EqualCommitments(s.trieRootByIndex[blockIndex-1], block.PreviousTrieRoot()))
	if !isNext {
		// reorg
		s.trieRootByIndex = map[uint32]common.VCommitment{}
	}
	s.trieRootByIndex[blockIndex] = block.TrieRoot()
	s.db.setLatestTrieRoot(trieRoot)
}

func (s *store) BlockByIndex(index uint32) Block {
	return s.BlockByTrieRoot(s.findTrieRootByIndex(index))
}

func (s *store) findTrieRootByIndex(index uint32) common.VCommitment {
	trieRoot := func() common.VCommitment {
		s.mu.RLock()
		defer s.mu.RUnlock()
		return s.trieRootByIndex[index]
	}()
	if trieRoot != nil {
		return trieRoot
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	latestTrieRoot := s.db.latestTrieRoot()
	latestBlockIndex := s.StateByTrieRoot(latestTrieRoot).BlockIndex()
	s.trieRootByIndex[latestBlockIndex] = latestTrieRoot

	for i := latestBlockIndex; i > 0 && i > index; i-- {
		s.trieRootByIndex[i-1] = s.BlockByTrieRoot(s.trieRootByIndex[i]).PreviousTrieRoot()
	}
	return s.trieRootByIndex[index]
}

func (s *store) LatestBlock() Block {
	return s.BlockByIndex(s.LatestBlockIndex())
}

func (s *store) LatestBlockIndex() uint32 {
	latestTrieRoot := s.db.latestTrieRoot()
	return s.StateByTrieRoot(latestTrieRoot).BlockIndex()
}

func (s *store) LatestState() State {
	return s.StateByIndex(s.LatestBlockIndex())
}

func (s *store) StateByIndex(index uint32) State {
	return s.StateByTrieRoot(s.BlockByIndex(index).TrieRoot())
}
