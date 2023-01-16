package smGPAUtils

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/state"
)

type blockWAL struct {
	*logger.WrappedLogger

	dir     string
	metrics *BlockWALMetrics
}

const constFileSuffix = ".blk"

func NewBlockWAL(log *logger.Logger, baseDir string, chainID isc.ChainID, metrics *BlockWALMetrics) (BlockWAL, error) {
	dir := filepath.Join(baseDir, chainID.String())
	if err := os.MkdirAll(dir, 0o777); err != nil {
		return nil, fmt.Errorf("BlockWAL cannot create folder %v: %w", dir, err)
	}
	result := &blockWAL{
		WrappedLogger: logger.NewWrappedLogger(log.Named("wal")),
		dir:           dir,
		metrics:       metrics,
	}
	result.LogDebugf("BlockWAL created in folder %v", dir)
	return result, nil
}

// Overwrites, if block is already in WAL
func (bwT *blockWAL) Write(block state.Block) error {
	commitment := block.L1Commitment()
	fileName := fileName(commitment.BlockHash())
	filePath := filepath.Join(bwT.dir, fileName)
	bwT.LogDebugf("Writing block %s to wal; file name - %s", commitment, fileName)
	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o666)
	if err != nil {
		bwT.metrics.failedWrites.Inc()
		return fmt.Errorf("openning file %s for writing failed: %w", fileName, err)
	}
	defer f.Close()
	blockBytes := block.Bytes()
	n, err := f.Write(blockBytes)
	if err != nil {
		bwT.metrics.failedReads.Inc()
		return fmt.Errorf("writing block data to file %s failed: %w", fileName, err)
	}
	if len(blockBytes) != n {
		bwT.metrics.failedReads.Inc()
		return fmt.Errorf("only %v of total %v bytes of block were written to file %s", n, len(blockBytes), fileName)
	}
	bwT.metrics.segments.Inc()
	return nil
}

func (bwT *blockWAL) Contains(blockHash state.BlockHash) bool {
	_, err := os.Stat(filepath.Join(bwT.dir, fileName(blockHash)))
	return err == nil
}

func (bwT *blockWAL) Read(blockHash state.BlockHash) (state.Block, error) {
	fileName := fileName(blockHash)
	filePath := filepath.Join(bwT.dir, fileName)
	f, err := os.OpenFile(filePath, os.O_RDONLY, 0o666)
	if err != nil {
		bwT.metrics.failedReads.Inc()
		return nil, fmt.Errorf("opening file %s for reading failed: %w", fileName, err)
	}
	defer f.Close()
	stat, err := f.Stat()
	if err != nil {
		bwT.metrics.failedReads.Inc()
		return nil, fmt.Errorf("reading file %s information failed: %w", fileName, err)
	}
	blockBytes := make([]byte, stat.Size())
	n, err := bufio.NewReader(f).Read(blockBytes)
	if err != nil {
		bwT.metrics.failedReads.Inc()
		return nil, fmt.Errorf("reading file %s failed: %w", fileName, err)
	}
	if int64(n) != stat.Size() {
		bwT.metrics.failedReads.Inc()
		return nil, fmt.Errorf("only %v of total %v bytes of file %s were read", n, stat.Size(), fileName)
	}
	block, err := state.BlockFromBytes(blockBytes)
	if err != nil {
		bwT.metrics.failedReads.Inc()
		return nil, fmt.Errorf("error parsing block from bytes read from file %s: %w", fileName, err)
	}
	return block, nil
}

func (bwT *blockWAL) Contents() []state.BlockHash {
	result := []state.BlockHash{}
	folder, err := os.Open(bwT.dir)
	if err != nil {
		bwT.LogWarnf("Cannot open wal folder to fetch its contents: %w", err)
		return result
	}
	files, err := folder.ReadDir(0)
	if err != nil {
		bwT.LogWarnf("Cannot read contents of wal folder: %w", err)
		return result
	}
	for _, file := range files {
		fileName := file.Name()
		if file.IsDir() {
			bwT.LogWarnf("Wal folder file %s is a subfolder; ignoring it", fileName)
			continue
		}
		if !strings.HasSuffix(fileName, constFileSuffix) {
			bwT.LogWarnf("Wal folder file %s does not has expected suffix %s; ignoring it", fileName, constFileSuffix)
			continue
		}
		blockHashString := fileName[:len(fileName)-len(constFileSuffix)]
		blockHash, err := state.BlockHashFromString(blockHashString)
		if err != nil {
			bwT.LogWarnf("Wal folder file %s is not named after block hash: %w", fileName, err)
			continue
		}
		result = append(result, blockHash)
	}
	return result
}

func fileName(blockHash state.BlockHash) string {
	return blockHash.String() + constFileSuffix
}
