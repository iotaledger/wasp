// Copyright 2022 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package state

import (
	"time"

	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/buffered"
	"github.com/iotaledger/wasp/packages/trie"
)

// Store manages the storage of a chain's state.
//
// The chain state is a key-value store that is updated every time a batch
// of requests is executed by the VM.
//
// The purpose of the Store is to store not only the latest version of the chain
// state, but also past versions (up to a limit).
//
// Each version of the key-value pairs is stored in an immutable trie (provided by
// the trie.go package). Therefore each *state index* corresponds to a unique
// *trie root*.
//
// For each trie root, the Store also stores a Block, which contains the mutations
// between the previous and current states, and allows to calculate the L1 commitment.
type Store interface {
	// HasTrieRoot returns true if the given trie root exists in the store
	HasTrieRoot(trie.Hash) bool
	// BlockByTrieRoot fetches the Block that corresponds to the given trie root
	BlockByTrieRoot(trie.Hash) (Block, error)
	// StateByTrieRoot returns the chain state corresponding to the given trie root
	StateByTrieRoot(trie.Hash) (State, error)

	// SetLatest sets the given trie root to be considered the latest one in the chain.
	// This affects all `*ByIndex` and `Latest*` functions.
	SetLatest(trieRoot trie.Hash) error
	// BlockByIndex returns the block that corresponds to the given state index (see SetLatest).
	BlockByIndex(uint32) (Block, error)
	// StateByIndex returns the chain state corresponding to the given state index (see SetLatest).
	StateByIndex(uint32) (State, error)
	// LatestBlockIndex returns the index of the latest block, if set (see SetLatest)
	LatestBlockIndex() (uint32, error)
	// LatestBlock returns the latest block of the chain, if set (see SetLatest)
	LatestBlock() (Block, error)
	// LatestState returns the latest chain state, if set (see SetLatest)
	LatestState() (State, error)

	// NewOriginStateDraft starts a new StateDraft for the origin block
	NewOriginStateDraft() StateDraft

	// NewStateDraft starts a new StateDraft.
	// The newly created StateDraft will already contain a few mutations updating the common values:
	// - timestamp
	// - block index
	// - previous L1 commitment
	NewStateDraft(timestamp time.Time, prevL1Commitment *L1Commitment) (StateDraft, error)

	// NewEmptyStateDraft starts a new StateDraft without updating any the common values.
	// It may be used to replay a block given the mutations.
	// Note that calling any of the StateCommonValues methods may return invalid data before
	// applying the mutations.
	NewEmptyStateDraft(prevL1Commitment *L1Commitment) (StateDraft, error)

	// Commit commits the given state, creating a new block and trie root in the DB.
	// SetLatest must be called manually to consider the new state as the latest one.
	Commit(StateDraft) Block

	// ExtractBlock performs a dry-run of Commit, discarding all changes that would be
	// made to the DB.
	ExtractBlock(StateDraft) Block
}

// A Block contains the mutations between the previous and current states,
// and allows to calculate the L1 commitment.
// Blocks are immutable.
type Block interface {
	Mutations() *buffered.Mutations
	MutationsReader() kv.KVStoreReader
	TrieRoot() trie.Hash
	PreviousL1Commitment() *L1Commitment
	StateIndex() uint32
	// L1Commitment contains the TrieRoot + block Hash
	L1Commitment() *L1Commitment
	// Hash is computed from Mutations + PreviousL1Commitment
	Hash() BlockHash
	Bytes() []byte
}

type StateCommonValues interface {
	BlockIndex() uint32
	Timestamp() time.Time
	PreviousL1Commitment() *L1Commitment
}

// State is an immutable view of a specific version of the chain state.
type State interface {
	kv.KVStoreReader
	TrieRoot() trie.Hash
	GetMerkleProof(key []byte) *trie.MerkleProof
	StateCommonValues
}

// StateDraft allows to mutate the chain state based on a specific trie root.
// All mutations are stored in-memory until committed.
type StateDraft interface {
	kv.KVStore
	BaseL1Commitment() *L1Commitment
	Mutations() *buffered.Mutations
	StateCommonValues
}
