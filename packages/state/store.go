// Copyright 2022 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// Package state implements the core state management functionality for the IOTA ledger
package state

import (
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/samber/lo"

	"github.com/iotaledger/wasp/v2/packages/kv/buffered"
	"github.com/iotaledger/wasp/v2/packages/kvstore"
	"github.com/iotaledger/wasp/v2/packages/metrics"
	"github.com/iotaledger/wasp/v2/packages/trie"
)

// store is the implementation of the Store interface
type store struct {
	// db is the backing key-value store
	db *storeDB

	// stateCache is a cache of immutable state readers by trie root. Reusing the
	// State instances allows to better take advantage of its internal caches.
	stateCache *lru.Cache[trie.Hash, *state]

	metrics *metrics.ChainStateMetrics

	// writeMutex ensures that writes cannot be executed in parallel, because
	// the trie refcounts are mutable
	writeMutex *sync.Mutex

	refcountsEnabled bool
}

const cacheSize = 100

func NewStore(db kvstore.KVStore, refcountsEnabled bool, writeMutex *sync.Mutex) (Store, error) {
	return NewStoreWithMetrics(db, refcountsEnabled, writeMutex, nil)
}

func NewStoreWithMetrics(db kvstore.KVStore, refcountsEnabled bool, writeMutex *sync.Mutex, metrics *metrics.ChainStateMetrics) (Store, error) {
	stateCache, err := lru.New[trie.Hash, *state](cacheSize)
	if err != nil {
		return nil, err
	}

	storedb := &storeDB{db}
	err = trie.UpdateRefcountsFlag(trieStore(storedb), refcountsEnabled)
	if err != nil {
		return nil, err
	}

	return &store{
		db:               storedb,
		stateCache:       stateCache,
		metrics:          metrics,
		writeMutex:       writeMutex,
		refcountsEnabled: refcountsEnabled,
	}, nil
}

func NewStoreReadonly(db kvstore.KVStore) (Store, error) {
	stateCache, err := lru.New[trie.Hash, *state](cacheSize)
	if err != nil {
		return nil, err
	}

	storedb := &storeDB{db}
	refcountsEnabled := trie.IsRefcountsEnabled(trieStore(storedb))

	return &store{
		db:               storedb,
		stateCache:       stateCache,
		metrics:          nil,
		writeMutex:       nil, // will not panic if only read-only methods are called
		refcountsEnabled: refcountsEnabled,
	}, nil
}

func (s *store) blockByTrieRoot(root trie.Hash) (Block, error) {
	return s.db.readBlock(root)
}

func (s *store) IsEmpty() bool {
	return s.db.isEmpty()
}

func (s *store) HasTrieRoot(root trie.Hash) bool {
	return s.db.hasBlock(root)
}

func (s *store) BlockByTrieRoot(root trie.Hash) (Block, error) {
	return s.blockByTrieRoot(root)
}

func (s *store) stateByTrieRoot(root trie.Hash) (*state, error) {
	if r, ok := s.stateCache.Get(root); ok {
		return r, nil
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
	if prevL1Commitment == nil {
		return nil, errors.New("nil prevL1Commitment")
	}
	prevState, err := s.stateByTrieRoot(prevL1Commitment.TrieRoot())
	if err != nil {
		return nil, err
	}
	return newEmptyStateDraft(prevL1Commitment, prevState), nil
}

func (s *store) extractBlock(d StateDraft) (
	newBlock Block,
	muts *buffered.Mutations,
	stats *trie.CommitStats,
	err error,
) {
	buf, bufDB := s.db.buffered()

	var baseTrieRoot trie.Hash
	{
		if d == nil {
			return nil, nil, nil, errors.New("state.StateDraft is nil")
		}

		baseL1Commitment := d.BaseL1Commitment()
		if baseL1Commitment != nil {
			if !s.db.hasBlock(baseL1Commitment.TrieRoot()) {
				return nil, nil, nil, errors.New("cannot commit state: base trie root not found")
			}
			baseTrieRoot = baseL1Commitment.TrieRoot()
		} else {
			// will be pruned later
			baseTrieRoot, err = bufDB.initTrie(s.refcountsEnabled)
			if err != nil {
				return nil, nil, nil, err
			}
		}
	}

	// compute state db mutations
	newBlock, stats = func() (Block, *trie.CommitStats) {
		tr, err := bufDB.trieUpdatable(baseTrieRoot)
		if err != nil {
			// should not happen
			panic(err)
		}
		for k, v := range d.Mutations().Sets {
			tr.Update([]byte(k), v)
		}
		for k := range d.Mutations().Dels {
			tr.Delete([]byte(k))
		}
		trieRoot, _, commitStats := tr.Commit(trieStore(bufDB))
		block := &block{
			trieRoot:             trieRoot,
			mutations:            d.Mutations(),
			previousL1Commitment: d.BaseL1Commitment(),
		}
		bufDB.saveBlock(block)
		return block, commitStats
	}()

	if s.refcountsEnabled && d.BaseL1Commitment() == nil {
		// we must prune the baseTrieRoot that we created above
		_, err := trie.Prune(trieStore(bufDB), baseTrieRoot)
		if err != nil {
			return nil, nil, nil, err
		}
	}

	return newBlock, buf.muts, stats, nil
}

func (s *store) ExtractBlock(d StateDraft) (Block, error) {
	block, _, _, err := s.extractBlock(d)
	return block, err
}

func (s *store) Commit(d StateDraft) (Block, bool, *trie.CommitStats, error) {
	s.writeMutex.Lock()
	defer s.writeMutex.Unlock()

	start := time.Now()
	block, muts, stats, err := s.extractBlock(d)
	if err != nil {
		return nil, false, nil, err
	}
	s.db.commitToDB(muts)
	if s.metrics != nil {
		var createdNodes, createdValues uint
		if s.refcountsEnabled {
			createdNodes = stats.CreatedNodes
			createdValues = stats.CreatedValues
		}
		s.metrics.BlockCommitted(time.Since(start), s.refcountsEnabled, createdNodes, createdValues)
	}
	return block, s.refcountsEnabled, stats, nil
}

func (s *store) Prune(trieRoot trie.Hash) (trie.PruneStats, error) {
	s.writeMutex.Lock()
	defer s.writeMutex.Unlock()

	start := time.Now()
	state, err := s.StateByTrieRoot(trieRoot)
	if err != nil {
		return trie.PruneStats{}, err
	}
	blockIndex := state.BlockIndex()
	buf, bufDB := s.db.buffered()
	stats, err := trie.Prune(trieStore(bufDB), trieRoot)
	if err != nil {
		return trie.PruneStats{}, err
	}
	s.db.pruneBlock(trieRoot)
	s.db.commitToDB(buf.muts)
	s.stateCache.Remove(trieRoot)
	s.updateLargestPrunedBlockIndex(blockIndex)
	if s.metrics != nil {
		s.metrics.BlockPruned(time.Since(start), stats.DeletedNodes, stats.DeletedValues)
	}
	return stats, nil
}

func (s *store) updateLargestPrunedBlockIndex(prunedBlockIndex uint32) {
	largestPrunedBlockIndex, err := s.db.largestPrunedBlockIndex()
	if errors.Is(err, ErrNoBlocksPruned) {
		s.db.setLargestPrunedBlockIndex(prunedBlockIndex)
	} else if err != nil {
		panic(err) // should not happen: no other error can be returned from `largestPrunedBlockIndex`
	} else if prunedBlockIndex > largestPrunedBlockIndex {
		s.db.setLargestPrunedBlockIndex(prunedBlockIndex)
	}
}

func (s *store) LargestPrunedBlockIndex() (uint32, error) {
	return s.db.largestPrunedBlockIndex()
}

func (s *store) SetLatest(trieRoot trie.Hash) error {
	s.writeMutex.Lock()
	defer s.writeMutex.Unlock()

	_, err := s.BlockByTrieRoot(trieRoot)
	if err != nil {
		return err
	}
	s.db.setLatestTrieRoot(trieRoot)
	return nil
}

func (s *store) ClearLatest() error {
	s.writeMutex.Lock()
	defer s.writeMutex.Unlock()

	s.db.mustDel(keyLatestTrieRoot())
	return nil
}

func (s *store) LatestBlock() (Block, error) {
	root, err := s.db.latestTrieRoot()
	if err != nil {
		return nil, err
	}
	return s.BlockByTrieRoot(root)
}

func (s *store) LatestBlockIndex() (uint32, error) {
	latestTrieRoot, err := s.LatestTrieRoot()
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
	root, err := s.db.latestTrieRoot()
	if err != nil {
		return nil, err
	}
	return s.StateByTrieRoot(root)
}

func (s *store) LatestTrieRoot() (trie.Hash, error) {
	return s.db.latestTrieRoot()
}

func (s *store) TakeSnapshot(root trie.Hash, w io.Writer) error {
	return s.db.takeSnapshot(root, w)
}

func (s *store) RestoreSnapshot(root trie.Hash, r io.Reader, refcountsEnabled bool) error {
	if s.db.hasBlock(root) {
		return nil
	}

	s.writeMutex.Lock()
	defer s.writeMutex.Unlock()

	err := s.db.restoreSnapshot(root, r, refcountsEnabled)
	if err != nil {
		return err
	}

	block, err := s.blockByTrieRoot(root)
	if err != nil {
		return err
	}
	if block.PreviousL1Commitment() != nil {
		previousTrieRoot := block.PreviousL1Commitment().TrieRoot()
		if !s.db.hasBlock(previousTrieRoot) {
			s.updateLargestPrunedBlockIndex(block.StateIndex() - 1)
		}
	}
	return nil
}

func (s *store) IsRefcountsEnabled() bool {
	return trie.IsRefcountsEnabled(trieStore(s.db))
}

func (s *store) CheckIntegrity(w io.Writer) {
	fmt.Fprint(w, "[begin store::CheckIntegrity]\n")

	// check PrefixLatestTrieRoot
	if s.db.hasLatestTrieRoot() {
		latestTrieRoot := lo.Must(s.LatestTrieRoot())
		if !s.HasTrieRoot(latestTrieRoot) {
			panic(fmt.Sprintf("latest trie root %s not found", latestTrieRoot))
		}
		fmt.Fprintf(w, "latest trie root: %s\n", latestTrieRoot)
	}

	// check PrefixLargestPrunedBlockIndex
	latestPruned, latestErr := s.db.largestPrunedBlockIndex()
	hasPrunedBlocks := latestErr == nil
	if hasPrunedBlocks {
		fmt.Fprintf(w, "latest pruned block index: %d\n", latestPruned)
	}

	// check PrefixBlockByTrieRoot
	fmt.Fprint(w, "checking blocks...\n")
	var trieRoots []trie.Hash
	_, refcounts := trie.NewRefcounts(trieStore(s.db))
	lo.Must0(s.db.IterateKeys(keyBlockByTrieRootNoTrieRoot(), func(key kvstore.Key) bool {
		trieRoot := trie.Hash(key[1:])
		var n uint32
		if refcounts != nil {
			n = refcounts.GetNode(trieRoot)
			if n == 0 {
				panic(fmt.Sprintf("trie root %s has refcount 0", trieRoot))
			}
		} else {
			n = 1
		}
		for range n {
			trieRoots = append(trieRoots, trieRoot)
		}
		return true
	}))
	for _, trieRoot := range trieRoots {
		block := lo.Must(s.BlockByTrieRoot(trieRoot))
		if block.PreviousL1Commitment() != nil {
			previousTrieRoot := block.PreviousL1Commitment().TrieRoot()
			if s.HasTrieRoot(previousTrieRoot) {
				previousBlock := lo.Must(s.BlockByTrieRoot(previousTrieRoot))
				if block.StateIndex() != previousBlock.StateIndex()+1 {
					panic(fmt.Sprintf("block index mismatch: expected %d, got %d", previousBlock.StateIndex()+1, block.StateIndex()))
				}
			} else if !hasPrunedBlocks || latestPruned < block.StateIndex()-1 {
				panic(fmt.Sprintf("previous block %d %s not found", block.StateIndex()-1, previousTrieRoot))
			}
		}
		fmt.Fprintf(w, "block %d %s .. ok\n", block.StateIndex(), block.TrieRoot())
	}

	// check PrefixTrie
	trie.DebugDump(trieStore(s.db), trieRoots, w)

	fmt.Fprint(w, "[end store::CheckIntegrity]\n")
}
