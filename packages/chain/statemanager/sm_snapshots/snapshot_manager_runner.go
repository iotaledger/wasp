package sm_snapshots

import (
	"context"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/shutdown"
	"github.com/iotaledger/wasp/packages/util/pipe"
)

type snapshotInfoCallback struct {
	SnapshotInfo
	callback chan<- error
}

// To avoid code duplication, a common parts of regular and mocked snapshot managers
// are extracted to `snapshotManagerRunner`.
type snapshotManagerRunner struct {
	log                 *logger.Logger
	ctx                 context.Context
	shutdownCoordinator *shutdown.Coordinator

	updatePipe         pipe.Pipe[bool]
	blockCommittedPipe pipe.Pipe[SnapshotInfo]
	loadSnapshotPipe   pipe.Pipe[*snapshotInfoCallback]

	core snapshotManagerCore
}

func newSnapshotManagerRunner(
	ctx context.Context,
	shutdownCoordinator *shutdown.Coordinator,
	core snapshotManagerCore,
	log *logger.Logger,
) *snapshotManagerRunner {
	result := &snapshotManagerRunner{
		log:                 log,
		ctx:                 ctx,
		shutdownCoordinator: shutdownCoordinator,
		updatePipe:          pipe.NewInfinitePipe[bool](),
		blockCommittedPipe:  pipe.NewInfinitePipe[SnapshotInfo](),
		loadSnapshotPipe:    pipe.NewInfinitePipe[*snapshotInfoCallback](),
		core:                core,
	}
	go result.run()
	return result
}

// -------------------------------------
// Implementations of SnapshotManager interface
// -------------------------------------

func (smrT *snapshotManagerRunner) UpdateAsync() {
	smrT.updatePipe.In() <- true
}

func (smrT *snapshotManagerRunner) BlockCommittedAsync(snapshotInfo SnapshotInfo) {
	if smrT.core.createSnapshotsNeeded() {
		smrT.blockCommittedPipe.In() <- snapshotInfo
	}
}

func (smrT *snapshotManagerRunner) LoadSnapshotAsync(snapshotInfo SnapshotInfo) <-chan error {
	callback := make(chan error, 1)
	smrT.loadSnapshotPipe.In() <- &snapshotInfoCallback{
		SnapshotInfo: snapshotInfo,
		callback:     callback,
	}
	return callback
}

// -------------------------------------
// Internal functions
// -------------------------------------

func (smrT *snapshotManagerRunner) run() {
	updatePipeCh := smrT.updatePipe.Out()
	blockCommittedPipeCh := smrT.blockCommittedPipe.Out()
	loadSnapshotPipeCh := smrT.loadSnapshotPipe.Out()
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
		case _, ok := <-updatePipeCh:
			if ok {
				smrT.core.handleUpdate()
			} else {
				updatePipeCh = nil
			}
		case snapshotInfo, ok := <-blockCommittedPipeCh:
			if ok {
				smrT.core.handleBlockCommitted(snapshotInfo)
			} else {
				blockCommittedPipeCh = nil
			}
		case snapshotInfoC, ok := <-loadSnapshotPipeCh:
			if ok {
				func() {
					defer close(snapshotInfoC.callback)
					smrT.core.handleLoadSnapshot(snapshotInfoC.SnapshotInfo, snapshotInfoC.callback)
				}()
			} else {
				loadSnapshotPipeCh = nil
			}
		case <-smrT.ctx.Done():
			continue
		}
	}
}
