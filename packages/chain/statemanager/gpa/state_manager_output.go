package gpa

import (
	"github.com/iotaledger/wasp/packages/chain/statemanager/gpa/sm_inputs"
	"github.com/iotaledger/wasp/packages/chain/statemanager/snapshots"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/state"
)

type smOutputImpl struct {
	blocksCommitted []snapshots.SnapshotInfo
	nextInputs      []gpa.Input
}

var (
	_ gpa.Output         = &smOutputImpl{}
	_ StateManagerOutput = &smOutputImpl{}
)

func newOutput() StateManagerOutput {
	return &smOutputImpl{
		blocksCommitted: make([]snapshots.SnapshotInfo, 0),
		nextInputs:      make([]gpa.Input, 0),
	}
}

func (smoi *smOutputImpl) addBlockCommitted(stateIndex uint32, commitment *state.L1Commitment) {
	smoi.blocksCommitted = append(smoi.blocksCommitted, snapshots.NewSnapshotInfo(stateIndex, commitment))
}

func (smoi *smOutputImpl) TakeBlocksCommitted() []snapshots.SnapshotInfo {
	result := smoi.blocksCommitted
	smoi.blocksCommitted = make([]snapshots.SnapshotInfo, 0)
	return result
}

func (smoi *smOutputImpl) addBlocksToCommit(commitments []*state.L1Commitment) {
	smoi.nextInputs = append(smoi.nextInputs, sm_inputs.NewStateManagerBlocksToCommit(commitments))
}

func (smoi *smOutputImpl) TakeNextInputs() []gpa.Input {
	result := smoi.nextInputs
	smoi.nextInputs = make([]gpa.Input, 0)
	return result
}
