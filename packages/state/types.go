// Copyright 2022 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package state

import (
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/trie.go/common"
	"github.com/iotaledger/trie.go/models/trie_blake2b"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/buffered"
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
	// BlockByTrieRoot fetches the Block that corresponds to the given trie root
	BlockByTrieRoot(common.VCommitment) Block
	// StateByTrieRoot returns the chain state corresponding to the given trie root
	StateByTrieRoot(common.VCommitment) State

	SetApprovingOutputID(trieRoot common.VCommitment, oid *iotago.UTXOInput) // TODO: remove?

	// SetLatest sets the given trie root to be considered the latest one in the chain.
	// This affects all `*ByIndex` and `Latest*` functions.
	SetLatest(trieRoot common.VCommitment)
	// BlockByIndex returns the block that corresponds to the given state index.
	// The index must not be greater than the index of the latest block (see SetLatest).
	BlockByIndex(uint32) Block
	// StateByIndex returns the chain state corresponding to the given state index.
	// The index must not be greater than the index of the latest block (see SetLatest).
	StateByIndex(uint32) State
	// LatestBlockIndex returns the index of the latest block (see SetLatest)
	LatestBlockIndex() uint32
	// LatestBlock returns the latest block of the chain (see SetLatest)
	LatestBlock() Block
	// LatestState returns the latest chain state (see SetLatest)
	LatestState() State

	// NewOriginStateDraft starts a new StateDraft for the origin block
	NewOriginStateDraft() StateDraft
	// NewOriginStateDraft starts a new StateDraft
	NewStateDraft(timestamp time.Time, prevL1Commitment *L1Commitment) StateDraft
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
	PreviousTrieRoot() common.VCommitment
	TrieRoot() common.VCommitment
	ApprovingOutputID() *iotago.UTXOInput
	setApprovingOutputID(*iotago.UTXOInput)
	Bytes() []byte
	Hash() BlockHash
	L1Commitment() *L1Commitment
}

type StateCommonValues interface {
	ChainID() *isc.ChainID
	BlockIndex() uint32
	Timestamp() time.Time
	PreviousL1Commitment() *L1Commitment
}

// State is an immutable view of a specific version of the chain state.
type State interface {
	kv.KVStoreReader
	TrieRoot() common.VCommitment
	GetMerkleProof(key []byte) *trie_blake2b.MerkleProof
	StateCommonValues
}

// StateDraft allows to mutate the chain state based on a specific trie root.
// All mutations are stored in-memory until committed.
type StateDraft interface {
	kv.KVStore
	BaseTrieRoot() common.VCommitment
	Mutations() *buffered.Mutations
	StateCommonValues
}
