package smGPAUtils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
)

const constTestFolder = "basicWALTest"

func TestBlockWALBasic(t *testing.T) {
	log := testlogger.NewLogger(t)
	defer log.Sync()
	defer cleanupAfterTest(t)

	factory := NewBlockFactory(t)
	blocks, _ := factory.GetBlocks(5, 1)
	blocksInWAL := blocks[:4]
	walGood, err := NewBlockWAL(log, constTestFolder, factory.GetChainID(), NewBlockWALMetrics())
	require.NoError(t, err)
	walBad, err := NewBlockWAL(log, constTestFolder, isc.RandomChainID(), NewBlockWALMetrics())
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
		block, err := walGood.Read(blocks[i].Hash())
		require.NoError(t, err)
		require.True(t, blocks[i].Hash().Equals(block.Hash())) // Should be Equals instead of Hash().Equals()
		_, err = walBad.Read(blocks[i].Hash())
		require.Error(t, err)
	}
	_, err = walGood.Read(blocks[4].Hash())
	require.Error(t, err)
	_, err = walBad.Read(blocks[4].Hash())
	require.Error(t, err)
}

// Check if existing WAL record is overwritten
func TestBlockWALOverwrite(t *testing.T) {
	log := testlogger.NewLogger(t)
	defer log.Sync()
	defer cleanupAfterTest(t)

	factory := NewBlockFactory(t)
	blocks, _ := factory.GetBlocks(4, 1)
	wal, err := NewBlockWAL(log, constTestFolder, factory.GetChainID(), NewBlockWALMetrics())
	require.NoError(t, err)
	for i := range blocks {
		err = wal.Write(blocks[i])
		require.NoError(t, err)
	}
	pathFromHashFun := func(blockHash state.BlockHash) string {
		return filepath.Join(constTestFolder, factory.GetChainID().String(), fileName(blockHash))
	}
	file0Path := pathFromHashFun(blocks[0].Hash())
	file1Path := pathFromHashFun(blocks[1].Hash())
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
	require.True(t, blocks[0].Hash().Equals(block.Hash()))
	//require.True(t, blocks[0].Equals(block))
}

// Check if after restart wal is functioning correctly
func TestBlockWALRestart(t *testing.T) {
	log := testlogger.NewLogger(t)
	defer log.Sync()
	defer cleanupAfterTest(t)

	factory := NewBlockFactory(t)
	blocks, _ := factory.GetBlocks(4, 1)
	wal, err := NewBlockWAL(log, constTestFolder, factory.GetChainID(), NewBlockWALMetrics())
	require.NoError(t, err)
	for i := range blocks {
		err = wal.Write(blocks[i])
		require.NoError(t, err)
	}

	//Restart: WAL object is recreated
	wal, err = NewBlockWAL(log, constTestFolder, factory.GetChainID(), NewBlockWALMetrics())
	require.NoError(t, err)
	for i := range blocks {
		require.True(t, wal.Contains(blocks[i].Hash()))
		block, err := wal.Read(blocks[i].Hash())
		require.NoError(t, err)
		require.True(t, blocks[i].Hash().Equals(block.Hash())) // Should be Equals instead of Hash().Equals()
	}
}

func cleanupAfterTest(t *testing.T) {
	err := os.RemoveAll(constTestFolder)
	require.NoError(t, err)
}
