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
	"time"

	"github.com/iotaledger/hive.go/log"
	"github.com/iotaledger/hive.go/runtime/ioutils"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/metrics"
	"github.com/iotaledger/wasp/packages/shutdown"
	"github.com/iotaledger/wasp/packages/state"
)

type snapshotManagerImpl struct {
	*snapshotManagerRunner

	log     log.Logger
	ctx     context.Context
	chainID isc.ChainID
	metrics *metrics.ChainSnapshotsMetrics

	snapshotter      snapshotter
	localPath        string
	baseNetworkPaths []string
	snapshotToLoad   *state.BlockHash
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
	constLocalAddress                        = "file://"
	constSchemeHTTP                          = "http(s)"
	constSchemeFile                          = "file"
)

func NewSnapshotManager(
	ctx context.Context,
	shutdownCoordinator *shutdown.Coordinator,
	chainID isc.ChainID,
	snapshotToLoad *state.BlockHash,
	createPeriod uint32,
	delayPeriod uint32,
	baseLocalPath string,
	baseNetworkPaths []string,
	store state.Store,
	metrics *metrics.ChainSnapshotsMetrics,
	log log.Logger,
) (SnapshotManager, error) {
	localPath := filepath.Join(baseLocalPath, chainID.String())
	snapMLog := log.NewChildLogger("Snap")
	result := &snapshotManagerImpl{
		log:              snapMLog,
		ctx:              ctx,
		chainID:          chainID,
		metrics:          metrics,
		snapshotter:      newSnapshotter(store),
		localPath:        localPath,
		baseNetworkPaths: baseNetworkPaths,
		snapshotToLoad:   snapshotToLoad,
	}
	if err := ioutils.CreateDirectory(localPath, 0o777); err != nil {
		return nil, fmt.Errorf("cannot create folder %s: %v", localPath, err)
	}
	result.cleanTempFiles() // To be able to make snapshots, which were not finished. See comment in `createSnapshot` function
	snapMLog.LogDebugf("Snapshot manager created; folder %v is used for snapshots", localPath)
	result.snapshotManagerRunner = newSnapshotManagerRunner(ctx, store, shutdownCoordinator, createPeriod, delayPeriod, result, snapMLog)
	return result, nil
}

// -------------------------------------
// Implementations of SnapshotManager interface
// -------------------------------------

// NOTE: implementations are inherited from snapshotManagerRunner

// -------------------------------------
// Implementations of snapshotManagerCore interface
// -------------------------------------

// Snapshot file name includes state index and state hash. Snapshot manager first
// writes the state to temporary file and only then moves it to permanent location.
// Writing is done in separate thread to not interfere with normal snapshot manager
// routine, as it may be lengthy. If snapshot manager detects that the temporary
// file, needed to create a snapshot, already exists, it assumes that another go
// routine is already making a snapshot and returns. For this reason it is important
// to delete all temporary files on snapshot manager start.
func (smiT *snapshotManagerImpl) createSnapshot(snapshotInfo SnapshotInfo) {
	start := time.Now()
	stateIndex := snapshotInfo.StateIndex()
	commitment := snapshotInfo.Commitment()
	smiT.log.LogDebugf("Creating snapshot %v %s...", stateIndex, commitment)
	tmpFileName := tempSnapshotFileName(stateIndex, commitment.BlockHash())
	tmpFilePath := filepath.Join(smiT.localPath, tmpFileName)
	exists, _, _ := ioutils.PathExists(tmpFilePath)
	if exists {
		smiT.log.LogDebugf("Creating snapshot %v %s: skipped making snapshot as it is already being produced", stateIndex, commitment)
		return
	}
	f, err := os.OpenFile(tmpFilePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o666)
	if err != nil {
		smiT.log.LogErrorf("Creating snapshot %v %s: failed to create temporary snapshot file %s: %v", stateIndex, commitment, tmpFilePath, err)
		f.Close()
		return
	}
	go func() {
		defer f.Close()

		smiT.log.LogDebugf("Creating snapshot %v %s: storing it to file", stateIndex, commitment)
		err := smiT.snapshotter.storeSnapshot(snapshotInfo, f)
		if err != nil {
			smiT.log.LogErrorf("Creating snapshot %v %s: failed to write snapshot to temporary file %s: %v", stateIndex, commitment, tmpFilePath, err)
			return
		}

		finalFileName := snapshotFileName(stateIndex, commitment.BlockHash())
		finalFilePath := filepath.Join(smiT.localPath, finalFileName)
		err = os.Rename(tmpFilePath, finalFilePath)
		if err != nil {
			smiT.log.LogErrorf("Creating snapshot %v %s: failed to move temporary snapshot file %s to permanent location %s: %v",
				stateIndex, commitment, tmpFilePath, finalFilePath, err)
			return
		}
		smiT.snapshotCreated(snapshotInfo)
		smiT.log.LogInfof("Creating snapshot %v %s: snapshot created in %s", stateIndex, commitment, finalFilePath)
		smiT.metrics.SnapshotCreated(time.Since(start), stateIndex)
	}()
}

func (smiT *snapshotManagerImpl) loadSnapshot() SnapshotInfo {
	snapshotPaths := make([]string, 0)
	snapshotInfos := make([]SnapshotInfo, 0)

	var considerSnapshotFun func(snapshotInfo SnapshotInfo, path string)
	var searchCondition string
	if smiT.snapshotToLoad == nil {
		largestIndex := uint32(0)
		considerSnapshotFun = func(snapshotInfo SnapshotInfo, path string) {
			if snapshotInfo.StateIndex() < largestIndex {
				smiT.log.LogDebugf("Snapshot %s found in %s; it is ignored, because its index is lower than current largest index %v",
					path, snapshotInfo, largestIndex)
				return
			}
			if snapshotInfo.StateIndex() == largestIndex {
				snapshotPaths = append(snapshotPaths, path)
				snapshotInfos = append(snapshotInfos, snapshotInfo)
				smiT.log.LogDebugf("Snapshot %s found in %s; it is added to the list of considered snapshots, because its index mathec current largest index",
					path, snapshotInfo)
				return
			}
			// NOTE: snapshotInfo.StateIndex() > largestIndex
			snapshotPaths = []string{path}
			snapshotInfos = []SnapshotInfo{snapshotInfo}
			smiT.log.LogDebugf("Snapshot %s found in %s; it is now the only considered snapshot, because its index is larger than former largest index %v",
				path, snapshotInfo, largestIndex)
			largestIndex = snapshotInfo.StateIndex()
		}
		searchCondition = fmt.Sprintf("state index %v", largestIndex)
	} else {
		considerSnapshotFun = func(snapshotInfo SnapshotInfo, path string) {
			if snapshotInfo.BlockHash().Equals(*smiT.snapshotToLoad) {
				snapshotPaths = append(snapshotPaths, path)
				snapshotInfos = append(snapshotInfos, snapshotInfo)
				smiT.log.LogDebugf("Snapshot %s found in %s; it is added to the list of considered snapshots, because its hash matches what was requested",
					path, snapshotInfo)
				return
			}
			smiT.log.LogDebugf("Snapshot %s found in %s; it is ignored, because its hash does not match what was requested", path, snapshotInfo)
		}
		searchCondition = fmt.Sprintf("block hash %s", *smiT.snapshotToLoad)
	}

	smiT.searchLocalSnapshots(considerSnapshotFun)
	smiT.searchNetworkSnapshots(smiT.baseNetworkPaths, considerSnapshotFun)
	smiT.log.LogDebugf("%v snapshots with %s will be considered for loading in this order: %v", len(snapshotPaths), searchCondition, snapshotPaths)

	for i := range snapshotPaths {
		err := smiT.loadSnapshotFromPath(snapshotInfos[i], snapshotPaths[i])
		if err == nil {
			smiT.log.LogInfof("Snapshot %s successfully loaded from %s", snapshotInfos[i], snapshotPaths[i])
			return snapshotInfos[i]
		}
		smiT.log.LogErrorf("Failed to load snapshot %s from %s: %v", snapshotInfos[i], snapshotPaths[i], err)
	}
	smiT.log.LogWarnf("Failed to load any snapshot; will continue with empty store")
	return nil
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
		smiT.log.LogErrorf("Failed to obtain temporary snapshot file list: %v", err)
		return
	}

	removed := 0
	for _, tempFile := range tempFiles {
		err = os.Remove(tempFile)
		if err != nil {
			smiT.log.LogWarnf("Failed to remove temporary snapshot file %s: %v", tempFile, err)
		} else {
			removed++
		}
	}
	smiT.log.LogDebugf("Removed %v out of %v temporary snapshot files", removed, len(tempFiles))
}

func (smiT *snapshotManagerImpl) searchLocalSnapshots(considerSnapshotFun func(SnapshotInfo, string)) {
	fileRegExp := snapshotFileNameString("*", "*")
	fileRegExpWithPath := filepath.Join(smiT.localPath, fileRegExp)
	files, err := filepath.Glob(fileRegExpWithPath)
	if err != nil {
		smiT.log.LogErrorf("Search local snapshots: failed to obtain snapshot file list: %v", err)
		return
	}
	snapshotCount := 0
	for _, file := range files {
		func() { // Function to make the defers sooner
			f, err := os.Open(file)
			if err != nil {
				smiT.log.LogErrorf("Search local snapshots: failed to open snapshot file %s: %v", file, err)
				return
			}
			defer f.Close()
			snapshotInfo, err := readSnapshotInfo(f)
			if err != nil {
				smiT.log.LogErrorf("Search local snapshots: failed to read snapshot info from file %s: %v", file, err)
				return
			}
			considerSnapshotFun(snapshotInfo, constLocalAddress+file)
			snapshotCount++
		}()
	}
	smiT.log.LogDebugf("Search local snapshots: %v snapshot files found", snapshotCount)
}

func (smiT *snapshotManagerImpl) searchNetworkSnapshots(baseNetworkPaths []string, considerSnapshotFun func(SnapshotInfo, string)) {
	chainIDString := smiT.chainID.String()
	for _, baseNetworkPath := range baseNetworkPaths {
		func() { // Function to make the defers sooner
			baseNetworkPathWithChainID, err := url.JoinPath(baseNetworkPath, chainIDString)
			if err != nil {
				smiT.log.LogErrorf("Search network snapshots: unable to join paths %s and %s: %v", baseNetworkPath, chainIDString, err)
				return
			}
			scheme, basePath, err := smiT.splitURL(baseNetworkPathWithChainID)
			if err != nil {
				smiT.log.LogErrorf("Search network snapshots: unable to parse url %s: %v", baseNetworkPathWithChainID, err)
				return
			}
			reader, err := smiT.getReadCloser(scheme, "index file", basePath, constIndexFileName)
			if err != nil {
				smiT.log.LogErrorf("Search network snapshots: failed to open index file: %v", err)
				return
			}
			defer reader.Close()
			snapshotCount := 0
			scanner := bufio.NewScanner(reader) // Defaults to splitting input by newline character
			for scanner.Scan() {
				func() {
					snapshotFileName := scanner.Text()
					sReader, er := smiT.getReadCloser(scheme, "snapshot header", basePath, snapshotFileName)
					if er != nil {
						smiT.log.LogErrorf("Search network snapshots: failed to open snapshot file: %v", er)
						return
					}
					defer sReader.Close()
					snapshotInfo, er := readSnapshotInfo(sReader)
					if er != nil {
						smiT.log.LogErrorf("Search network snapshots: failed to read snapshot info from %s in %s: %v", snapshotFileName, basePath, er)
						return
					}
					baseNetworkPathSnapshot, er := url.JoinPath(baseNetworkPathWithChainID, snapshotFileName)
					if er != nil {
						smiT.log.LogErrorf("Search network snapshots: unable to join paths %s and %s: %v", baseNetworkPathWithChainID, snapshotFileName, er)
						return
					}
					considerSnapshotFun(snapshotInfo, baseNetworkPathSnapshot)
					snapshotCount++
				}()
			}
			err = scanner.Err()
			if err != nil && !errors.Is(err, io.EOF) {
				smiT.log.LogErrorf("Search network snapshots: failed read index file from %s: %v", basePath, err)
			}
			smiT.log.LogDebugf("Search network snapshots: %v snapshot files found on %s", snapshotCount, baseNetworkPath)
		}()
	}
}

func (smiT *snapshotManagerImpl) loadSnapshotFromPath(snapshotInfo SnapshotInfo, url string) error {
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
		addProgressReporterFun := func(r io.Reader, f string, s uint64) io.Reader {
			return smiT.addProgressReporter(r, fmt.Sprintf("snapshot %s", snapshotInfo), f, s)
		}
		err := DownloadToFile(smiT.ctx, url, filePathLocal, constDownloadTimeout, addProgressReporterFun)
		if err != nil {
			return err
		}
		smiT.log.LogDebugf("Loading snapshot %s from url %s: snapshot successfully downloaded to %s", snapshotInfo, url, filePathLocal)
		return loadLocalFun(filePathLocal)
	}

	scheme, path, err := smiT.splitURL(url)
	if err != nil {
		return fmt.Errorf("loading snapshot %s failed: %v", snapshotInfo, err)
	}
	switch scheme {
	case constSchemeHTTP:
		smiT.log.LogDebugf("Loading snapshot %s from url %s...", snapshotInfo, path)
		return loadNetworkFun(path)
	case constSchemeFile:
		smiT.log.LogDebugf("Loading snapshot %s from file %s...", snapshotInfo, path)
		return loadLocalFun(path)
	default:
		return fmt.Errorf("loading snapshot %s failed: unknown scheme %s in %s", snapshotInfo, scheme, url)
	}
}

func (smiT *snapshotManagerImpl) splitURL(uString string) (scheme string, path string, err error) {
	uObj, err := url.Parse(uString)
	if err != nil {
		return "", "", err
	}
	switch uObj.Scheme {
	case "http", "https":
		return constSchemeHTTP, uString, nil
	case "file":
		return constSchemeFile, filepath.Join(uObj.Host, uObj.Path), nil
	default:
		return "", "", fmt.Errorf("unknown scheme %s", uObj.Scheme)
	}
}

func (smiT *snapshotManagerImpl) getReadCloser(scheme string, fileType string, basePath string, file string) (io.ReadCloser, error) {
	switch scheme {
	case constSchemeHTTP:
		fullPath, err := url.JoinPath(basePath, file)
		if err != nil {
			return nil, fmt.Errorf("unable to join paths %s and %s: %v", basePath, file, err)
		}
		downloader, err := NewDownloaderWithTimeout(smiT.ctx, fullPath, constDownloadTimeout)
		if err != nil {
			return nil, fmt.Errorf("failed to start downloading file from url %s: %v", fullPath, err)
		}
		r := smiT.addProgressReporter(downloader, fileType, fullPath, downloader.GetLength())
		return NewReaderWithClose(r, downloader.Close), nil
	case constSchemeFile:
		fullPath := filepath.Join(basePath, file)
		f, err := os.Open(fullPath)
		if err != nil {
			return nil, fmt.Errorf("failed to open file %s", fullPath)
		}
		return f, nil
	default:
		return nil, fmt.Errorf("unnknown scheme %s", scheme)
	}
}

func (smiT *snapshotManagerImpl) addProgressReporter(r io.Reader, fileType string, url string, length uint64) io.Reader {
	progressReporter := NewProgressReporter(smiT.log, fmt.Sprintf("Downloading %s from url %s", fileType, url), length)
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
