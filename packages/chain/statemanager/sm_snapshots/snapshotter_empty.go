package sm_snapshots

import (
	"io"
)

type emptySnapshotter struct{}

var _ snapshotter = &emptySnapshotter{}

func newEmptySnapshotter() snapshotter                                   { return &emptySnapshotter{} }
func (sn *emptySnapshotter) storeSnapshot(SnapshotInfo, io.Writer) error { return nil }
func (sn *emptySnapshotter) loadSnapshot(SnapshotInfo, io.Reader) error  { return nil }
