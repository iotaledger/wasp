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

// Putting slice into a map is not acceptable as if you want to append to slice,
// you'll have to re-include the appended slice into the map.
type SliceStruct[E any] interface {
	Add(E)
	Get(int) E
	Set(int, E)
	Length() int
	ForEach(func(int, E) bool) bool
	Clone() SliceStruct[E]              // Returns a new SliceStruct with exactly the same elements
	CloneDeep(func(E) E) SliceStruct[E] // Returns a new SliceStruct with every element of the old SliceStruct cloned using provided function
	ContainsBy(func(E) bool) bool
	Find(func(E) bool) (E, bool)
}
