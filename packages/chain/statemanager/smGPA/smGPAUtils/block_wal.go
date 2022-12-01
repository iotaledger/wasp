package smGPAUtils

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"

	"github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/state"
)

type blockWAL struct {
	dir     string
	log     *logger.Logger
	metrics *blockWALMetrics
}

func NewBlockWAL(baseDir string, chainID *isc.ChainID, log *logger.Logger) (BlockWAL, error) {
	dir := filepath.Join(baseDir, chainID.String())
	if err := os.MkdirAll(dir, 0o777); err != nil {
		return nil, fmt.Errorf("BlockWAL cannot create folder %v: %w", dir, err)
	}
	result := &blockWAL{
		dir:     dir,
		metrics: newBlockWALMetrics(),
		log:     log.Named("wal"),
	}
	result.log.Debugf("BlockWAL created in folder %v", dir)
	return result, nil
}

// Overwrites, if block is already in WAL
func (bwT *blockWAL) Write(block state.Block) error {
	blockHash := block.GetHash()
	fileName := fileName(blockHash)
	filePath := filepath.Join(bwT.dir, fileName)
	bwT.log.Debugf("Writing block %s to wal; file name - %s", blockHash, fileName)
	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o666)
	if err != nil {
		bwT.metrics.failedWrites.Inc()
		return fmt.Errorf("Openning file %s for writing failed: %w", fileName, err)
	}
	defer f.Close()
	blockBytes := block.Bytes()
	n, err := f.Write(blockBytes)
	if err != nil {
		bwT.metrics.failedReads.Inc()
		return fmt.Errorf("Writing block data to file %s failed: %w", fileName, err)
	}
	if len(blockBytes) != n {
		bwT.metrics.failedReads.Inc()
		return fmt.Errorf("Only %v of total %v bytes of block were written to file %s", n, len(blockBytes), fileName)
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
		return nil, fmt.Errorf("Opening file %s for reading failed: %w", fileName, err)
	}
	defer f.Close()
	stat, err := f.Stat()
	if err != nil {
		bwT.metrics.failedReads.Inc()
		return nil, fmt.Errorf("Reading file %s information failed: %w", fileName, err)
	}
	blockBytes := make([]byte, stat.Size())
	n, err := bufio.NewReader(f).Read(blockBytes)
	if err != nil {
		bwT.metrics.failedReads.Inc()
		return nil, fmt.Errorf("Reading file %s failed: %w", fileName, err)
	}
	if int64(n) != stat.Size() {
		bwT.metrics.failedReads.Inc()
		return nil, fmt.Errorf("Only %v of total %v bytes of file %s were read", n, stat.Size(), fileName)
	}
	block, err := state.BlockFromBytes(blockBytes)
	if err != nil {
		bwT.metrics.failedReads.Inc()
		return nil, fmt.Errorf("Error parsing block from bytes read from file %s: %w", fileName, err)
	}
	return block, nil
}

func fileName(blockHash state.BlockHash) string {
	return blockHash.String() + ".blk"
}
