package sm_snapshots

import (
	"io"

	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/trie"
)

type SnapshotManager interface {
	UpdateAsync()
	BlockCommittedAsync(SnapshotInfo)
	SnapshotExists(uint32, *state.L1Commitment) bool
	LoadSnapshotAsync(SnapshotInfo) <-chan error
}

type SnapshotManagerTest interface {
	SnapshotManager
	SnapshotReady(SnapshotInfo)
	IsSnapshotReady(SnapshotInfo) bool
	SetAfterSnapshotCreated(func(SnapshotInfo))
}

type SnapshotInfo interface {
	GetStateIndex() uint32
	GetCommitment() *state.L1Commitment
	GetTrieRoot() trie.Hash
	GetBlockHash() state.BlockHash
	String() string
	Equals(SnapshotInfo) bool
}

type snapshotManagerCore interface {
	createSnapshotsNeeded() bool
	handleUpdate()
	handleBlockCommitted(SnapshotInfo)
	handleLoadSnapshot(SnapshotInfo, chan<- error)
}

type snapshotter interface {
	storeSnapshot(SnapshotInfo, io.Writer) error
	loadSnapshot(SnapshotInfo, io.Reader) error
}
