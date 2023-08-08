package sm_gpa_utils

import (
	//"bufio"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/samber/lo"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/runtime/ioutils"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/metrics"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

type blockWAL struct {
	*logger.WrappedLogger

	dir     string
	metrics *metrics.ChainBlockWALMetrics
}

const (
	constBlockWALFileSuffix    = ".blk"
	constBlockWALTmpFileSuffix = ".tmp"
)

func NewBlockWAL(log *logger.Logger, baseDir string, chainID isc.ChainID, metrics *metrics.ChainBlockWALMetrics) (BlockWAL, error) {
	dir := filepath.Join(baseDir, chainID.String())
	if err := ioutils.CreateDirectory(dir, 0o777); err != nil {
		return nil, fmt.Errorf("BlockWAL cannot create folder %v: %w", dir, err)
	}

	result := &blockWAL{
		WrappedLogger: logger.NewWrappedLogger(log.Named("WAL")),
		dir:           dir,
		metrics:       metrics,
	}
	result.LogDebugf("BlockWAL created in folder %v", dir)
	return result, nil
}

// Overwrites, if block is already in WAL
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
		ww.WriteUint32(blockIndex)
		if ww.Err != nil {
			bwT.metrics.IncFailedWrites()
			return fmt.Errorf("failed to write block index into temporary file %s: %w", tmpFilePath, ww.Err)
		}
		err = block.Write(f)
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
	block, err := blockFromFilePath(filePath)
	if err != nil {
		bwT.metrics.IncFailedReads()
		return nil, err
	}
	return block, nil
}

// This reads all the existing blocks from the WAL dir and passes them to the supplied callback.
// The blocks are provided ordered by the state index, so that they can be applied to the store.
// This function reads blocks twice, but tries to minimize the amount of memory required to load the WAL.
func (bwT *blockWAL) ReadAllByStateIndex(cb func(stateIndex uint32, block state.Block) bool) error {
	blocksByStateIndex := map[uint32][]string{}
	checkFile := func(filePath string) {
		if !strings.HasSuffix(filePath, constBlockWALFileSuffix) {
			return
		}
		stateIndex, err := blockIndexFromFilePath(filePath)
		if err != nil {
			bwT.metrics.IncFailedReads()
			bwT.LogWarn("Unable to read %v: %v", filePath, err)
			return
		}
		stateIndexPaths, found := blocksByStateIndex[stateIndex]
		if found {
			stateIndexPaths = append(stateIndexPaths, filePath)
		} else {
			stateIndexPaths = []string{filePath}
		}
		blocksByStateIndex[stateIndex] = stateIndexPaths
	}

	var checkDir func(dirPath string, dirEntries []os.DirEntry)
	checkDir = func(dirPath string, dirEntries []os.DirEntry) {
		for _, dirEntry := range dirEntries {
			entryPath := filepath.Join(dirPath, dirEntry.Name())
			if dirEntry.IsDir() {
				subDirEntries, err := os.ReadDir(entryPath)
				if err == nil {
					checkDir(entryPath, subDirEntries)
				}
			} else {
				checkFile(entryPath)
			}
		}
	}

	dirEntries, err := os.ReadDir(bwT.dir)
	if err != nil {
		return err
	}
	checkDir(bwT.dir, dirEntries)

	allStateIndexes := lo.Keys(blocksByStateIndex)
	sort.Slice(allStateIndexes, func(i, j int) bool { return allStateIndexes[i] < allStateIndexes[j] })
	for _, stateIndex := range allStateIndexes {
		stateIndexPaths := blocksByStateIndex[stateIndex]
		for _, stateIndexPath := range stateIndexPaths {
			fileBlock, fileErr := blockFromFilePath(stateIndexPath)
			if fileErr != nil {
				bwT.metrics.IncFailedReads()
				bwT.LogWarn("Unable to read %v: %v", stateIndexPath, err)
				continue
			}
			if !cb(stateIndex, fileBlock) {
				return nil
			}
		}
	}
	return nil
}

func blockInfoFromFilePath[I any](filePath string, getInfoFun func(io.Reader) (I, error)) (I, error) {
	f, err := os.OpenFile(filePath, os.O_RDONLY, 0o666)
	if err != nil {
		var info I
		return info, fmt.Errorf("opening file %s for reading failed: %w", filePath, err)
	}
	defer f.Close()
	return getInfoFun(f)
}

func blockIndexFromFilePath(filePath string) (uint32, error) {
	return blockInfoFromFilePath(filePath, blockIndexFromReader)
}

func blockFromFilePath(filePath string) (state.Block, error) {
	return blockInfoFromFilePath(filePath, blockFromReader)
}

func blockIndexFromReader(r io.Reader) (uint32, error) {
	rr := rwutil.NewReader(r)
	info := rr.ReadUint32()
	return info, rr.Err
}

func blockFromReader(r io.Reader) (state.Block, error) {
	blockIndex, err := blockIndexFromReader(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read block index in header: %w", err)
	}
	block := state.NewBlock()
	err = block.Read(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read block: %w", err)
	}
	if blockIndex != block.StateIndex() {
		return nil, fmt.Errorf("block index in header %v does not match block index in block %v",
			blockIndex, block.StateIndex())
	}
	return block, nil
}

/*func blockFromFilePath(filePath string) (state.Block, error) {
	f, err := os.OpenFile(filePath, os.O_RDONLY, 0o666)
	if err != nil {
		return nil, fmt.Errorf("opening file %s for reading failed: %w", filePath, err)
	}
	defer f.Close()
	stat, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("reading file %s information failed: %w", filePath, err)
	}
	blockBytes := make([]byte, stat.Size())
	n, err := bufio.NewReader(f).Read(blockBytes)
	if err != nil {
		return nil, fmt.Errorf("reading file %s failed: %w", filePath, err)
	}
	if int64(n) != stat.Size() {
		return nil, fmt.Errorf("only %v of total %v bytes of file %s were read", n, stat.Size(), filePath)
	}
	block, err := state.BlockFromBytes(blockBytes)
	if err != nil {
		return nil, fmt.Errorf("error parsing block from bytes read from file %s: %w", filePath, err)
	}
	return block, nil
}*/

func blockWALSubFolderName(blockHash state.BlockHash) string {
	return hex.EncodeToString(blockHash[:1])
}

func blockWALFileName(blockHash state.BlockHash) string {
	return blockHash.String() + constBlockWALFileSuffix
}

func blockWALTmpFileName(blockHash state.BlockHash) string {
	return blockWALFileName(blockHash) + constBlockWALTmpFileSuffix
}
