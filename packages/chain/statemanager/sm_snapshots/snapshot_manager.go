package sm_snapshots

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
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
	"github.com/iotaledger/wasp/packages/metrics"
	"github.com/iotaledger/wasp/packages/shutdown"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util"
)

type commitmentSources struct {
	commitment *state.L1Commitment
	sources    []string
}

type snapshotManagerImpl struct {
	*snapshotManagerRunner

	log     *logger.Logger
	ctx     context.Context
	chainID isc.ChainID
	metrics *metrics.ChainSnapshotsMetrics

	lastIndexSnapshotted      uint32
	lastIndexSnapshottedMutex sync.Mutex
	createPeriod              uint32
	snapshotter               snapshotter

	availableSnapshots      *shrinkingmap.ShrinkingMap[uint32, *util.SliceStruct[*commitmentSources]]
	availableSnapshotsMutex sync.RWMutex

	localPath    string
	networkPaths []string
}

var (
	_ snapshotManagerCore = &snapshotManagerImpl{}
	_ SnapshotManager     = &snapshotManagerImpl{}
)

const (
	constDownloadTimeout                     = 10 * time.Minute
	constSnapshotIndexHashFileNameSepparator = "-"
	constSnapshotFileSuffix                  = ".snap"
	constSnapshotTmpFileSuffix               = ".tmp"
	constSnapshotDownloaded                  = "net"
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
	metrics *metrics.ChainSnapshotsMetrics,
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
	snapMLog := log.Named("Snap")
	result := &snapshotManagerImpl{
		log:                       snapMLog,
		ctx:                       ctx,
		chainID:                   chainID,
		metrics:                   metrics,
		lastIndexSnapshotted:      0,
		lastIndexSnapshottedMutex: sync.Mutex{},
		createPeriod:              createPeriod,
		snapshotter:               newSnapshotter(store),
		availableSnapshots:        shrinkingmap.New[uint32, *util.SliceStruct[*commitmentSources]](),
		availableSnapshotsMutex:   sync.RWMutex{},
		localPath:                 localPath,
		networkPaths:              networkPaths,
	}
	if err := ioutils.CreateDirectory(localPath, 0o777); err != nil {
		return nil, fmt.Errorf("cannot create folder %s: %v", localPath, err)
	}
	if result.createSnapshotsNeeded() {
		result.cleanTempFiles() // To be able to make snapshots, which were not finished. See comment in `handleBlockCommitted` function
		snapMLog.Debugf("Snapshot manager created; folder %v is used for snapshots", localPath)
	} else {
		snapMLog.Debugf("Snapshot manager created; folder %v is used to download snapshots; no snapshots will be produced", localPath)
	}
	result.snapshotManagerRunner = newSnapshotManagerRunner(ctx, shutdownCoordinator, result, snapMLog)
	return result, nil
}

// -------------------------------------
// Implementations of SnapshotManager interface
// -------------------------------------

func (smiT *snapshotManagerImpl) SnapshotExists(stateIndex uint32, commitment *state.L1Commitment) bool {
	smiT.availableSnapshotsMutex.RLock()
	defer smiT.availableSnapshotsMutex.RUnlock()

	commitments, exists := smiT.availableSnapshots.Get(stateIndex)
	if !exists {
		return false
	}
	return commitments.ContainsBy(func(elem *commitmentSources) bool { return elem.commitment.Equals(commitment) && len(elem.sources) > 0 })
}

// NOTE: other implementations are inherited from snapshotManagerRunner

// -------------------------------------
// Implementations of snapshotManagerCore interface
// -------------------------------------

func (smiT *snapshotManagerImpl) createSnapshotsNeeded() bool {
	return smiT.createPeriod > 0
}

func (smiT *snapshotManagerImpl) handleUpdate() {
	start := time.Now()
	result := shrinkingmap.New[uint32, *util.SliceStruct[*commitmentSources]]()
	smiT.handleUpdateLocal(result)
	smiT.handleUpdateNetwork(result)

	smiT.availableSnapshotsMutex.Lock()
	smiT.availableSnapshots = result
	smiT.availableSnapshotsMutex.Unlock()
	smiT.metrics.SnapshotsUpdated(time.Since(start))
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
	start := time.Now()
	stateIndex := snapshotInfo.StateIndex()
	var lastIndexSnapshotted uint32
	smiT.lastIndexSnapshottedMutex.Lock()
	lastIndexSnapshotted = smiT.lastIndexSnapshotted
	smiT.lastIndexSnapshottedMutex.Unlock()
	if (stateIndex > lastIndexSnapshotted) && (stateIndex%smiT.createPeriod == 0) { // TODO: what if snapshotted state has been reverted?
		commitment := snapshotInfo.Commitment()
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
			smiT.metrics.SnapshotCreated(time.Since(start), stateIndex)
		}()
	}
}

func (smiT *snapshotManagerImpl) handleLoadSnapshot(snapshotInfo SnapshotInfo, callback chan<- error) {
	start := time.Now()
	smiT.log.Debugf("Loading snapshot %s", snapshotInfo)
	// smiT.availableSnapshotsMutex.RLock() // Probably locking is not needed as it happens on the same thread as editing available snapshots
	commitments, exists := smiT.availableSnapshots.Get(snapshotInfo.StateIndex())
	// smiT.availableSnapshotsMutex.RUnlock()
	if !exists {
		err := fmt.Errorf("failed to obtain snapshot commitments of index %v", snapshotInfo.StateIndex())
		smiT.log.Errorf("Loading snapshot %s: %v", snapshotInfo, err)
		callback <- err
		return
	}
	cs, exists := commitments.Find(func(c *commitmentSources) bool {
		return c.commitment.Equals(snapshotInfo.Commitment())
	})
	if !exists {
		err := fmt.Errorf("failed to obtain sources of snapshot %s", snapshotInfo)
		smiT.log.Errorf("Loading snapshot %s: %v", snapshotInfo, err)
		callback <- err
		return
	}

	loadSnapshotFun := func(r io.Reader) error {
		err := smiT.snapshotter.loadSnapshot(snapshotInfo, r)
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
	loadNetworkFun := func(url string) error {
		fileNameLocal := downloadedSnapshotFileName(snapshotInfo.StateIndex(), snapshotInfo.BlockHash())
		filePathLocal := filepath.Join(smiT.localPath, fileNameLocal)
		localPathFun, err := DownloadToFile(smiT.ctx, url, filePathLocal, constDownloadTimeout, smiT.addProgressReporter)
		if err != nil {
			return err
		}
		return loadLocalFun(localPathFun)
	}
	loadFun := func(source string) error {
		if strings.HasPrefix(source, constLocalAddress) {
			filePath := strings.TrimPrefix(source, constLocalAddress)
			smiT.log.Debugf("Loading snapshot %s: reading local file %s", snapshotInfo, filePath)
			return loadLocalFun(filePath)
		}
		smiT.log.Debugf("Loading snapshot %s: downloading file %s", snapshotInfo, source)
		return loadNetworkFun(source)
	}

	var err error
	for _, source := range cs.sources {
		e := loadFun(source)
		if e == nil {
			smiT.log.Debugf("Loading snapshot %s succeeded", snapshotInfo)
			callback <- nil
			smiT.metrics.SnapshotLoaded(time.Since(start))
			return
		}
		smiT.log.Errorf("Loading snapshot %s: %v", snapshotInfo, e)
		err = errors.Join(err, e)
	}
	callback <- err
}

// -------------------------------------
// Internal functions
// -------------------------------------

// This happens strictly before snapshot manager starts to produce new snapshots.
// So there is no way that this function will delete temp file, which is needed.
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

func (smiT *snapshotManagerImpl) handleUpdateLocal(result *shrinkingmap.ShrinkingMap[uint32, *util.SliceStruct[*commitmentSources]]) {
	fileRegExp := snapshotFileNameString("*", "*")
	fileRegExpWithPath := filepath.Join(smiT.localPath, fileRegExp)
	files, err := filepath.Glob(fileRegExpWithPath)
	if err != nil {
		if smiT.createSnapshotsNeeded() {
			smiT.log.Errorf("Update local: failed to obtain snapshot file list: %v", err)
		} else {
			// If snapshots are not created, snapshot dir is not supposed to exists; unless, it was created by other runs of Wasp or manually
			smiT.log.Warnf("Update local: cannot obtain snapshot file list (possibly, it does not exist): %v", err)
		}
		return
	}
	snapshotCount := 0
	for _, file := range files {
		func() { // Function to make the defers sooner
			f, err := os.Open(file)
			if err != nil {
				smiT.log.Errorf("Update local: failed to open snapshot file %s: %v", file, err)
			}
			defer f.Close()
			snapshotInfo, err := readSnapshotInfo(f)
			if err != nil {
				smiT.log.Errorf("Update local: failed to read snapshot info from file %s: %v", file, err)
				return
			}
			addSource(result, snapshotInfo, constLocalAddress+file)
			snapshotCount++
		}()
	}
	smiT.log.Debugf("Update local: %v snapshot files found", snapshotCount)
}

func (smiT *snapshotManagerImpl) handleUpdateNetwork(result *shrinkingmap.ShrinkingMap[uint32, *util.SliceStruct[*commitmentSources]]) {
	for _, networkPath := range smiT.networkPaths {
		func() { // Function to make the defers sooner
			indexFilePath, err := url.JoinPath(networkPath, constIndexFileName)
			if err != nil {
				smiT.log.Errorf("Update network: unable to join paths %s and %s: %v", networkPath, constIndexFileName, err)
				return
			}
			reader, err := smiT.initiateDownload(indexFilePath, constDownloadTimeout)
			if err != nil {
				smiT.log.Errorf("Update network: failed to download index file: %v", err)
				return
			}
			defer reader.Close()
			snapshotCount := 0
			scanner := bufio.NewScanner(reader) // Defaults to splitting input by newline character
			for scanner.Scan() {
				func() {
					snapshotFileName := scanner.Text()
					snapshotFilePath, er := url.JoinPath(networkPath, snapshotFileName)
					if er != nil {
						smiT.log.Errorf("Update network: unable to join paths %s and %s: %v", networkPath, snapshotFileName, er)
						return
					}
					sReader, er := smiT.initiateDownload(snapshotFilePath, constDownloadTimeout)
					if er != nil {
						smiT.log.Errorf("Update network: failed to download snapshot file: %v", er)
						return
					}
					defer sReader.Close()
					snapshotInfo, er := readSnapshotInfo(sReader)
					if er != nil {
						smiT.log.Errorf("Update network: failed to read snapshot info from %s: %v", snapshotFilePath, er)
						return
					}
					addSource(result, snapshotInfo, snapshotFilePath)
					snapshotCount++
				}()
			}
			err = scanner.Err()
			if err != nil && !errors.Is(err, io.EOF) {
				smiT.log.Errorf("Update network: failed reading index file %s: %v", indexFilePath, err)
			}
			smiT.log.Debugf("Update network: %v snapshot files found on %s", snapshotCount, networkPath)
		}()
	}
}

func (smiT *snapshotManagerImpl) initiateDownload(url string, timeout time.Duration) (io.ReadCloser, error) {
	downloader, err := NewDownloaderWithTimeout(smiT.ctx, url, timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to start downloading file from url %s: %v", url, err)
	}
	r := smiT.addProgressReporter(downloader, url, downloader.GetLength())
	return NewReaderWithClose(r, downloader.Close), nil
}

func (smiT *snapshotManagerImpl) addProgressReporter(r io.Reader, url string, length uint64) io.Reader {
	progressReporter := NewProgressReporter(smiT.log, fmt.Sprintf("downloading file %s", url), length)
	return io.TeeReader(r, progressReporter)
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

func downloadedSnapshotFileName(index uint32, blockHash state.BlockHash) string {
	return downloadedSnapshotFileNameString(fmt.Sprint(index), blockHash.String())
}

func downloadedSnapshotFileNameString(index, blockHash string) string {
	return index + constSnapshotIndexHashFileNameSepparator + blockHash +
		constSnapshotIndexHashFileNameSepparator + constSnapshotDownloaded + constSnapshotFileSuffix
}

func addSource(result *shrinkingmap.ShrinkingMap[uint32, *util.SliceStruct[*commitmentSources]], si SnapshotInfo, path string) {
	makeNewComSourcesFun := func() *commitmentSources {
		return &commitmentSources{
			commitment: si.Commitment(),
			sources:    []string{path},
		}
	}
	comSourcesArray, exists := result.Get(si.StateIndex())
	if exists {
		comSources, ok := comSourcesArray.Find(func(elem *commitmentSources) bool { return elem.commitment.Equals(si.Commitment()) })
		if ok {
			comSources.sources = append(comSources.sources, path)
		} else {
			comSourcesArray.Add(makeNewComSourcesFun())
		}
	} else {
		comSourcesArray = util.NewSliceStruct[*commitmentSources](makeNewComSourcesFun())
		result.Set(si.StateIndex(), comSourcesArray)
	}
}
