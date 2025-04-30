package snapshots

import (
	"bytes"
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/log"

	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/util/timeutil"
)

type MockedSnapshotManager struct {
	*snapshotManagerRunner

	t            *testing.T
	log          log.Logger
	createPeriod uint32

	readySnapshots      map[uint32]*util.SliceStruct[*state.L1Commitment]
	readySnapshotsMutex sync.Mutex

	snapshotCommitTime time.Duration
	timeProvider       timeutil.TimeProvider

	origStore      state.Store
	nodeStore      state.Store
	snapshotToLoad SnapshotInfo

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
	timeProvider timeutil.TimeProvider,
	log log.Logger,
) *MockedSnapshotManager {
	result := &MockedSnapshotManager{
		t:                   t,
		log:                 log.NewChildLogger("MSnap"),
		createPeriod:        createPeriod,
		readySnapshots:      make(map[uint32]*util.SliceStruct[*state.L1Commitment]),
		readySnapshotsMutex: sync.Mutex{},
		snapshotCommitTime:  snapshotCommitTime,
		timeProvider:        timeProvider,
		origStore:           origStore,
		nodeStore:           nodeStore,
		snapshotToLoad:      snapshotToLoad,
	}
	result.snapshotManagerRunner = newSnapshotManagerRunner(context.Background(), nodeStore, nil, createPeriod, delayPeriod, result, result.log)
	return result
}

// -------------------------------------
// Implementations of SnapshotManager interface
// -------------------------------------

// NOTE: implementations are inherited from snapshotManagerRunner

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
	msmT.log.LogDebugf("Creating snapshot %s...", snapshotInfo)
	go func() {
		<-msmT.timeProvider.After(msmT.snapshotCommitTime)
		msmT.snapshotCreatedCount.Add(1)
		msmT.snapshotReady(snapshotInfo)
		msmT.log.LogDebugf("Creating snapshot %s: completed", snapshotInfo)
		msmT.snapshotCreateFinalisedCount.Add(1)
		msmT.snapshotCreated(snapshotInfo)
	}()
}

func (msmT *MockedSnapshotManager) loadSnapshot() SnapshotInfo {
	if msmT.snapshotToLoad == nil {
		return nil
	}
	msmT.log.LogDebugf("Loading snapshot %s...", msmT.snapshotToLoad)
	snapshot := new(bytes.Buffer)
	err := msmT.origStore.TakeSnapshot(msmT.snapshotToLoad.TrieRoot(), snapshot)
	require.NoError(msmT.t, err)
	err = msmT.nodeStore.RestoreSnapshot(msmT.snapshotToLoad.TrieRoot(), snapshot)
	require.NoError(msmT.t, err)
	msmT.log.LogDebugf("Loading snapshot %s: snapshot loaded", msmT.snapshotToLoad)
	return msmT.snapshotToLoad
}

// -------------------------------------
// Internal functions
// -------------------------------------

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
