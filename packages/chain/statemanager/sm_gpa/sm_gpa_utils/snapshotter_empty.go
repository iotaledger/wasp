package sm_gpa_utils

import (
	"github.com/iotaledger/wasp/packages/state"
)

type emptySnapshotter struct{}

var _ Snapshotter = &emptySnapshotter{}

func NewEmptySnapshotter() Snapshotter                  { return &emptySnapshotter{} }
func (sn *emptySnapshotter) BlockCommitted(state.Block) {}
