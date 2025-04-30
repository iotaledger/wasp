package snapshots

import (
	"context"
	"sync"

	"github.com/samber/lo"

	"github.com/iotaledger/hive.go/log"

	"github.com/iotaledger/wasp/packages/shutdown"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util/pipe"
)

// To avoid code duplication, a common parts of regular and mocked snapshot managers
// are extracted to `snapshotManagerRunner`.
type snapshotManagerRunner struct {
	log                 log.Logger
	ctx                 context.Context
	shutdownCoordinator *shutdown.Coordinator

	blockCommittedPipe pipe.Pipe[SnapshotInfo]

	lastIndexSnapshotted      uint32
	lastIndexSnapshottedMutex sync.Mutex
	loadedSnapshotStateIndex  uint32
	createPeriod              uint32
	delayPeriod               uint32
	queue                     []SnapshotInfo

	core snapshotManagerCore
}

func newSnapshotManagerRunner(
	ctx context.Context,
	store state.Store,
	shutdownCoordinator *shutdown.Coordinator,
	createPeriod uint32,
	delayPeriod uint32,
	core snapshotManagerCore,
	log log.Logger,
) *snapshotManagerRunner {
	result := &snapshotManagerRunner{
		log:                       log,
		ctx:                       ctx,
		shutdownCoordinator:       shutdownCoordinator,
		blockCommittedPipe:        pipe.NewInfinitePipe[SnapshotInfo](),
		lastIndexSnapshotted:      0,
		lastIndexSnapshottedMutex: sync.Mutex{},
		loadedSnapshotStateIndex:  0,
		createPeriod:              createPeriod,
		delayPeriod:               delayPeriod,
		queue:                     make([]SnapshotInfo, 0),
		core:                      core,
	}
	if store.IsEmpty() {
		loadedSnapshotInfo := result.core.loadSnapshot()
		if loadedSnapshotInfo != nil {
			result.loadedSnapshotStateIndex = loadedSnapshotInfo.StateIndex()
		}
	}
	go result.run()
	return result
}

// -------------------------------------
// Implementations of SnapshotManager interface
// -------------------------------------

func (smrT *snapshotManagerRunner) GetLoadedSnapshotStateIndex() uint32 {
	return smrT.loadedSnapshotStateIndex
}

func (smrT *snapshotManagerRunner) BlockCommittedAsync(snapshotInfo SnapshotInfo) {
	if smrT.createSnapshotsNeeded() {
		smrT.blockCommittedPipe.In() <- snapshotInfo
	}
}

// -------------------------------------
// Api for snapshotManagerCore implementations
// -------------------------------------

func (smrT *snapshotManagerRunner) snapshotCreated(snapshotInfo SnapshotInfo) {
	stateIndex := snapshotInfo.StateIndex()
	smrT.lastIndexSnapshottedMutex.Lock()
	defer smrT.lastIndexSnapshottedMutex.Unlock()
	if stateIndex > smrT.lastIndexSnapshotted {
		smrT.lastIndexSnapshotted = stateIndex
		smrT.queue = lo.Filter(smrT.queue, func(si SnapshotInfo, index int) bool { return si.StateIndex() > smrT.lastIndexSnapshotted })
	}
}

// -------------------------------------
// Internal functions
// -------------------------------------

func (smrT *snapshotManagerRunner) run() {
	defer smrT.blockCommittedPipe.Close()
	blockCommittedPipeCh := smrT.blockCommittedPipe.Out()
	for {
		if smrT.ctx.Err() != nil {
			if smrT.shutdownCoordinator == nil {
				return
			}
			if smrT.shutdownCoordinator.CheckNestedDone() {
				smrT.log.LogDebugf("Stopping snapshot manager, because context was closed")
				smrT.shutdownCoordinator.Done()
				return
			}
		}
		select {
		case snapshotInfo, ok := <-blockCommittedPipeCh:
			if ok {
				smrT.handleBlockCommitted(snapshotInfo)
			} else {
				blockCommittedPipeCh = nil
			}
		case <-smrT.ctx.Done():
			continue
		}
	}
}

func (smrT *snapshotManagerRunner) createSnapshotsNeeded() bool {
	return smrT.createPeriod > 0
}

func (smrT *snapshotManagerRunner) handleBlockCommitted(snapshotInfo SnapshotInfo) {
	sisToCreate := func() []SnapshotInfo { // Function to unlock the mutex quicker
		stateIndex := snapshotInfo.StateIndex()
		var lastIndexSnapshotted uint32
		smrT.lastIndexSnapshottedMutex.Lock()
		defer smrT.lastIndexSnapshottedMutex.Unlock()
		lastIndexSnapshotted = smrT.lastIndexSnapshotted
		if (stateIndex > lastIndexSnapshotted) && (stateIndex%smrT.createPeriod == 0) { // TODO: what if snapshotted state has been reverted?
			smrT.queue = append(smrT.queue, snapshotInfo)
		}
		stateIndexToCommit := stateIndex - smrT.delayPeriod
		if (stateIndexToCommit > lastIndexSnapshotted) && (stateIndexToCommit%smrT.createPeriod == 0) {
			return lo.Filter(smrT.queue, func(si SnapshotInfo, index int) bool { return si.StateIndex() == stateIndexToCommit })
		}
		return []SnapshotInfo{}
	}()
	for i, siToCreate := range sisToCreate {
		if !(lo.ContainsBy(sisToCreate[:i], func(si SnapshotInfo) bool { return si.Equals(siToCreate) })) {
			smrT.core.createSnapshot(siToCreate)
		}
	}
}
