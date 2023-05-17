package sm_snapshots

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/samber/lo"

	"github.com/iotaledger/hive.go/ds/shrinkingmap"
	"github.com/iotaledger/hive.go/logger"
	//	consGR "github.com/iotaledger/wasp/packages/chain/cons/cons_gr"
	//	"github.com/iotaledger/wasp/packages/chain/statemanager/sm_gpa/sm_gpa_utils"
	//	"github.com/iotaledger/wasp/packages/chain/statemanager/sm_gpa/sm_inputs"
	//	"github.com/iotaledger/wasp/packages/chain/statemanager/sm_utils"
	//	"github.com/iotaledger/wasp/packages/cryptolib"
	//	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
	//	"github.com/iotaledger/wasp/packages/metrics"
	//	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/shutdown"
	"github.com/iotaledger/wasp/packages/state"
	//	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/util/pipe"
)

type snapshotInfoCallback struct {
	SnapshotInfo
	callback chan<- error
}

type snapshotManagerImpl struct {
	log                 *logger.Logger
	ctx                 context.Context
	shutdownCoordinator *shutdown.Coordinator
	chainID             isc.ChainID

	lastIndexSnapshotted      uint32
	lastIndexSnapshottedMutex sync.Mutex
	createPeriod              uint32
	snapshotter               snapshotter

	availableSnapshots      *shrinkingmap.ShrinkingMap[uint32, []*state.L1Commitment]
	availableSnapshotsMutex sync.RWMutex

	updatePipe         pipe.Pipe[bool]
	blockCommittedPipe pipe.Pipe[SnapshotInfo]
	loadSnapshotPipe   pipe.Pipe[*snapshotInfoCallback]
}

var _ SnapshotManager = &snapshotManagerImpl{}

const downloadTimeout = 10 * time.Minute

func NewSnapshotManager(
	ctx context.Context,
	shutdownCoordinator *shutdown.Coordinator,
	chainID isc.ChainID,
	baseDir string,
	createPeriod uint32,
	store state.Store,
	log *logger.Logger,
) (SnapshotManager, error) {
	snapshotterImpl, err := newSnapshotter(log, baseDir, chainID, store)
	if err != nil {
		return nil, err
	}
	result := &snapshotManagerImpl{
		log:                       log,
		ctx:                       ctx,
		shutdownCoordinator:       shutdownCoordinator,
		chainID:                   chainID,
		lastIndexSnapshotted:      0,
		lastIndexSnapshottedMutex: sync.Mutex{},
		createPeriod:              createPeriod,
		snapshotter:               snapshotterImpl,
		availableSnapshots:        shrinkingmap.New[uint32, []*state.L1Commitment](),
		availableSnapshotsMutex:   sync.RWMutex{},
		updatePipe:                pipe.NewInfinitePipe[bool](),
		blockCommittedPipe:        pipe.NewInfinitePipe[SnapshotInfo](),
		loadSnapshotPipe:          pipe.NewInfinitePipe[*snapshotInfoCallback](),
	}
	go result.run()
	return result, nil
}

// -------------------------------------
// Implementations of SnapshotManager interface
// -------------------------------------

func (smiT *snapshotManagerImpl) UpdateAsync() {
	smiT.updatePipe.In() <- true
}

func (smiT *snapshotManagerImpl) BlockCommittedAsync(snapshotInfo SnapshotInfo) {
	smiT.blockCommittedPipe.In() <- snapshotInfo
}

func (smiT *snapshotManagerImpl) SnapshotExists(stateIndex uint32, commitment *state.L1Commitment) bool {
	smiT.availableSnapshotsMutex.RLock()
	defer smiT.availableSnapshotsMutex.RUnlock()

	commitments, exists := smiT.availableSnapshots.Get(stateIndex)
	if !exists {
		return false
	}
	return lo.ContainsBy(commitments, func(elem *state.L1Commitment) bool { return elem.Equals(commitment) })
}

func (smiT *snapshotManagerImpl) LoadSnapshotAsync(snapshotInfo SnapshotInfo) <-chan error {
	callback := make(chan error, 1)
	smiT.loadSnapshotPipe.In() <- &snapshotInfoCallback{
		SnapshotInfo: snapshotInfo,
		callback:     callback,
	}
	return callback
}

// -------------------------------------
// Internal functions
// -------------------------------------

func (smiT *snapshotManagerImpl) run() { //nolint:gocyclo
	updatePipeCh := smiT.updatePipe.Out()
	blockCommittedPipeCh := smiT.blockCommittedPipe.Out()
	loadSnapshotPipeCh := smiT.loadSnapshotPipe.Out()
	for {
		if smiT.ctx.Err() != nil {
			if smiT.shutdownCoordinator == nil {
				return
			}
			// TODO what should the statemgr wait for?
			if smiT.shutdownCoordinator.CheckNestedDone() {
				smiT.log.Debugf("Stopping snapshot manager, because context was closed")
				smiT.shutdownCoordinator.Done()
				return
			}
		}
		select {
		case _, ok := <-updatePipeCh:
			if ok {
				smiT.handleUpdate()
			} else {
				updatePipeCh = nil
			}
		case snapshotInfo, ok := <-blockCommittedPipeCh:
			if ok {
				smiT.handleBlockCommitted(snapshotInfo)
			} else {
				blockCommittedPipeCh = nil
			}
		case snapshotInfoC, ok := <-loadSnapshotPipeCh:
			if ok {
				smiT.handleLoadSnapshot(snapshotInfoC)
			} else {
				loadSnapshotPipeCh = nil
			}
		case <-smiT.ctx.Done():
			continue
		}
	}
}

func (smiT *snapshotManagerImpl) handleUpdate() {}

func (smiT *snapshotManagerImpl) handleBlockCommitted(snapshotInfo SnapshotInfo) {
	blockIndex := snapshotInfo.GetStateIndex()
	var lastIndexSnapshotted uint32
	smiT.lastIndexSnapshottedMutex.Lock()
	lastIndexSnapshotted = smiT.lastIndexSnapshotted
	smiT.lastIndexSnapshottedMutex.Unlock()
	if (blockIndex > lastIndexSnapshotted) && (blockIndex%smiT.createPeriod == 0) { // TODO: what if snapshotted state has been reverted?
		smiT.snapshotter.createSnapshotAsync(blockIndex, snapshotInfo.GetCommitment(), func() {
			smiT.lastIndexSnapshottedMutex.Lock()
			if blockIndex > smiT.lastIndexSnapshotted {
				smiT.lastIndexSnapshotted = blockIndex
			}
			smiT.lastIndexSnapshottedMutex.Unlock()
		})
	}
}

func (smiT *snapshotManagerImpl) handleLoadSnapshot(snapshotInfoCallback *snapshotInfoCallback) {
	callback := snapshotInfoCallback.callback
	defer close(callback)

	downloadCtx, downloadCtxCancel := context.WithTimeout(smiT.ctx, downloadTimeout)
	defer downloadCtxCancel()

	url := "TODO"
	request, err := http.NewRequestWithContext(downloadCtx, http.MethodGet, url, http.NoBody)
	if err != nil {
		callback <- fmt.Errorf("failed creating request with url %s: %w", url, err)
		return
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		callback <- fmt.Errorf("http request to url %s failed: %w", url, err)
		return
	}
	defer func() { _ = response.Body.Close() }()

	if response.StatusCode != http.StatusOK {
		callback <- fmt.Errorf("http request to %s got status code %v", url, response.StatusCode)
		return
	}

	progressReporter := NewProgressReporter(smiT.log, fmt.Sprintf("downloading snapshot from %s", url), uint64(response.ContentLength))
	reader := io.TeeReader(response.Body, progressReporter)
	err = smiT.snapshotter.loadSnapshot(reader)
	if err != nil {
		callback <- fmt.Errorf("downoloading snapshot from %s failed: %w", url, err)
		return
	}
	callback <- nil
}
