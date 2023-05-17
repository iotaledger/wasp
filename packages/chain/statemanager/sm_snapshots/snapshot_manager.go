package sm_snapshots

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/samber/lo"

	"github.com/iotaledger/hive.go/ds/shrinkingmap"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/runtime/ioutils"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/shutdown"
	"github.com/iotaledger/wasp/packages/state"
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

	localPath string

	updatePipe         pipe.Pipe[bool]
	blockCommittedPipe pipe.Pipe[SnapshotInfo]
	loadSnapshotPipe   pipe.Pipe[*snapshotInfoCallback]
}

var _ SnapshotManager = &snapshotManagerImpl{}

const (
	constDownloadTimeout       = 10 * time.Minute
	constSnapshotFileSuffix    = ".snap"
	constSnapshotTmpFileSuffix = ".tmp"
)

func NewSnapshotManager(
	ctx context.Context,
	shutdownCoordinator *shutdown.Coordinator,
	chainID isc.ChainID,
	baseDir string,
	createPeriod uint32,
	store state.Store,
	log *logger.Logger,
) (SnapshotManager, error) {
	localPath := filepath.Join(baseDir, chainID.String())
	if err := ioutils.CreateDirectory(localPath, 0o777); err != nil {
		return nil, fmt.Errorf("cannot create folder %s: %w", localPath, err)
	}
	result := &snapshotManagerImpl{
		log:                       log,
		ctx:                       ctx,
		shutdownCoordinator:       shutdownCoordinator,
		chainID:                   chainID,
		lastIndexSnapshotted:      0,
		lastIndexSnapshottedMutex: sync.Mutex{},
		createPeriod:              createPeriod,
		snapshotter:               newSnapshotter(store),
		availableSnapshots:        shrinkingmap.New[uint32, []*state.L1Commitment](),
		availableSnapshotsMutex:   sync.RWMutex{},
		localPath:                 localPath,
		updatePipe:                pipe.NewInfinitePipe[bool](),
		blockCommittedPipe:        pipe.NewInfinitePipe[SnapshotInfo](),
		loadSnapshotPipe:          pipe.NewInfinitePipe[*snapshotInfoCallback](),
	}
	result.cleanTempFiles() // To be able to make snapshots, which were not finished. See comment in `handleBlockCommitted` function
	go result.run()
	log.Debugf("Snapshotter created; folder %v is used for snapshots", localPath)
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

func (smiT *snapshotManagerImpl) cleanTempFiles() {
	tempFileRegExp := tempSnapshotFileNameString("*")
	tempFileRegExpWithPath := filepath.Join(smiT.localPath, tempFileRegExp)
	tempFiles, err := filepath.Glob(tempFileRegExpWithPath)
	if err != nil {
		smiT.log.Errorf("Failed to obtain temporary snapshot file list: %v", err)
		return
	}

	removed := 0
	for _, tempFile := range tempFiles {
		err = os.Remove(tempFile)
		if err != nil {
			smiT.log.Warnf("Failed to remove temporary snapshot file %s: %v", tempFile, err)
		} else {
			removed++
		}
	}
	smiT.log.Debugf("Removed %v out of %v temporary snapshot files", removed, len(tempFiles))
}

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

// Snapshot manager makes snapshot of every `period`th state only, if this state hasn't
// been snapshotted before. The snapshot file name includes state index and state hash.
// Snapshot manager first writes the state to temporary file and only then moves it to
// permanent location. Writing is done in separate thread to not interfere with
// normal State manager routine, as it may be lengthy. If snapshot manager detects that
// the temporary file, needed to create a snapshot, already exists, it assumes
// that another go routine is already making a snapshot and returns. For this reason
// it is important to delete all temporary files on snapshot manager start.
func (smiT *snapshotManagerImpl) handleBlockCommitted(snapshotInfo SnapshotInfo) {
	stateIndex := snapshotInfo.GetStateIndex()
	var lastIndexSnapshotted uint32
	smiT.lastIndexSnapshottedMutex.Lock()
	lastIndexSnapshotted = smiT.lastIndexSnapshotted
	smiT.lastIndexSnapshottedMutex.Unlock()
	if (stateIndex > lastIndexSnapshotted) && (stateIndex%smiT.createPeriod == 0) { // TODO: what if snapshotted state has been reverted?
		commitment := snapshotInfo.GetCommitment()
		smiT.log.Debugf("Creating snapshot %v %s...", stateIndex, commitment)
		tmpFileName := tempSnapshotFileName(commitment.BlockHash())
		tmpFilePath := filepath.Join(smiT.localPath, tmpFileName)
		exists, _, _ := ioutils.PathExists(tmpFilePath)
		if exists {
			smiT.log.Debugf("Creating snapshot %v %s: skipped making snapshot as it is already being produced", stateIndex, commitment)
			return
		}
		f, err := os.OpenFile(tmpFilePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o666)
		if err != nil {
			smiT.log.Errorf("Creating snapshot %v %s: failed to create temporary snapshot file %s: %w", stateIndex, commitment, tmpFilePath, err)
			f.Close()
			return
		}
		go func() {
			defer f.Close()

			smiT.log.Debugf("Creating snapshot %v %s: storing it to file", stateIndex, commitment)
			err := smiT.snapshotter.storeSnapshot(snapshotInfo, f)
			if err != nil {
				smiT.log.Errorf("Creating snapshot %v %s: failed to write snapshot to temporary file %s: %w", stateIndex, commitment, tmpFilePath, err)
				return
			}

			finalFileName := snapshotFileName(commitment.BlockHash())
			finalFilePath := filepath.Join(smiT.localPath, finalFileName)
			err = os.Rename(tmpFilePath, finalFilePath)
			if err != nil {
				smiT.log.Errorf("Creating snapshot %v %s: failed to move temporary snapshot file %s to permanent location %s: %w",
					stateIndex, commitment, tmpFilePath, finalFilePath, err)
				return
			}

			smiT.lastIndexSnapshottedMutex.Lock()
			if stateIndex > smiT.lastIndexSnapshotted {
				smiT.lastIndexSnapshotted = stateIndex
			}
			smiT.lastIndexSnapshottedMutex.Unlock()
			smiT.log.Infof("Creating snapshot %v %s: snapshot created in %s", stateIndex, commitment, finalFilePath)
		}()
	}
}

func (smiT *snapshotManagerImpl) handleLoadSnapshot(snapshotInfoCallback *snapshotInfoCallback) {
	callback := snapshotInfoCallback.callback
	defer close(callback)

	downloadCtx, downloadCtxCancel := context.WithTimeout(smiT.ctx, constDownloadTimeout)
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
	err = smiT.snapshotter.loadSnapshot(snapshotInfoCallback.SnapshotInfo, reader)
	if err != nil {
		callback <- fmt.Errorf("downoloading snapshot from %s failed: %w", url, err)
		return
	}
	callback <- nil
}

func tempSnapshotFileName(blockHash state.BlockHash) string {
	return tempSnapshotFileNameString(blockHash.String())
}

func tempSnapshotFileNameString(blockHash string) string {
	return snapshotFileNameString(blockHash) + constSnapshotTmpFileSuffix
}

func snapshotFileName(blockHash state.BlockHash) string {
	return snapshotFileNameString(blockHash.String())
}

func snapshotFileNameString(blockHash string) string {
	return blockHash + constSnapshotFileSuffix
}
