package sm_snapshots

import (
	"context"
	"sync"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/shutdown"
	"github.com/iotaledger/wasp/packages/util/pipe"
)

// To avoid code duplication, a common parts of regular and mocked snapshot managers
// are extracted to `snapshotManagerRunner`.
type snapshotManagerRunner struct {
	log                 *logger.Logger
	ctx                 context.Context
	shutdownCoordinator *shutdown.Coordinator

	blockCommittedPipe pipe.Pipe[SnapshotInfo]

	lastIndexSnapshotted      uint32
	lastIndexSnapshottedMutex sync.Mutex
	createPeriod              uint32

	core snapshotManagerCore
}

func newSnapshotManagerRunner(
	ctx context.Context,
	shutdownCoordinator *shutdown.Coordinator,
	createPeriod uint32,
	core snapshotManagerCore,
	log *logger.Logger,
) *snapshotManagerRunner {
	result := &snapshotManagerRunner{
		log:                       log,
		ctx:                       ctx,
		shutdownCoordinator:       shutdownCoordinator,
		blockCommittedPipe:        pipe.NewInfinitePipe[SnapshotInfo](),
		lastIndexSnapshotted:      0,
		lastIndexSnapshottedMutex: sync.Mutex{},
		createPeriod:              createPeriod,
		core:                      core,
	}
	go result.run()
	return result
}

// -------------------------------------
// Implementations of SnapshotManager interface
// -------------------------------------

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
	if stateIndex > smrT.lastIndexSnapshotted {
		smrT.lastIndexSnapshotted = stateIndex
	}
	smrT.lastIndexSnapshottedMutex.Unlock()
}

// -------------------------------------
// Internal functions
// -------------------------------------

func (smrT *snapshotManagerRunner) run() {
	blockCommittedPipeCh := smrT.blockCommittedPipe.Out()
	for {
		if smrT.ctx.Err() != nil {
			if smrT.shutdownCoordinator == nil {
				return
			}
			if smrT.shutdownCoordinator.CheckNestedDone() {
				smrT.log.Debugf("Stopping snapshot manager, because context was closed")
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
	stateIndex := snapshotInfo.StateIndex()
	var lastIndexSnapshotted uint32
	smrT.lastIndexSnapshottedMutex.Lock()
	lastIndexSnapshotted = smrT.lastIndexSnapshotted
	smrT.lastIndexSnapshottedMutex.Unlock()
	if (stateIndex > lastIndexSnapshotted) && (stateIndex%smrT.createPeriod == 0) { // TODO: what if snapshotted state has been reverted?
		smrT.core.createSnapshot(snapshotInfo)
	}
}
