package sm_snapshots

type snapshotManagerEmpty struct{}

var _ SnapshotManager = &snapshotManagerEmpty{}

func NewEmptySnapshotManager() SnapshotManager                    { return &snapshotManagerEmpty{} }
func (*snapshotManagerEmpty) BlockCommittedAsync(SnapshotInfo)    {}
func (*snapshotManagerEmpty) GetLoadedSnapshotStateIndex() uint32 { return 0 }
