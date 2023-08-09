package sm_snapshots

import (
	"bytes"
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/chain/statemanager/sm_gpa/sm_gpa_utils"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util"
)

type MockedSnapshotManager struct {
	*snapshotManagerRunner

	t            *testing.T
	createPeriod uint32

	availableSnapshots      map[uint32]*util.SliceStruct[*state.L1Commitment]
	availableSnapshotsMutex sync.RWMutex
	readySnapshots          map[uint32]*util.SliceStruct[*state.L1Commitment]
	readySnapshotsMutex     sync.Mutex

	origStore state.Store
	nodeStore state.Store

	snapshotCommitTime      time.Duration
	snapshotLoadTime        time.Duration
	timeProvider            sm_gpa_utils.TimeProvider
	afterSnapshotCreatedFun func(SnapshotInfo)

	updateCount                  atomic.Uint32
	snapshotCreateRequestCount   atomic.Uint32
	snapshotCreatedCount         atomic.Uint32
	snapshotCreateFinalisedCount atomic.Uint32
	snapshotLoadRequestCount     atomic.Uint32
	snapshotLoadedCount          atomic.Uint32
}

var (
	_ snapshotManagerCore = &MockedSnapshotManager{}
	_ SnapshotManager     = &MockedSnapshotManager{}
)

func NewMockedSnapshotManager(
	t *testing.T,
	createPeriod uint32,
	origStore state.Store,
	nodeStore state.Store,
	snapshotCommitTime time.Duration,
	snapshotLoadTime time.Duration,
	timeProvider sm_gpa_utils.TimeProvider,
	log *logger.Logger,
) *MockedSnapshotManager {
	result := &MockedSnapshotManager{
		t:                       t,
		createPeriod:            createPeriod,
		availableSnapshots:      make(map[uint32]*util.SliceStruct[*state.L1Commitment]),
		availableSnapshotsMutex: sync.RWMutex{},
		readySnapshots:          make(map[uint32]*util.SliceStruct[*state.L1Commitment]),
		readySnapshotsMutex:     sync.Mutex{},
		origStore:               origStore,
		nodeStore:               nodeStore,
		snapshotCommitTime:      snapshotCommitTime,
		snapshotLoadTime:        snapshotLoadTime,
		timeProvider:            timeProvider,
		afterSnapshotCreatedFun: func(SnapshotInfo) {},
	}
	result.snapshotManagerRunner = newSnapshotManagerRunner(context.Background(), nil, result, log.Named("MSnap"))
	return result
}

// -------------------------------------
// Implementations of SnapshotManager interface
// -------------------------------------

func (msmT *MockedSnapshotManager) SnapshotExists(stateIndex uint32, commitment *state.L1Commitment) bool {
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
// Additional API functions of MockedSnapshotManager
// -------------------------------------

func (msmT *MockedSnapshotManager) SnapshotReady(snapshotInfo SnapshotInfo) {
	msmT.readySnapshotsMutex.Lock()
	defer msmT.readySnapshotsMutex.Unlock()

	commitments, ok := msmT.readySnapshots[snapshotInfo.StateIndex()]
	if ok {
		if !commitments.ContainsBy(func(comm *state.L1Commitment) bool { return comm.Equals(snapshotInfo.Commitment()) }) {
			commitments.Add(snapshotInfo.Commitment())
		}
	} else {
		msmT.readySnapshots[snapshotInfo.StateIndex()] = util.NewSliceStruct(snapshotInfo.Commitment())
	}
}

func (msmT *MockedSnapshotManager) IsSnapshotReady(snapshotInfo SnapshotInfo) bool {
	msmT.readySnapshotsMutex.Lock()
	defer msmT.readySnapshotsMutex.Unlock()

	commitments, ok := msmT.readySnapshots[snapshotInfo.StateIndex()]
	if !ok {
		return false
	}
	return commitments.ContainsBy(func(elem *state.L1Commitment) bool { return elem.Equals(snapshotInfo.Commitment()) })
}

func (msmT *MockedSnapshotManager) SetAfterSnapshotCreated(fun func(SnapshotInfo)) {
	msmT.afterSnapshotCreatedFun = fun
}

func (msmT *MockedSnapshotManager) WaitNodeUpdateCount(count uint32, sleepTime time.Duration, maxSleepCount int) bool {
	return wait(func() bool { return msmT.updateCount.Load() == count }, sleepTime, maxSleepCount)
}

func (msmT *MockedSnapshotManager) WaitSnapshotCreateRequestCount(count uint32, sleepTime time.Duration, maxSleepCount int) bool {
	return wait(func() bool { return msmT.snapshotCreateRequestCount.Load() == count }, sleepTime, maxSleepCount)
}

func (msmT *MockedSnapshotManager) WaitSnapshotCreatedCount(count uint32, sleepTime time.Duration, maxSleepCount int) bool {
	return wait(func() bool { return msmT.snapshotCreatedCount.Load() == count }, sleepTime, maxSleepCount)
}

func (msmT *MockedSnapshotManager) WaitSnapshotCreateFinalisedCount(count uint32, sleepTime time.Duration, maxSleepCount int) bool {
	return wait(func() bool { return msmT.snapshotCreateFinalisedCount.Load() == count }, sleepTime, maxSleepCount)
}

func (msmT *MockedSnapshotManager) WaitSnapshotLoadRequestCount(count uint32, sleepTime time.Duration, maxSleepCount int) bool {
	return wait(func() bool { return msmT.snapshotLoadRequestCount.Load() == count }, sleepTime, maxSleepCount)
}

func (msmT *MockedSnapshotManager) WaitSnapshotLoadedCount(count uint32, sleepTime time.Duration, maxSleepCount int) bool {
	return wait(func() bool { return msmT.snapshotLoadedCount.Load() == count }, sleepTime, maxSleepCount)
}

// -------------------------------------
// Implementations of snapshotManagerCore interface
// -------------------------------------

func (msmT *MockedSnapshotManager) createSnapshotsNeeded() bool {
	return msmT.createPeriod > 0
}

func (msmT *MockedSnapshotManager) handleUpdate() {
	msmT.readySnapshotsMutex.Lock()
	defer msmT.readySnapshotsMutex.Unlock()

	availableSnapshots := make(map[uint32]*util.SliceStruct[*state.L1Commitment])
	count := 0
	for index, commitments := range msmT.readySnapshots {
		clonedCommitments := commitments.Clone()
		availableSnapshots[index] = clonedCommitments
		count += clonedCommitments.Length()
	}
	msmT.log.Debugf("Update: %v snapshots found", count)

	msmT.availableSnapshotsMutex.Lock()
	defer msmT.availableSnapshotsMutex.Unlock()
	msmT.availableSnapshots = availableSnapshots
	msmT.updateCount.Add(1)
}

func (msmT *MockedSnapshotManager) handleBlockCommitted(snapshotInfo SnapshotInfo) {
	stateIndex := snapshotInfo.StateIndex()
	if stateIndex%msmT.createPeriod == 0 {
		msmT.snapshotCreateRequestCount.Add(1)
		msmT.log.Debugf("Creating snapshot %s...", snapshotInfo)
		go func() {
			<-msmT.timeProvider.After(msmT.snapshotCommitTime)
			msmT.snapshotCreatedCount.Add(1)
			msmT.SnapshotReady(snapshotInfo)
			msmT.afterSnapshotCreatedFun(snapshotInfo)
			msmT.log.Debugf("Creating snapshot %s: completed", snapshotInfo)
			msmT.snapshotCreateFinalisedCount.Add(1)
		}()
	}
}

func (msmT *MockedSnapshotManager) handleLoadSnapshot(snapshotInfo SnapshotInfo, callback chan<- error) {
	msmT.snapshotLoadRequestCount.Add(1)
	msmT.log.Debugf("Loading snapshot %s...", snapshotInfo)
	commitments, ok := msmT.availableSnapshots[snapshotInfo.StateIndex()]
	require.True(msmT.t, ok)
	require.True(msmT.t, commitments.ContainsBy(func(elem *state.L1Commitment) bool {
		return elem.Equals(snapshotInfo.Commitment())
	}))
	<-msmT.timeProvider.After(msmT.snapshotLoadTime)
	snapshot := new(bytes.Buffer)
	err := msmT.origStore.TakeSnapshot(snapshotInfo.TrieRoot(), snapshot)
	require.NoError(msmT.t, err)
	err = msmT.nodeStore.RestoreSnapshot(snapshotInfo.TrieRoot(), snapshot)
	require.NoError(msmT.t, err)
	callback <- nil
	msmT.log.Debugf("Loading snapshot %s: snapshot loaded", snapshotInfo)
	msmT.snapshotLoadedCount.Add(1)
}

func wait(predicateFun func() bool, sleepTime time.Duration, maxSleepCount int) bool {
	if predicateFun() {
		return true
	}
	for i := 0; i < maxSleepCount; i++ {
		time.Sleep(sleepTime)
		if predicateFun() {
			return true
		}
	}
	return false
}
