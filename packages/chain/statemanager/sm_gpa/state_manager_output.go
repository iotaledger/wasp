package sm_gpa

import (
	"github.com/iotaledger/wasp/packages/chain/statemanager/sm_snapshots"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/state"
)

type smOutputImpl struct {
	blocksCommitted []sm_snapshots.SnapshotInfo
	snapshotsToLoad []sm_snapshots.SnapshotInfo
	updateSnapshots bool
}

var (
	_ gpa.Output         = &smOutputImpl{}
	_ StateManagerOutput = &smOutputImpl{}
)

func newOutput() StateManagerOutput {
	return &smOutputImpl{
		blocksCommitted: make([]sm_snapshots.SnapshotInfo, 0),
		snapshotsToLoad: make([]sm_snapshots.SnapshotInfo, 0, 1),
	}
}

func (smoi *smOutputImpl) addBlockCommitted(stateIndex uint32, commitment *state.L1Commitment) {
	smoi.blocksCommitted = append(smoi.blocksCommitted, sm_snapshots.NewSnapshotInfo(stateIndex, commitment))
}

func (smoi *smOutputImpl) addSnapshotToLoad(stateIndex uint32, commitment *state.L1Commitment) {
	smoi.snapshotsToLoad = append(smoi.snapshotsToLoad, sm_snapshots.NewSnapshotInfo(stateIndex, commitment))
}

func (smoi *smOutputImpl) setUpdateSnapshots() {
	smoi.updateSnapshots = true
}

func (smoi *smOutputImpl) TakeBlocksCommitted() []sm_snapshots.SnapshotInfo {
	result := smoi.blocksCommitted
	smoi.blocksCommitted = make([]sm_snapshots.SnapshotInfo, 0)
	return result
}

func (smoi *smOutputImpl) TakeSnapshotToLoad() sm_snapshots.SnapshotInfo {
	if len(smoi.snapshotsToLoad) == 0 {
		return nil
	}
	result := smoi.snapshotsToLoad[0]
	smoi.snapshotsToLoad = smoi.snapshotsToLoad[1:]
	return result
}

func (smoi *smOutputImpl) TakeUpdateSnapshots() bool {
	if smoi.updateSnapshots {
		smoi.updateSnapshots = false
		return true
	}
	return false
}
