package sm_gpa_utils

import (
	"bufio"
	"fmt"
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
	tmpFileName := blockWALTmpFileName(commitment.BlockHash())
	tmpFilePath := filepath.Join(bwT.dir, tmpFileName)
	err := func() error { // Function is used to make defered close occur when it is needed even if write is successful
		f, err := os.OpenFile(tmpFilePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o666)
		if err != nil {
			bwT.metrics.IncFailedWrites()
			return fmt.Errorf("failed to create temporary file %s for writing block index %v: %w", tmpFileName, blockIndex, err)
		}
		defer f.Close()
		blockBytes := block.Bytes()
		n, err := f.Write(blockBytes)
		if err != nil {
			bwT.metrics.IncFailedWrites()
			return fmt.Errorf("writing block index %v data to temporary file %s failed: %w", blockIndex, tmpFileName, err)
		}
		if len(blockBytes) != n {
			bwT.metrics.IncFailedWrites()
			return fmt.Errorf("only %v of total %v bytes of block index %v were written to temporary file %s", n, len(blockBytes), blockIndex, tmpFileName)
		}
		return nil
	}()
	if err != nil {
		return err
	}
	finalFileName := blockWALFileName(commitment.BlockHash())
	finalFilePath := filepath.Join(bwT.dir, finalFileName)
	err = os.Rename(tmpFilePath, finalFilePath)
	if err != nil {
		return fmt.Errorf("failed to move temporary WAL file %s to permanent location %s: %v",
			tmpFilePath, finalFilePath, err)
	}

	bwT.metrics.BlockWritten(block.StateIndex())
	bwT.LogDebugf("Block index %v %s written to wal; file name - %s", blockIndex, commitment, finalFileName)
	return nil
}

func (bwT *blockWAL) Contains(blockHash state.BlockHash) bool {
	_, err := os.Stat(filepath.Join(bwT.dir, blockWALFileName(blockHash)))
	return err == nil
}

func (bwT *blockWAL) Read(blockHash state.BlockHash) (state.Block, error) {
	fileName := blockWALFileName(blockHash)
	filePath := filepath.Join(bwT.dir, fileName)
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
	dirEntries, err := os.ReadDir(bwT.dir)
	if err != nil {
		return err
	}
	blocksByStateIndex := map[uint32][]string{}
	for _, dirEntry := range dirEntries {
		if !dirEntry.Type().IsRegular() {
			continue
		}
		if !strings.HasSuffix(dirEntry.Name(), constBlockWALFileSuffix) {
			continue
		}
		filePath := filepath.Join(bwT.dir, dirEntry.Name())
		fileBlock, fileErr := blockFromFilePath(filePath)
		if fileErr != nil {
			bwT.metrics.IncFailedReads()
			bwT.LogWarn("Unable to read %v: %v", filePath, err)
			continue
		}
		stateIndex := fileBlock.StateIndex()
		stateIndexPaths, found := blocksByStateIndex[stateIndex]
		if found {
			stateIndexPaths = append(stateIndexPaths, filePath)
		} else {
			stateIndexPaths = []string{filePath}
		}
		blocksByStateIndex[stateIndex] = stateIndexPaths
	}
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

func blockFromFilePath(filePath string) (state.Block, error) {
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
}

func blockWALFileName(blockHash state.BlockHash) string {
	return blockHash.String() + constBlockWALFileSuffix
}

func blockWALTmpFileName(blockHash state.BlockHash) string {
	return blockWALFileName(blockHash) + constBlockWALTmpFileSuffix
}
