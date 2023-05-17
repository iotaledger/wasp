package sm_snapshots

import (
	"io"

	"github.com/iotaledger/wasp/packages/state"
)

type emptySnapshotter struct{}

var _ snapshotter = &emptySnapshotter{}

func newEmptySnapshotter() snapshotter                                               { return &emptySnapshotter{} }
func (sn *emptySnapshotter) createSnapshotAsync(uint32, *state.L1Commitment, func()) {}
func (sn *emptySnapshotter) loadSnapshot(io.Reader) error                            { return nil }
