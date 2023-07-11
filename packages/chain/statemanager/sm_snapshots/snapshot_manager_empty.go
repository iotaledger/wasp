package sm_snapshots

import (
	"github.com/iotaledger/wasp/packages/state"
)

type snapshotManagerEmpty struct{}

var (
	_ SnapshotManager     = &snapshotManagerEmpty{}
	_ SnapshotManagerTest = &snapshotManagerEmpty{}
)

func NewEmptySnapshotManager() SnapshotManagerTest                            { return &snapshotManagerEmpty{} }
func (*snapshotManagerEmpty) UpdateAsync()                                    {}
func (*snapshotManagerEmpty) BlockCommittedAsync(SnapshotInfo)                {}
func (*snapshotManagerEmpty) SnapshotExists(uint32, *state.L1Commitment) bool { return false }
func (*snapshotManagerEmpty) SnapshotReady(SnapshotInfo)                      {}
func (*snapshotManagerEmpty) IsSnapshotReady(SnapshotInfo) bool               { return false }
func (*snapshotManagerEmpty) SetAfterSnapshotCreated(func(SnapshotInfo))      {}

func (*snapshotManagerEmpty) LoadSnapshotAsync(SnapshotInfo) <-chan error {
	callback := make(chan error, 1)
	callback <- nil
	return callback
}
