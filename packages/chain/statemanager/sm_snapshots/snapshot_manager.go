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

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/runtime/ioutils"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/metrics"
	"github.com/iotaledger/wasp/packages/shutdown"
	"github.com/iotaledger/wasp/packages/state"
)

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

	localPath string
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
	localPath := filepath.Join(baseLocalPath, chainID.String())
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
		localPath:                 localPath,
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
	if store.IsEmpty() {
		result.loadSnapshot(baseNetworkPaths)
	}
	result.snapshotManagerRunner = newSnapshotManagerRunner(ctx, shutdownCoordinator, result, snapMLog)
	return result, nil
}

// -------------------------------------
// Implementations of SnapshotManager interface
// -------------------------------------

// NOTE: implementation is inherited from snapshotManagerRunner

// -------------------------------------
// Implementations of snapshotManagerCore interface
// -------------------------------------

func (smiT *snapshotManagerImpl) createSnapshotsNeeded() bool {
	return smiT.createPeriod > 0
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

func (smiT *snapshotManagerImpl) loadSnapshot(baseNetworkPaths []string) {
	snapshotPaths := make([]string, 0)
	snapshotInfos := make([]SnapshotInfo, 0)
	largestIndex := uint32(0)
	considerSnapshotFun := func(snapshotInfo SnapshotInfo, path string) {
		if snapshotInfo.StateIndex() < largestIndex {
			smiT.log.Debugf("Snapshot %s found in %s; it is ignored as its index is lower than current largest index %v", path, snapshotInfo, largestIndex)
			return
		}
		if snapshotInfo.StateIndex() == largestIndex {
			snapshotPaths = append(snapshotPaths, path)
			snapshotInfos = append(snapshotInfos, snapshotInfo)
			smiT.log.Debugf("Snapshot %s found in %s; it is added to the list of considered snapshots", path, snapshotInfo)
			return
		}
		// NOTE: snapshotInfo.StateIndex() > largestIndex
		snapshotPaths = []string{path}
		snapshotInfos = []SnapshotInfo{snapshotInfo}
		smiT.log.Debugf("Snapshot %s found in %s; it is now the only considered snapshot as its index is larger than former largest index %v", path, snapshotInfo, largestIndex)
		largestIndex = snapshotInfo.StateIndex()
	}

	smiT.searchLocalSnapshots(considerSnapshotFun)
	smiT.searchNetworkSnapshots(baseNetworkPaths, considerSnapshotFun)
	smiT.log.Debugf("%v snapshots with state index %v will be considered for loading in this order: %v", len(snapshotPaths), largestIndex, snapshotPaths)

	for i := range snapshotPaths {
		err := smiT.loadSnapshotFromPath(snapshotInfos[i], snapshotPaths[i])
		if err == nil {
			smiT.log.Debugf("Snapshot %s successfully loaded from %s", snapshotInfos[i], snapshotPaths[i])
			return
		}
		smiT.log.Errorf("Failed to load snapshot %s from %s: %v", snapshotInfos[i], snapshotPaths[i], err)
	}
	smiT.log.Warnf("Failed to load any snapshot; will continue with empty store")
}

func (smiT *snapshotManagerImpl) searchLocalSnapshots(considerSnapshotFun func(SnapshotInfo, string)) {
	fileRegExp := snapshotFileNameString("*", "*")
	fileRegExpWithPath := filepath.Join(smiT.localPath, fileRegExp)
	files, err := filepath.Glob(fileRegExpWithPath)
	if err != nil {
		smiT.log.Errorf("Search local snapshots: failed to obtain snapshot file list: %v", err)
		return
	}
	snapshotCount := 0
	for _, file := range files {
		func() { // Function to make the defers sooner
			f, err := os.Open(file)
			if err != nil {
				smiT.log.Errorf("Search local snapshots: failed to open snapshot file %s: %v", file, err)
				return
			}
			defer f.Close()
			snapshotInfo, err := readSnapshotInfo(f)
			if err != nil {
				smiT.log.Errorf("Search local snapshots: failed to read snapshot info from file %s: %v", file, err)
				return
			}
			considerSnapshotFun(snapshotInfo, constLocalAddress+file)
			snapshotCount++
		}()
	}
	smiT.log.Debugf("Search local snapshots: %v snapshot files found", snapshotCount)
}

func (smiT *snapshotManagerImpl) searchNetworkSnapshots(baseNetworkPaths []string, considerSnapshotFun func(SnapshotInfo, string)) {
	chainIDString := smiT.chainID.String()
	for _, baseNetworkPath := range baseNetworkPaths {
		func() { // Function to make the defers sooner
			indexFilePath, err := url.JoinPath(baseNetworkPath, chainIDString, constIndexFileName)
			if err != nil {
				smiT.log.Errorf("Search network snapshots: unable to join paths %s, %s and %s: %v", baseNetworkPath, chainIDString, constIndexFileName, err)
				return
			}
			reader, err := smiT.initiateDownload(indexFilePath, constDownloadTimeout)
			if err != nil {
				smiT.log.Errorf("Search network snapshots: failed to download index file: %v", indexFilePath, err)
				return
			}
			defer reader.Close()
			snapshotCount := 0
			scanner := bufio.NewScanner(reader) // Defaults to splitting input by newline character
			for scanner.Scan() {
				func() {
					snapshotFileName := scanner.Text()
					snapshotFilePath, er := url.JoinPath(baseNetworkPath, chainIDString, snapshotFileName)
					if er != nil {
						smiT.log.Errorf("Search network snapshots: unable to join paths %s, %s and %s: %v", baseNetworkPath, chainIDString, snapshotFileName, er)
						return
					}
					sReader, er := smiT.initiateDownload(snapshotFilePath, constDownloadTimeout)
					if er != nil {
						smiT.log.Errorf("Search network snapshots: failed to download snapshot file: %v", er)
						return
					}
					defer sReader.Close()
					snapshotInfo, er := readSnapshotInfo(sReader)
					if er != nil {
						smiT.log.Errorf("Search network snapshots: failed to read snapshot info from %s: %v", snapshotFilePath, er)
						return
					}
					considerSnapshotFun(snapshotInfo, snapshotFilePath)
					snapshotCount++
				}()
			}
			err = scanner.Err()
			if err != nil && !errors.Is(err, io.EOF) {
				smiT.log.Errorf("Search network snapshots: failed reading index file %s: %v", indexFilePath, err)
			}
			smiT.log.Debugf("Search network snapshots: %v snapshot files found on %s", snapshotCount, baseNetworkPath)
		}()
	}
}

func (smiT *snapshotManagerImpl) loadSnapshotFromPath(snapshotInfo SnapshotInfo, path string) error {
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
		err := DownloadToFile(smiT.ctx, url, filePathLocal, constDownloadTimeout, smiT.addProgressReporter)
		if err != nil {
			return err
		}
		return loadLocalFun(filePathLocal)
	}

	if strings.HasPrefix(path, constLocalAddress) {
		filePath := strings.TrimPrefix(path, constLocalAddress)
		smiT.log.Debugf("Loading snapshot %s from file %s...", snapshotInfo, filePath)
		return loadLocalFun(filePath)
	}
	smiT.log.Debugf("Loading snapshot %s from url %s...", snapshotInfo, path)
	return loadNetworkFun(path)
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
