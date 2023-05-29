package sm_snapshots

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

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

type commitmentSources struct {
	commitment *state.L1Commitment
	sources    []string
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

	availableSnapshots      *shrinkingmap.ShrinkingMap[uint32, SliceStruct[*commitmentSources]]
	availableSnapshotsMutex sync.RWMutex

	localPath    string
	networkPaths []string

	updatePipe         pipe.Pipe[bool]
	blockCommittedPipe pipe.Pipe[SnapshotInfo]
	loadSnapshotPipe   pipe.Pipe[*snapshotInfoCallback]
}

var _ SnapshotManager = &snapshotManagerImpl{}

const (
	constDownloadTimeout                     = 10 * time.Minute
	constSnapshotIndexHashFileNameSepparator = "-"
	constSnapshotFileSuffix                  = ".snap"
	constSnapshotTmpFileSuffix               = ".tmp"
	constIndexFileName                       = "INDEX" // Index file contains a new-line separated list of snapshot files
	constLocalAddress                        = "local://"
)

func NewSnapshotManager(
	ctx context.Context,
	shutdownCoordinator *shutdown.Coordinator,
	chainID isc.ChainID,
	createPeriod uint32,
	baseLocalPath string,
	baseNetworkPaths []string,
	store state.Store,
	log *logger.Logger,
) (SnapshotManager, error) {
	chainIDString := chainID.String()
	localPath := filepath.Join(baseLocalPath, chainIDString)
	networkPaths := make([]string, len(baseNetworkPaths))
	var err error
	for i := range baseNetworkPaths {
		networkPaths[i], err = url.JoinPath(baseNetworkPaths[i], chainIDString)
		if err != nil {
			return nil, fmt.Errorf("cannot append chain ID to network path %s: %v", baseNetworkPaths[i], err)
		}
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
		availableSnapshots:        shrinkingmap.New[uint32, SliceStruct[*commitmentSources]](),
		availableSnapshotsMutex:   sync.RWMutex{},
		localPath:                 localPath,
		networkPaths:              networkPaths,
		updatePipe:                pipe.NewInfinitePipe[bool](),
		blockCommittedPipe:        pipe.NewInfinitePipe[SnapshotInfo](),
		loadSnapshotPipe:          pipe.NewInfinitePipe[*snapshotInfoCallback](),
	}
	if result.createSnapshotsNeeded() {
		if err := ioutils.CreateDirectory(localPath, 0o777); err != nil {
			return nil, fmt.Errorf("cannot create folder %s: %v", localPath, err)
		}
		result.cleanTempFiles() // To be able to make snapshots, which were not finished. See comment in `handleBlockCommitted` function
		log.Debugf("Snapshot manager created; folder %v is used for snapshots", localPath)
	} else {
		log.Debugf("Snapshot manager created; no snapshots will be produced")
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
	if smiT.createSnapshotsNeeded() {
		smiT.blockCommittedPipe.In() <- snapshotInfo
	}
}

func (smiT *snapshotManagerImpl) SnapshotExists(stateIndex uint32, commitment *state.L1Commitment) bool {
	smiT.availableSnapshotsMutex.RLock()
	defer smiT.availableSnapshotsMutex.RUnlock()

	commitments, exists := smiT.availableSnapshots.Get(stateIndex)
	if !exists {
		return false
	}
	return commitments.ContainsBy(func(elem *commitmentSources) bool { return elem.commitment.Equals(commitment) && len(elem.sources) > 0 })
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
	tempFileRegExp := tempSnapshotFileNameString("*", "*")
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

func (smiT *snapshotManagerImpl) run() {
	updatePipeCh := smiT.updatePipe.Out()
	blockCommittedPipeCh := smiT.blockCommittedPipe.Out()
	loadSnapshotPipeCh := smiT.loadSnapshotPipe.Out()
	for {
		if smiT.ctx.Err() != nil {
			if smiT.shutdownCoordinator == nil {
				return
			}
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

func (smiT *snapshotManagerImpl) handleUpdate() {
	result := shrinkingmap.New[uint32, SliceStruct[*commitmentSources]]()
	smiT.handleUpdateLocal(result)
	smiT.handleUpdateNetwork(result)

	smiT.availableSnapshotsMutex.Lock()
	smiT.availableSnapshots = result
	smiT.availableSnapshotsMutex.Unlock()
}

func (smiT *snapshotManagerImpl) handleUpdateLocal(result *shrinkingmap.ShrinkingMap[uint32, SliceStruct[*commitmentSources]]) {
	fileRegExp := snapshotFileNameString("*", "*")
	fileRegExpWithPath := filepath.Join(smiT.localPath, fileRegExp)
	files, err := filepath.Glob(fileRegExpWithPath)
	if err != nil {
		if smiT.createSnapshotsNeeded() {
			smiT.log.Errorf("Failed to obtain snapshot file list: %v", err)
		} else {
			// If snapshots are not created, snapshot dir is not supposed to exists; unless, it was created by other runs of Wasp or manually
			smiT.log.Warnf("Cannot obtain local snapshot file list (possibly, it does not exist): %v", err)
		}
		return
	}
	snapshotCount := 0
	for _, file := range files {
		func() { // Function to make the defers sooner
			f, err := os.Open(file)
			if err != nil {
				smiT.log.Errorf("Failed to open snapshot file %s: %v", file, err)
			}
			defer f.Close()
			snapshotInfo, err := readSnapshotInfo(f)
			if err != nil {
				smiT.log.Errorf("Failed to read snapshot info from file %s: %v", file, err)
				return
			}
			addSource(result, snapshotInfo, constLocalAddress+file)
			snapshotCount++
		}()
	}
	smiT.log.Debugf("%v snapshot files found locally", snapshotCount)
}

func (smiT *snapshotManagerImpl) handleUpdateNetwork(result *shrinkingmap.ShrinkingMap[uint32, SliceStruct[*commitmentSources]]) {
	for _, networkPath := range smiT.networkPaths {
		func() { // Function to make the defers sooner
			indexFilePath, err := url.JoinPath(networkPath, constIndexFileName)
			if err != nil {
				smiT.log.Errorf("Unable to join paths %s and %s: %v", networkPath, constIndexFileName, err)
				return
			}
			cancelFun, reader, err := downloadFile(smiT.ctx, smiT.log, indexFilePath, constDownloadTimeout)
			defer cancelFun()
			if err != nil {
				smiT.log.Errorf("Failed to download index file: %v", err)
				return
			}
			snapshotCount := 0
			scanner := bufio.NewScanner(reader) // Defaults to splitting input by newline character
			for scanner.Scan() {
				func() {
					snapshotFileName := scanner.Text()
					snapshotFilePath, er := url.JoinPath(networkPath, snapshotFileName)
					if er != nil {
						smiT.log.Errorf("Unable to join paths %s and %s: %v", networkPath, snapshotFileName, er)
						return
					}
					sCancelFun, sReader, er := downloadFile(smiT.ctx, smiT.log, snapshotFilePath, constDownloadTimeout)
					defer sCancelFun()
					if er != nil {
						smiT.log.Errorf("Failed to download snapshot file: %v", er)
						return
					}
					snapshotInfo, er := readSnapshotInfo(sReader)
					if er != nil {
						smiT.log.Errorf("Failed to download read snapshot info from %s: %v", snapshotFilePath, er)
						return
					}
					addSource(result, snapshotInfo, snapshotFilePath)
					snapshotCount++
				}()
			}
			err = scanner.Err()
			if err != nil && !errors.Is(err, io.EOF) {
				smiT.log.Errorf("Failed reading index file %s: %v", indexFilePath, err)
			}
			smiT.log.Debugf("%v snapshot files found on %s", snapshotCount, networkPath)
		}()
	}
}

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
		tmpFileName := tempSnapshotFileName(stateIndex, commitment.BlockHash())
		tmpFilePath := filepath.Join(smiT.localPath, tmpFileName)
		exists, _, _ := ioutils.PathExists(tmpFilePath)
		if exists {
			smiT.log.Debugf("Creating snapshot %v %s: skipped making snapshot as it is already being produced", stateIndex, commitment)
			return
		}
		f, err := os.OpenFile(tmpFilePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o666)
		if err != nil {
			smiT.log.Errorf("Creating snapshot %v %s: failed to create temporary snapshot file %s: %v", stateIndex, commitment, tmpFilePath, err)
			f.Close()
			return
		}
		go func() {
			defer f.Close()

			smiT.log.Debugf("Creating snapshot %v %s: storing it to file", stateIndex, commitment)
			err := smiT.snapshotter.storeSnapshot(snapshotInfo, f)
			if err != nil {
				smiT.log.Errorf("Creating snapshot %v %s: failed to write snapshot to temporary file %s: %v", stateIndex, commitment, tmpFilePath, err)
				return
			}

			finalFileName := snapshotFileName(stateIndex, commitment.BlockHash())
			finalFilePath := filepath.Join(smiT.localPath, finalFileName)
			err = os.Rename(tmpFilePath, finalFilePath)
			if err != nil {
				smiT.log.Errorf("Creating snapshot %v %s: failed to move temporary snapshot file %s to permanent location %s: %v",
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

	smiT.log.Debugf("Loading snapshot %s", snapshotInfoCallback.SnapshotInfo)
	// smiT.availableSnapshotsMutex.RLock() // Probably locking is not needed as it happens on the same thread as editing available snapshots
	commitments, exists := smiT.availableSnapshots.Get(snapshotInfoCallback.SnapshotInfo.GetStateIndex())
	// smiT.availableSnapshotsMutex.RUnlock()
	if !exists {
		err := fmt.Errorf("failed to obtain snapshot commitments of index %v", snapshotInfoCallback.SnapshotInfo.GetStateIndex())
		smiT.log.Errorf("Loading snapshot %s: %v", snapshotInfoCallback.SnapshotInfo, err)
		callback <- err
		return
	}
	cs, exists := commitments.Find(func(c *commitmentSources) bool {
		return c.commitment.Equals(snapshotInfoCallback.SnapshotInfo.GetCommitment())
	})
	if !exists {
		err := fmt.Errorf("failed to obtain sources of snapshot %s", snapshotInfoCallback.SnapshotInfo)
		smiT.log.Errorf("Loading snapshot %s: %v", snapshotInfoCallback.SnapshotInfo, err)
		callback <- err
		return
	}

	loadSnapshotFun := func(r io.Reader) error {
		err := smiT.snapshotter.loadSnapshot(snapshotInfoCallback.SnapshotInfo, r)
		if err != nil {
			return fmt.Errorf("loading snapshot failed: %v", err)
		}
		return nil
	}
	loadLocalFun := func(path string) error {
		f, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("failed to open snapshot file %s", path)
		}
		defer f.Close()
		return loadSnapshotFun(f)
	}
	loadNetworkFun := func(ctx context.Context, url string) error {
		closeFun, reader, err := downloadFile(ctx, smiT.log, url, constDownloadTimeout)
		defer closeFun()
		if err != nil {
			return err
		}
		return loadSnapshotFun(reader)
	}
	loadFun := func(source string) error {
		if strings.HasPrefix(source, constLocalAddress) {
			filePath := strings.TrimPrefix(source, constLocalAddress)
			smiT.log.Debugf("Loading snapshot %s: reading local file %s", snapshotInfoCallback.SnapshotInfo, filePath)
			return loadLocalFun(filePath)
		}
		smiT.log.Debugf("Loading snapshot %s: downloading file %s", snapshotInfoCallback.SnapshotInfo, source)
		return loadNetworkFun(smiT.ctx, source)
	}

	var err error
	for _, source := range cs.sources {
		e := loadFun(source)
		if e == nil {
			smiT.log.Debugf("Loading snapshot %s succeeded", snapshotInfoCallback.SnapshotInfo)
			callback <- nil
			return
		}
		smiT.log.Errorf("Loading snapshot %s: %v", snapshotInfoCallback.SnapshotInfo, e)
		err = errors.Join(err, e)
	}
	callback <- err
}

func (smiT *snapshotManagerImpl) createSnapshotsNeeded() bool {
	return smiT.createPeriod > 0
}

func tempSnapshotFileName(index uint32, blockHash state.BlockHash) string {
	return tempSnapshotFileNameString(fmt.Sprint(index), blockHash.String())
}

func tempSnapshotFileNameString(index, blockHash string) string {
	return snapshotFileNameString(index, blockHash) + constSnapshotTmpFileSuffix
}

func snapshotFileName(index uint32, blockHash state.BlockHash) string {
	return snapshotFileNameString(fmt.Sprint(index), blockHash.String())
}

func snapshotFileNameString(index, blockHash string) string {
	return index + constSnapshotIndexHashFileNameSepparator + blockHash + constSnapshotFileSuffix
}

func downloadFile(ctx context.Context, log *logger.Logger, url string, timeout time.Duration) (context.CancelFunc, io.Reader, error) {
	downloadCtx, downloadCtxCancel := context.WithTimeout(ctx, timeout)

	request, err := http.NewRequestWithContext(downloadCtx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return downloadCtxCancel, nil, fmt.Errorf("failed creating request with url %s: %v", url, err)
	}

	response, err := http.DefaultClient.Do(request) //nolint:bodyclose// it will be closed, when the caller calls `cancelFun`
	if err != nil {
		return downloadCtxCancel, nil, fmt.Errorf("http request to file url %s failed: %v", url, err)
	}
	cancelFun := func() {
		response.Body.Close()
		downloadCtxCancel()
	}

	if response.StatusCode != http.StatusOK {
		return cancelFun, nil, fmt.Errorf("http request to %s got status code %v", url, response.StatusCode)
	}

	progressReporter := NewProgressReporter(log, fmt.Sprintf("downloading file %s", url), uint64(response.ContentLength))
	reader := io.TeeReader(NewBlockingReader(response.Body), progressReporter)
	return cancelFun, reader, nil
}

func addSource(result *shrinkingmap.ShrinkingMap[uint32, SliceStruct[*commitmentSources]], si SnapshotInfo, path string) {
	makeNewComSourcesFun := func() *commitmentSources {
		return &commitmentSources{
			commitment: si.GetCommitment(),
			sources:    []string{path},
		}
	}
	comSourcesArray, exists := result.Get(si.GetStateIndex())
	if exists {
		comSources, ok := comSourcesArray.Find(func(elem *commitmentSources) bool { return elem.commitment.Equals(si.GetCommitment()) })
		if ok {
			comSources.sources = append(comSources.sources, path)
		} else {
			comSourcesArray.Add(makeNewComSourcesFun())
		}
	} else {
		comSourcesArray = NewSliceStruct[*commitmentSources](makeNewComSourcesFun())
		result.Set(si.GetStateIndex(), comSourcesArray)
	}
}
