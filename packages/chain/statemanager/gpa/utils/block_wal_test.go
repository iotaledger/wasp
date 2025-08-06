package utils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/v2/packages/isc/isctest"
	"github.com/iotaledger/wasp/v2/packages/state"
	"github.com/iotaledger/wasp/v2/packages/testutil/testlogger"
)

const constTestFolder = "basicWALTest"

func TestBlockWALBasic(t *testing.T) {
	log := testlogger.NewLogger(t)
	defer log.Shutdown()
	defer cleanupAfterTest(t)

	factory := NewBlockFactory(t)
	blocks := factory.GetBlocks(5, 1)
	blocksInWAL := blocks[:4]
	walGood, err := NewBlockWAL(log, constTestFolder, factory.GetChainID(), mockBlockWALMetrics())
	require.NoError(t, err)
	walBad, err := NewBlockWAL(log, constTestFolder, isctest.RandomChainID(), mockBlockWALMetrics())
	require.NoError(t, err)
	for i := range blocksInWAL {
		err = walGood.Write(blocks[i])
		require.NoError(t, err)
	}
	for i := range blocksInWAL {
		require.True(t, walGood.Contains(blocks[i].Hash()))
		require.False(t, walBad.Contains(blocks[i].Hash()))
	}
	require.False(t, walGood.Contains(blocks[4].Hash()))
	require.False(t, walBad.Contains(blocks[4].Hash()))
	for i := range blocksInWAL {
		block, err2 := walGood.Read(blocks[i].Hash())
		require.NoError(t, err2)
		require.True(t, blocks[i].Equals(block))
		_, err2 = walBad.Read(blocks[i].Hash())
		require.Error(t, err2)
	}
	_, err = walGood.Read(blocks[4].Hash())
	require.Error(t, err)
	_, err = walBad.Read(blocks[4].Hash())
	require.Error(t, err)
}

// Check if block prior to version 1 is read (that has no version data)
func TestBlockWALLegacy(t *testing.T) {
	log := testlogger.NewLogger(t)
	defer log.Shutdown()
	defer cleanupAfterTest(t)

	factory := NewBlockFactory(t)
	blocks := factory.GetBlocks(4, 1)
	wal, err := NewBlockWAL(log, constTestFolder, factory.GetChainID(), mockBlockWALMetrics())
	require.NoError(t, err)
	writeBlocksLegacy(t, factory.GetChainID(), blocks)
	for i := range blocks {
		block, err := wal.Read(blocks[i].Hash())
		require.NoError(t, err)
		require.True(t, blocks[i].Equals(block))
	}
}

// Check if existing block in WAL is found even if it is not in a subfolder
func TestBlockWALNoSubfolder(t *testing.T) {
	log := testlogger.NewLogger(t)
	defer log.Shutdown()
	defer cleanupAfterTest(t)

	factory := NewBlockFactory(t)
	blocks := factory.GetBlocks(4, 1)
	wal, err := NewBlockWAL(log, constTestFolder, factory.GetChainID(), mockBlockWALMetrics())
	require.NoError(t, err)
	for i := range blocks {
		err = wal.Write(blocks[i])
		require.NoError(t, err)
	}
	for _, block := range blocks {
		pathWithSubfolder := walPathFromHash(factory.GetChainID(), block.Hash())
		pathNoSubfolder := walPathNoSubfolderFromHash(factory.GetChainID(), block.Hash())
		err = os.Rename(pathWithSubfolder, pathNoSubfolder)
		require.NoError(t, err)
	}
	for _, block := range blocks {
		require.True(t, wal.Contains(block.Hash()))
		blockRead, err := wal.Read(block.Hash())
		require.NoError(t, err)
		require.True(t, block.Equals(blockRead))
	}
}

// Check if existing WAL record is overwritten
func TestBlockWALOverwrite(t *testing.T) {
	log := testlogger.NewLogger(t)
	defer log.Shutdown()
	defer cleanupAfterTest(t)

	factory := NewBlockFactory(t)
	blocks := factory.GetBlocks(4, 1)
	wal, err := NewBlockWAL(log, constTestFolder, factory.GetChainID(), mockBlockWALMetrics())
	require.NoError(t, err)
	for i := range blocks {
		err = wal.Write(blocks[i])
		require.NoError(t, err)
	}
	file0Path := walPathFromHash(factory.GetChainID(), blocks[0].Hash())
	file1Path := walPathFromHash(factory.GetChainID(), blocks[1].Hash())
	err = os.Rename(file1Path, file0Path)
	require.NoError(t, err)
	// block[1] is no longer in WAL
	// instead of block[0] WAL record there is block[1] record named as block[0] record

	require.True(t, wal.Contains(blocks[0].Hash()))
	require.False(t, wal.Contains(blocks[1].Hash()))
	block, err := wal.Read(blocks[0].Hash())
	require.NoError(t, err)
	// blocks[0] read, but hash is of blocks[1] - the situation is as expected
	// It simulates a messed up block and checks if further Write rectifies the
	// situation
	require.True(t, blocks[1].Hash().Equals(block.Hash()))

	err = wal.Write(blocks[0])
	require.NoError(t, err)
	require.True(t, wal.Contains(blocks[0].Hash()))
	block, err = wal.Read(blocks[0].Hash())
	require.NoError(t, err)
	require.True(t, blocks[0].Equals(block))
}

// Check if after restart wal is functioning correctly
func TestBlockWALRestart(t *testing.T) {
	log := testlogger.NewLogger(t)
	defer log.Shutdown()
	defer cleanupAfterTest(t)

	factory := NewBlockFactory(t)
	blocks := factory.GetBlocks(4, 1)
	wal, err := NewBlockWAL(log, constTestFolder, factory.GetChainID(), mockBlockWALMetrics())
	require.NoError(t, err)
	for i := range blocks {
		err = wal.Write(blocks[i])
		require.NoError(t, err)
	}

	// Restart: WAL object is recreated
	wal, err = NewBlockWAL(log, constTestFolder, factory.GetChainID(), mockBlockWALMetrics())
	require.NoError(t, err)
	for i := range blocks {
		require.True(t, wal.Contains(blocks[i].Hash()))
		block, err := wal.Read(blocks[i].Hash())
		require.NoError(t, err)
		require.True(t, blocks[i].Equals(block))
	}
}

func testReadAllByStateIndex(t *testing.T, addToWALFun func(BlockWAL, []state.Block)) {
	log := testlogger.NewLogger(t)
	defer log.Shutdown()
	defer cleanupAfterTest(t)

	factory := NewBlockFactory(t)
	mainBlocks := 50
	branchBlocks := 20
	branchBlockIndex := mainBlocks - branchBlocks - 1
	blocksMain := factory.GetBlocks(mainBlocks, 1)
	blocksBranch := factory.GetBlocksFrom(branchBlocks, 1, blocksMain[branchBlockIndex].L1Commitment(), 2)
	wal, err := NewBlockWAL(log, constTestFolder, factory.GetChainID(), mockBlockWALMetrics())
	require.NoError(t, err)
	addToWALFun(factory.GetChainID(), wal, blocksMain)
	addToWALFun(factory.GetChainID(), wal, blocksBranch)

	var blocksRead []state.Block
	err = wal.ReadAllByStateIndex(func(stateIndex uint32, block state.Block) bool {
		require.Equal(t, stateIndex, block.StateIndex())
		blocksRead = append(blocksRead, block)
		return true
	})
	require.NoError(t, err)

	for i := 0; i <= branchBlockIndex; i++ {
		require.Equal(t, uint32(i+1), blocksRead[i].StateIndex())
		require.True(t, blocksMain[i].Equals(blocksRead[i]))
	}
	for i := branchBlockIndex + 1; i < mainBlocks; i++ {
		blocksReadIndex := i*2 - branchBlockIndex - 1
		block1 := blocksRead[blocksReadIndex]
		block2 := blocksRead[blocksReadIndex+1]
		require.Equal(t, uint32(i+1), block1.StateIndex())
		require.Equal(t, uint32(i+1), block2.StateIndex())
		if !blocksMain[i].L1Commitment().Equals(block1.L1Commitment()) {
			block1, block2 = block2, block1
		}
		require.True(t, blocksMain[i].Equals(block1))
		require.True(t, blocksBranch[i-branchBlockIndex-1].Equals(block2))
	}
}

func TestReadAllByStateIndexV1(t *testing.T) {
	testReadAllByStateIndex(t, func(wal BlockWAL, blocks []state.Block) {
		for _, block := range blocks {
			err := wal.Write(block)
			require.NoError(t, err)
		}
	})
}

func TestReadAllByStateIndexLegacy(t *testing.T) {
	testReadAllByStateIndex(t, func(wal BlockWAL, blocks []state.Block) {
		writeBlocksLegacy(t, blocks)
	})
}

func walPathFromHash(blockHash state.BlockHash) string {
	return filepath.Join(constTestFolder, chainID.String(), blockWALSubFolderName(blockHash), blockWALFileName(blockHash))
}

func walPathNoSubfolderFromHash(blockHash state.BlockHash) string {
	return filepath.Join(constTestFolder, chainID.String(), blockWALFileName(blockHash))
}

func writeBlocksLegacy(t *testing.T, blocks []state.Block) {
	for _, block := range blocks {
		filePath := walPathNoSubfolderFromHash(block.Hash())
		f, err := os.OpenFile(filePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o666)
		require.NoError(t, err)
		err = bcs.MarshalStream(&block, f)
		require.NoError(t, err)
		err = f.Close()
		require.NoError(t, err)
	}
}

func cleanupAfterTest(t *testing.T) {
	err := os.RemoveAll(constTestFolder)
	require.NoError(t, err)
}
