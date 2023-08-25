package sm_snapshots

import (
	"context"

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
		blockCommittedPipe:  pipe.NewInfinitePipe[SnapshotInfo](),
		core:                core,
	}
	go result.run()
	return result
}

// -------------------------------------
// Implementations of SnapshotManager interface
// -------------------------------------

func (smrT *snapshotManagerRunner) BlockCommittedAsync(snapshotInfo SnapshotInfo) {
	if smrT.core.createSnapshotsNeeded() {
		smrT.blockCommittedPipe.In() <- snapshotInfo
	}
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
				smrT.core.handleBlockCommitted(snapshotInfo)
			} else {
				blockCommittedPipeCh = nil
			}
		case <-smrT.ctx.Done():
			continue
		}
	}
}
