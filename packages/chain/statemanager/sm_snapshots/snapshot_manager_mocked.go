package sm_snapshots

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/kvstore/mapdb"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/state"
)

type mockedSnapshotManager struct {
	*snapshotManagerRunner

	t            *testing.T
	createPeriod uint32

	availableSnapshots      map[uint32]SliceStruct[*state.L1Commitment]
	availableSnapshotsMutex sync.RWMutex
	readySnapshots          map[uint32]SliceStruct[*state.L1Commitment]
	readySnapshotsMutex     sync.Mutex

	origStore state.Store
	nodeStore state.Store

	snapshotCommitTime time.Duration
	snapshotLoadTime   time.Duration
}

var (
	_ snapshotManagerCore = &mockedSnapshotManager{}
	_ SnapshotManager     = &mockedSnapshotManager{}
	_ SnapshotManagerTest = &mockedSnapshotManager{}
)

func NewMockedSnapshotManager(
	t *testing.T,
	createPeriod uint32,
	origStore state.Store,
	nodeStore state.Store,
	snapshotCommitTime time.Duration,
	snapshotLoadTime time.Duration,
	log *logger.Logger,
) SnapshotManagerTest {
	result := &mockedSnapshotManager{
		t:                       t,
		createPeriod:            createPeriod,
		availableSnapshots:      make(map[uint32]SliceStruct[*state.L1Commitment]),
		availableSnapshotsMutex: sync.RWMutex{},
		readySnapshots:          make(map[uint32]SliceStruct[*state.L1Commitment]),
		readySnapshotsMutex:     sync.Mutex{},
		origStore:               origStore,
		nodeStore:               nodeStore,
		snapshotCommitTime:      snapshotCommitTime,
		snapshotLoadTime:        snapshotLoadTime,
	}
	result.snapshotManagerRunner = newSnapshotManagerRunner(context.Background(), nil, result, log)
	return result
}

// -------------------------------------
// Implementations of SnapshotManager interface
// -------------------------------------

func (msmT *mockedSnapshotManager) SnapshotExists(stateIndex uint32, commitment *state.L1Commitment) bool {
	msmT.availableSnapshotsMutex.RLock()
	defer msmT.availableSnapshotsMutex.RUnlock()

	commitments, ok := msmT.availableSnapshots[stateIndex]
	if !ok {
		return false
	}
	return commitments.ContainsBy(func(comm *state.L1Commitment) bool { return comm.Equals(commitment) })
}

// NOTE: other implementations are inherited from snapshotManagerRunner

// -------------------------------------
// Implementations of SnapshotManagerTest interface
// -------------------------------------

func (msmT *mockedSnapshotManager) SnapshotReady(snapshotInfo SnapshotInfo) {
	msmT.readySnapshotsMutex.Lock()
	defer msmT.readySnapshotsMutex.Unlock()

	commitments, ok := msmT.readySnapshots[snapshotInfo.GetStateIndex()]
	if ok {
		commitments.Add(snapshotInfo.GetCommitment())
	} else {
		msmT.readySnapshots[snapshotInfo.GetStateIndex()] = NewSliceStruct(snapshotInfo.GetCommitment())
	}
}

// -------------------------------------
// Implementations of snapshotManagerCore interface
// -------------------------------------

func (msmT *mockedSnapshotManager) createSnapshotsNeeded() bool {
	return msmT.createPeriod > 0
}

func (msmT *mockedSnapshotManager) handleUpdate() {
	msmT.readySnapshotsMutex.Lock()
	defer msmT.readySnapshotsMutex.Unlock()
	availableSnapshots := make(map[uint32]SliceStruct[*state.L1Commitment])
	for index, commitments := range msmT.readySnapshots {
		availableSnapshots[index] = commitments.Clone()
	}

	msmT.availableSnapshotsMutex.Lock()
	defer msmT.availableSnapshotsMutex.Unlock()
	msmT.availableSnapshots = availableSnapshots
}

func (msmT *mockedSnapshotManager) handleBlockCommitted(snapshotInfo SnapshotInfo) {
	stateIndex := snapshotInfo.GetStateIndex()
	if stateIndex%msmT.createPeriod == 0 {
		go func() {
			time.Sleep(msmT.snapshotCommitTime)
			msmT.SnapshotReady(snapshotInfo)
		}()
	}
}

func (msmT *mockedSnapshotManager) handleLoadSnapshot(snapshotInfo SnapshotInfo, callback chan<- error) {
	commitments, ok := msmT.readySnapshots[snapshotInfo.GetStateIndex()]
	require.True(msmT.t, ok)
	require.True(msmT.t, commitments.ContainsBy(func(elem *state.L1Commitment) bool {
		return elem.Equals(snapshotInfo.GetCommitment())
	}))
	time.Sleep(msmT.snapshotLoadTime)
	snapshot := mapdb.NewMapDB()
	err := msmT.origStore.TakeSnapshot(snapshotInfo.GetTrieRoot(), snapshot)
	require.NoError(msmT.t, err)
	err = msmT.nodeStore.RestoreSnapshot(snapshotInfo.GetTrieRoot(), snapshot)
	require.NoError(msmT.t, err)
	callback <- nil
}
