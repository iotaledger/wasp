package utils

import (
	"fmt"
	"io"
	"time"

	"github.com/iotaledger/wasp/v2/packages/state"
	"github.com/iotaledger/wasp/v2/packages/trie"
)

type readOnlyStore struct {
	store state.Store
}

var _ state.Store = &readOnlyStore{}

func NewReadOnlyStore(store state.Store) state.Store {
	return &readOnlyStore{store: store}
}

func (ros *readOnlyStore) IsEmpty() bool {
	return ros.store.IsEmpty()
}

func (ros *readOnlyStore) HasTrieRoot(trieRoot trie.Hash) bool {
	return ros.store.HasTrieRoot(trieRoot)
}

func (ros *readOnlyStore) BlockByTrieRoot(trieRoot trie.Hash) (state.Block, error) {
	return ros.store.BlockByTrieRoot(trieRoot)
}

func (ros *readOnlyStore) StateByTrieRoot(trieRoot trie.Hash) (state.State, error) {
	return ros.store.StateByTrieRoot(trieRoot)
}

func (ros *readOnlyStore) SetLatest(trie.Hash) error {
	return fmt.Errorf("cannot write to read-only store")
}

func (ros *readOnlyStore) LatestBlockIndex() (uint32, error) {
	return ros.store.LatestBlockIndex()
}

func (ros *readOnlyStore) LatestBlock() (state.Block, error) {
	return ros.store.LatestBlock()
}

func (ros *readOnlyStore) LatestState() (state.State, error) {
	return ros.store.LatestState()
}

func (ros *readOnlyStore) LatestTrieRoot() (trie.Hash, error) {
	return ros.store.LatestTrieRoot()
}

func (ros *readOnlyStore) NewOriginStateDraft() state.StateDraft {
	panic("Cannot create origin state draft in read-only store")
}

func (ros *readOnlyStore) NewStateDraft(time.Time, *state.L1Commitment) (state.StateDraft, error) {
	return nil, fmt.Errorf("cannot create state draft in read-only store")
}

func (ros *readOnlyStore) NewEmptyStateDraft(prevL1Commitment *state.L1Commitment) (state.StateDraft, error) {
	return nil, fmt.Errorf("cannot create empty state draft in read-only store")
}

func (ros *readOnlyStore) Commit(state.StateDraft) (state.Block, trie.CommitStats) {
	panic("Cannot commit to read-only store")
}

func (ros *readOnlyStore) ExtractBlock(stateDraft state.StateDraft) state.Block {
	return ros.store.ExtractBlock(stateDraft)
}

func (ros *readOnlyStore) Prune(trie.Hash) (trie.PruneStats, error) {
	panic("Cannot prune read-only store")
}

func (ros *readOnlyStore) LargestPrunedBlockIndex() (uint32, error) {
	return ros.store.LargestPrunedBlockIndex()
}

func (ros *readOnlyStore) TakeSnapshot(trieRoot trie.Hash, w io.Writer) error {
	return ros.store.TakeSnapshot(trieRoot, w)
}

func (ros *readOnlyStore) RestoreSnapshot(trie.Hash, io.Reader) error {
	return fmt.Errorf("cannot write snapshot into read-only store")
}
