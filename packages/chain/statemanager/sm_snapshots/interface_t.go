package sm_snapshots

type SnapshotManagerTest interface {
	SnapshotManager
	SnapshotReady(SnapshotInfo)
	IsSnapshotReady(SnapshotInfo) bool
	SetAfterSnapshotCreated(func(SnapshotInfo))
}
