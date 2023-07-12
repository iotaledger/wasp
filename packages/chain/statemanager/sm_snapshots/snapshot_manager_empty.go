package sm_snapshots

import (
	"github.com/iotaledger/wasp/packages/state"
)

type snapshotManagerEmpty struct{}

var _ SnapshotManager = &snapshotManagerEmpty{}

func NewEmptySnapshotManager() SnapshotManager                                { return &snapshotManagerEmpty{} }
func (*snapshotManagerEmpty) UpdateAsync()                                    {}
func (*snapshotManagerEmpty) BlockCommittedAsync(SnapshotInfo)                {}
func (*snapshotManagerEmpty) SnapshotExists(uint32, *state.L1Commitment) bool { return false }
func (*snapshotManagerEmpty) LoadSnapshotAsync(SnapshotInfo) <-chan error {
	callback := make(chan error, 1)
	callback <- nil
	return callback
}
