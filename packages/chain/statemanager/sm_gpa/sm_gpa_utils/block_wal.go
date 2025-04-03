package sm_gpa_utils

import (
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/samber/lo"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/hive.go/log"
	"github.com/iotaledger/hive.go/runtime/ioutils"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/metrics"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

type blockWAL struct {
	log.Logger

	dir     string
	metrics *metrics.ChainBlockWALMetrics
}

const (
	constBlockWALFileSuffix    = ".blk"
	constBlockWALTmpFileSuffix = ".tmp"
)

func NewBlockWAL(log log.Logger, baseDir string, chainID isc.ChainID, metrics *metrics.ChainBlockWALMetrics) (BlockWAL, error) {
	dir := filepath.Join(baseDir, chainID.String())
	if err := ioutils.CreateDirectory(dir, 0o777); err != nil {
		return nil, fmt.Errorf("BlockWAL cannot create folder %v: %w", dir, err)
	}

	result := &blockWAL{
		Logger:  log.NewChildLogger("WAL"),
		dir:     dir,
		metrics: metrics,
	}
	result.LogDebugf("BlockWAL created in folder %v", dir)
	return result, nil
}

// Overwrites, if block is already in WAL
// Block format (version 1):
//   - Version (4 bytes, unsigned int); value 1
//   - State index (4 bytes, unsigned int)
//   - Block bytes
//
// Block format (legacy = version 0):
//   - Block bytes
func (bwT *blockWAL) Write(block state.Block) error {
	blockIndex := block.StateIndex()
	commitment := block.L1Commitment()
	subfolderName := blockWALSubFolderName(commitment.BlockHash())
	folderPath := filepath.Join(bwT.dir, subfolderName)
	if err := ioutils.CreateDirectory(folderPath, 0o777); err != nil {
		return fmt.Errorf("failed to create folder %s for writing block: %w", folderPath, err)
	}
	tmpFileName := blockWALTmpFileName(commitment.BlockHash())
	tmpFilePath := filepath.Join(folderPath, tmpFileName)
	err := func() error { // Function is used to make defered close occur when it is needed even if write is successful
		f, err := os.OpenFile(tmpFilePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o666)
		if err != nil {
			bwT.metrics.IncFailedWrites()
			return fmt.Errorf("failed to create temporary file %s for writing block: %w", tmpFilePath, err)
		}
		defer f.Close()
		ww := rwutil.NewWriter(f)
		ww.WriteUint32(1) // Version; 4 bytes (instead of just 1) to lower number of possible collisions with legacy WAL format
		ww.WriteUint32(blockIndex)
		if ww.Err != nil {
			bwT.metrics.IncFailedWrites()
			return fmt.Errorf("failed to write block index into temporary file %s: %w", tmpFilePath, ww.Err)
		}
		err = bcs.MarshalStream(&block, f)
		if err != nil {
			bwT.metrics.IncFailedWrites()
			return fmt.Errorf("writing block to temporary file %s failed: %w", tmpFilePath, err)
		}
		return nil
	}()
	if err != nil {
		return err
	}
	finalFileName := blockWALFileName(commitment.BlockHash())
	finalFilePath := filepath.Join(folderPath, finalFileName)
	err = os.Rename(tmpFilePath, finalFilePath)
	if err != nil {
		return fmt.Errorf("failed to move temporary WAL file %s to permanent location %s: %v",
			tmpFilePath, finalFilePath, err)
	}

	bwT.metrics.BlockWritten(block.StateIndex())
	bwT.LogDebugf("Block index %v %s written to wal; file name - %s", blockIndex, commitment, finalFilePath)
	return nil
}

func (bwT *blockWAL) blockFilepath(blockHash state.BlockHash) (string, bool) {
	subfolderName := blockWALSubFolderName(blockHash)
	fileName := blockWALFileName(blockHash)

	pathWithSubFolder := filepath.Join(bwT.dir, subfolderName, fileName)
	_, err := os.Stat(pathWithSubFolder)
	if err == nil {
		return pathWithSubFolder, true
	}

	// Checked for backward compatibility and for ease of adding some blocks from other sources
	pathNoSubFolder := filepath.Join(bwT.dir, fileName)
	_, err = os.Stat(pathNoSubFolder)
	if err == nil {
		return pathNoSubFolder, true
	}
	return "", false
}

func (bwT *blockWAL) Contains(blockHash state.BlockHash) bool {
	_, exists := bwT.blockFilepath(blockHash)
	return exists
}

func (bwT *blockWAL) Read(blockHash state.BlockHash) (state.Block, error) {
	filePath, exists := bwT.blockFilepath(blockHash)
	if !exists {
		return nil, fmt.Errorf("block hash %s is not present in WAL", blockHash)
	}
	block, err := BlockFromFilePath(filePath)
	if err != nil {
		bwT.metrics.IncFailedReads()
		return nil, err
	}
	return block, nil
}

// This reads all the existing blocks from the WAL dir and passes them to the supplied callback.
// The blocks are provided ordered by the state index, so that they can be applied to the store.
// This function reads blocks twice, but tries to minimize the amount of memory required to load the WAL.
func (bwT *blockWAL) ReadAllByStateIndex(cb func(stateIndex uint32, block state.Block) bool) error { //nolint:funlen
	bwT.LogDebugf("Reading entire WAL...")
	blocksByStateIndex := map[uint32][]string{}
	checkFile := func(filePath string) bool {
		if !strings.HasSuffix(filePath, constBlockWALFileSuffix) {
			return false
		}
		stateIndex, err := BlockIndexFromFilePath(filePath)
		if err != nil {
			bwT.metrics.IncFailedReads()
			bwT.LogWarn("Reading entire WAL: unable to read block index from %v: %v", filePath, err)
			return false
		}
		stateIndexPaths, found := blocksByStateIndex[stateIndex]
		if found {
			stateIndexPaths = append(stateIndexPaths, filePath)
		} else {
			stateIndexPaths = []string{filePath}
		}
		blocksByStateIndex[stateIndex] = stateIndexPaths
		return true
	}

	blocksTotal := 0
	var checkDir func(dirPath string, dirEntries []os.DirEntry)
	checkDir = func(dirPath string, dirEntries []os.DirEntry) {
		bwT.LogDebugf("Reading entire WAL: checking folder %v with %v entries...", dirPath, len(dirEntries))
		entries := 0
		blocks := 0
		for _, dirEntry := range dirEntries {
			entryPath := filepath.Join(dirPath, dirEntry.Name())
			if dirEntry.IsDir() {
				subDirEntries, err := os.ReadDir(entryPath)
				if err == nil {
					checkDir(entryPath, subDirEntries)
				}
			} else {
				isBlock := checkFile(entryPath)
				if isBlock {
					blocks++
					blocksTotal++
				}
			}
			entries++
			if entries%1000 == 0 {
				bwT.LogDebugf("Reading entire WAL: checking folder %v: %v entries checked, %v blocks found (%v in total)...",
					dirPath, entries, blocks, blocksTotal)
			}
		}
		bwT.LogDebugf("Reading entire WAL: checking folder %v completed: %v entries checked, %v blocks found (%v in total)",
			dirPath, entries, blocks, blocksTotal)
	}

	dirEntries, err := os.ReadDir(bwT.dir)
	if err != nil {
		return err
	}
	checkDir(bwT.dir, dirEntries)

	bwT.LogDebugf("Reading entire WAL: %v blocks found, sorting them by index...", blocksTotal)
	allStateIndexes := lo.Keys(blocksByStateIndex)
	slices.Sort(allStateIndexes)
	bwT.LogDebugf("Reading entire WAL: blocks sorted, notifying caller...")
	for _, stateIndex := range allStateIndexes {
		stateIndexPaths := blocksByStateIndex[stateIndex]
		for _, stateIndexPath := range stateIndexPaths {
			fileBlock, fileErr := BlockFromFilePath(stateIndexPath)
			if fileErr != nil {
				bwT.metrics.IncFailedReads()
				bwT.LogWarn("Reading entire WAL: unable to read block from %v: %v", stateIndexPath, err)
				continue
			}
			if !cb(stateIndex, fileBlock) {
				return nil
			}
		}
	}
	bwT.LogDebugf("Reading entire WAL completed")
	return nil
}

func blockInfoFromFilePath[I any](filePath string, getInfoFun func(uint32, io.Reader) (I, error)) (I, error) {
	f, err := os.OpenFile(filePath, os.O_RDONLY, 0o666)
	var info I
	if err != nil {
		return info, fmt.Errorf("opening file %s for reading failed: %w", filePath, err)
	}
	defer f.Close()
	rr := rwutil.NewReader(f)
	version := rr.ReadUint32()
	if rr.Err != nil {
		return info, fmt.Errorf("failed reading file version: %w", rr.Err)
	}
	var errV error
	if version == 1 {
		info, errV = getInfoFun(version, f)
		if errV == nil {
			return info, nil
		}
		// error reading as version 1, maybe it's legacy version?
	}
	// backwards compatibility - reading legacy version
	// NOTE: reopening file, because version bytes (or possibly more) has already been read
	f, err = os.OpenFile(filePath, os.O_RDONLY, 0o666)
	if err != nil {
		return info, fmt.Errorf("reopening file %s for reading failed: %w", filePath, err)
	}
	defer f.Close()
	info, err = getInfoFun(0, f)
	if errV == nil {
		return info, err
	}
	return info, fmt.Errorf("version %v error: %w, legacy version error: %w", version, errV, err)
}

func BlockIndexFromFilePath(filePath string) (uint32, error) {
	return blockInfoFromFilePath(filePath, blockIndexFromReader)
}

func BlockFromFilePath(filePath string) (state.Block, error) {
	return blockInfoFromFilePath(filePath, blockFromReader)
}

func blockIndexFromReader(version uint32, r io.Reader) (uint32, error) {
	switch version {
	case 1:
		rr := rwutil.NewReader(r)
		index := rr.ReadUint32()
		return index, rr.Err
	case 0:
		block := state.NewBlock()
		_, err := bcs.UnmarshalStreamInto(r, &block)
		if err != nil {
			return 0, err
		}
		return block.StateIndex(), nil
	default:
		return 0, fmt.Errorf("unknown block version %v", version)
	}
}

func blockFromReader(version uint32, r io.Reader) (state.Block, error) {
	switch version {
	case 1:
		blockIndex, err := blockIndexFromReader(version, r)
		if err != nil {
			return nil, fmt.Errorf("failed to read block index in header: %w", err)
		}
		block := state.NewBlock()
		_, err = bcs.UnmarshalStreamInto(r, &block)
		if err != nil {
			return nil, fmt.Errorf("failed to read block: %w", err)
		}
		if blockIndex != block.StateIndex() {
			return nil, fmt.Errorf("block index in header %v does not match block index in block %v",
				blockIndex, block.StateIndex())
		}
		return block, nil
	case 0:
		block := state.NewBlock()
		_, err := bcs.UnmarshalStreamInto(r, &block)
		return block, err
	default:
		return nil, fmt.Errorf("unknown block version %v", version)
	}
}

func blockWALSubFolderName(blockHash state.BlockHash) string {
	return hex.EncodeToString(blockHash[:1])
}

func blockWALFileName(blockHash state.BlockHash) string {
	return blockHash.String() + constBlockWALFileSuffix
}

func blockWALTmpFileName(blockHash state.BlockHash) string {
	return blockWALFileName(blockHash) + constBlockWALTmpFileSuffix
}
