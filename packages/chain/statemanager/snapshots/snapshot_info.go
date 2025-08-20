package snapshots

import (
	"fmt"

	"github.com/iotaledger/wasp/v2/packages/state"
	"github.com/iotaledger/wasp/v2/packages/trie"
)

type snapshotInfoImpl struct {
	index      uint32
	commitment *state.L1Commitment
}

var _ SnapshotInfo = &snapshotInfoImpl{}

func NewSnapshotInfo(index uint32, commitment *state.L1Commitment) SnapshotInfo {
	return &snapshotInfoImpl{
		index:      index,
		commitment: commitment,
	}
}

func (si *snapshotInfoImpl) StateIndex() uint32 {
	return si.index
}

func (si *snapshotInfoImpl) Commitment() *state.L1Commitment {
	return si.commitment
}

func (si *snapshotInfoImpl) TrieRoot() trie.Hash {
	return si.Commitment().TrieRoot()
}

func (si *snapshotInfoImpl) BlockHash() state.BlockHash {
	return si.Commitment().BlockHash()
}

func (si *snapshotInfoImpl) String() string {
	return fmt.Sprintf("%v %s", si.StateIndex(), si.Commitment())
}

func (si *snapshotInfoImpl) Equals(other SnapshotInfo) bool {
	if si == nil {
		return other == nil
	}
	if si.StateIndex() != other.StateIndex() {
		return false
	}
	return si.Commitment().Equals(other.Commitment())
}
