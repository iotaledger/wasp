package sm_snapshots

import (
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/trie"
)

type SnapshotManager interface {
	UpdateAsync()
	BlockCommittedAsync(SnapshotInfo)
	SnapshotExists(uint32, *state.L1Commitment) bool
	LoadSnapshotAsync(SnapshotInfo) <-chan error
}

type SnapshotInfo interface {
	GetStateIndex() uint32
	GetCommitment() *state.L1Commitment
	GetTrieRoot() trie.Hash
	GetBlockHash() state.BlockHash
}

type snapshotter interface {
	createSnapshotAsync(stateIndex uint32, commitment *state.L1Commitment, doneCallback func())
	loadSnapshot(filePath string) error
}

type snapshotList interface {
	GetStateIndex() uint32
	GetL1Commitments() []*state.L1Commitment
	Join()
}
