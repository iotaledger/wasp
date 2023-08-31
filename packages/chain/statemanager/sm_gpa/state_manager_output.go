package sm_gpa

import (
	"github.com/iotaledger/wasp/packages/chain/statemanager/sm_snapshots"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/state"
)

type smOutputImpl struct {
	blocksCommitted []sm_snapshots.SnapshotInfo
}

var (
	_ gpa.Output         = &smOutputImpl{}
	_ StateManagerOutput = &smOutputImpl{}
)

func newOutput() StateManagerOutput {
	return &smOutputImpl{
		blocksCommitted: make([]sm_snapshots.SnapshotInfo, 0),
	}
}

func (smoi *smOutputImpl) addBlockCommitted(stateIndex uint32, commitment *state.L1Commitment) {
	smoi.blocksCommitted = append(smoi.blocksCommitted, sm_snapshots.NewSnapshotInfo(stateIndex, commitment))
}

func (smoi *smOutputImpl) TakeBlocksCommitted() []sm_snapshots.SnapshotInfo {
	result := smoi.blocksCommitted
	smoi.blocksCommitted = make([]sm_snapshots.SnapshotInfo, 0)
	return result
}
