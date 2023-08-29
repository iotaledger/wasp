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
	log          *logger.Logger
	createPeriod uint32

	readySnapshots      map[uint32]*util.SliceStruct[*state.L1Commitment]
	readySnapshotsMutex sync.Mutex

	snapshotCommitTime       time.Duration
	timeProvider             sm_gpa_utils.TimeProvider
	afterSnapshotCreatedFun  func(SnapshotInfo)
	loadedSnapshotStateIndex uint32

	snapshotCreateRequestCount   atomic.Uint32
	snapshotCreatedCount         atomic.Uint32
	snapshotCreateFinalisedCount atomic.Uint32
}

var (
	_ snapshotManagerCore = &MockedSnapshotManager{}
	_ SnapshotManager     = &MockedSnapshotManager{}
)

func NewMockedSnapshotManager(
	t *testing.T,
	createPeriod uint32,
	delayPeriod uint32,
	origStore state.Store,
	nodeStore state.Store,
	snapshotToLoad SnapshotInfo,
	snapshotCommitTime time.Duration,
	timeProvider sm_gpa_utils.TimeProvider,
	log *logger.Logger,
) *MockedSnapshotManager {
	result := &MockedSnapshotManager{
		t:                        t,
		log:                      log.Named("MSnap"),
		createPeriod:             createPeriod,
		readySnapshots:           make(map[uint32]*util.SliceStruct[*state.L1Commitment]),
		readySnapshotsMutex:      sync.Mutex{},
		snapshotCommitTime:       snapshotCommitTime,
		timeProvider:             timeProvider,
		afterSnapshotCreatedFun:  func(SnapshotInfo) {},
		loadedSnapshotStateIndex: 0,
	}
	if nodeStore.IsEmpty() && snapshotToLoad != nil {
		result.loadSnapshot(origStore, nodeStore, snapshotToLoad)
	}
	result.snapshotManagerRunner = newSnapshotManagerRunner(context.Background(), nil, createPeriod, delayPeriod, result, result.log)
	return result
}

// -------------------------------------
// Implementations of SnapshotManager interface
// -------------------------------------

func (msmT *MockedSnapshotManager) GetLoadedSnapshotStateIndex() uint32 {
	return msmT.loadedSnapshotStateIndex
}

// NOTE: other implementation are inherited from snapshotManagerRunner

// -------------------------------------
// Additional API functions of MockedSnapshotManager
// -------------------------------------

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

func (msmT *MockedSnapshotManager) WaitSnapshotCreateRequestCount(count uint32, sleepTime time.Duration, maxSleepCount int) bool {
	return wait(func() bool { return msmT.snapshotCreateRequestCount.Load() == count }, sleepTime, maxSleepCount)
}

func (msmT *MockedSnapshotManager) WaitSnapshotCreatedCount(count uint32, sleepTime time.Duration, maxSleepCount int) bool {
	return wait(func() bool { return msmT.snapshotCreatedCount.Load() == count }, sleepTime, maxSleepCount)
}

func (msmT *MockedSnapshotManager) WaitSnapshotCreateFinalisedCount(count uint32, sleepTime time.Duration, maxSleepCount int) bool {
	return wait(func() bool { return msmT.snapshotCreateFinalisedCount.Load() == count }, sleepTime, maxSleepCount)
}

// -------------------------------------
// Implementations of snapshotManagerCore interface
// -------------------------------------

func (msmT *MockedSnapshotManager) createSnapshot(snapshotInfo SnapshotInfo) {
	msmT.snapshotCreateRequestCount.Add(1)
	msmT.log.Debugf("Creating snapshot %s...", snapshotInfo)
	go func() {
		<-msmT.timeProvider.After(msmT.snapshotCommitTime)
		msmT.snapshotCreatedCount.Add(1)
		msmT.snapshotReady(snapshotInfo)
		msmT.afterSnapshotCreatedFun(snapshotInfo)
		msmT.log.Debugf("Creating snapshot %s: completed", snapshotInfo)
		msmT.snapshotCreateFinalisedCount.Add(1)
		msmT.snapshotManagerRunner.snapshotCreated(snapshotInfo)
	}()
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

// -------------------------------------
// Internal functions
// -------------------------------------

func (msmT *MockedSnapshotManager) loadSnapshot(origStore state.Store, nodeStore state.Store, snapshotInfo SnapshotInfo) {
	msmT.log.Debugf("Loading snapshot %s...", snapshotInfo)
	snapshot := new(bytes.Buffer)
	err := origStore.TakeSnapshot(snapshotInfo.TrieRoot(), snapshot)
	require.NoError(msmT.t, err)
	err = nodeStore.RestoreSnapshot(snapshotInfo.TrieRoot(), snapshot)
	require.NoError(msmT.t, err)
	msmT.loadedSnapshotStateIndex = snapshotInfo.StateIndex()
	msmT.log.Debugf("Loading snapshot %s: snapshot loaded", snapshotInfo)
}

func (msmT *MockedSnapshotManager) snapshotReady(snapshotInfo SnapshotInfo) {
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
