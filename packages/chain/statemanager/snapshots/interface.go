package snapshots

import (
	"io"

	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/trie"
)

// SnapshotManager is responsible for servicing snapshot related queries in appropriate
// manner. Some of the requests are synchronous, but most of them are asynchronous. They
// can be handled in snapshot manager's thread or in another thread created by snapshot
// manager.
// Snapshot manager keeps and updates on request a list of available snapshots. However,
// only the information that snapshot exists is stored and not the entire snapshot. To
// store/load the snapshot, snapshot manager depends on `snapshotter`.
// Snapshot manager is also responsible for deciding if snapshot has to be created.
type SnapshotManager interface {
	GetLoadedSnapshotStateIndex() uint32
	BlockCommittedAsync(SnapshotInfo)
}

type SnapshotInfo interface {
	StateIndex() uint32
	Commitment() *state.L1Commitment
	TrieRoot() trie.Hash
	BlockHash() state.BlockHash
	String() string
	Equals(SnapshotInfo) bool
}

type snapshotManagerCore interface {
	createSnapshot(SnapshotInfo)
	loadSnapshot() SnapshotInfo
}

// snapshotter is responsible for moving the snapshot between store and external
// sources/destinations. It can:
// * take required snapshot from the store and write it to some `Writer` (`storeSnapshot` method)
// * read the snapshot from some `Reader` and put it to the store (`loadSnapshot` method).
type snapshotter interface {
	storeSnapshot(SnapshotInfo, io.Writer) error
	loadSnapshot(SnapshotInfo, io.Reader) error
}

type Downloader interface {
	io.Reader
	io.Closer
	GetLength() uint64
}
