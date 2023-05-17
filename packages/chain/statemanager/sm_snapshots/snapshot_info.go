package sm_snapshots

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/trie"
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

func (si *snapshotInfoImpl) GetStateIndex() uint32 {
	return si.index
}

func (si *snapshotInfoImpl) GetCommitment() *state.L1Commitment {
	return si.commitment
}

func (si *snapshotInfoImpl) GetTrieRoot() trie.Hash {
	return si.GetCommitment().TrieRoot()
}

func (si *snapshotInfoImpl) GetBlockHash() state.BlockHash {
	return si.GetCommitment().BlockHash()
}

func (si *snapshotInfoImpl) String() string {
	return fmt.Sprintf("%v:%s", si.GetStateIndex(), si.GetCommitment())
}

func (si *snapshotInfoImpl) Equals(other SnapshotInfo) bool {
	if si == nil {
		return other == nil
	}
	if si.GetStateIndex() != other.GetStateIndex() {
		return false
	}
	return si.GetCommitment().Equals(other.GetCommitment())
}
