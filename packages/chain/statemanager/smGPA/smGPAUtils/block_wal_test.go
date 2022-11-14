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

	chainID, blocks, _, _ := GetBlocks(t, 5, 1)
	blocksInWAL := blocks[:4]
	walGood, err := NewBlockWAL(constTestFolder, chainID, log)
	require.NoError(t, err)
	walBad, err := NewBlockWAL(constTestFolder, isc.RandomChainID(), log)
	require.NoError(t, err)
	for i := range blocksInWAL {
		err = walGood.Write(blocks[i])
		require.NoError(t, err)
	}
	for i := range blocksInWAL {
		require.True(t, walGood.Contains(blocks[i].GetHash()))
		require.False(t, walBad.Contains(blocks[i].GetHash()))
	}
	require.False(t, walGood.Contains(blocks[4].GetHash()))
	require.False(t, walBad.Contains(blocks[4].GetHash()))
	for i := range blocksInWAL {
		block, err := walGood.Read(blocks[i].GetHash())
		require.NoError(t, err)
		require.True(t, blocks[i].Equals(block))
		_, err = walBad.Read(blocks[i].GetHash())
		require.Error(t, err)
	}
	_, err = walGood.Read(blocks[4].GetHash())
	require.Error(t, err)
	_, err = walBad.Read(blocks[4].GetHash())
	require.Error(t, err)
}

// Check if existing WAL record is overwritten
func TestBlockWALOverwrite(t *testing.T) {
	log := testlogger.NewLogger(t)
	defer log.Sync()
	defer cleanupAfterTest(t)

	chainID, blocks, _, _ := GetBlocks(t, 4, 1)
	wal, err := NewBlockWAL(constTestFolder, chainID, log)
	require.NoError(t, err)
	for i := range blocks {
		err = wal.Write(blocks[i])
		require.NoError(t, err)
	}
	pathFromHashFun := func(blockHash state.BlockHash) string {
		return filepath.Join(constTestFolder, chainID.String(), fileName(blockHash))
	}
	file0Path := pathFromHashFun(blocks[0].GetHash())
	file1Path := pathFromHashFun(blocks[1].GetHash())
	err = os.Rename(file1Path, file0Path)
	require.NoError(t, err)
	// block[1] is no longer in WAL
	// instead of block[0] WAL record there is block[1] record named as block[0] record

	require.True(t, wal.Contains(blocks[0].GetHash()))
	require.False(t, wal.Contains(blocks[1].GetHash()))
	block, err := wal.Read(blocks[0].GetHash())
	require.NoError(t, err)
	// blocks[0] read, but hash is of blocks[1] - the situation is as expected
	// It simulates a messed up block and checks if further Write rectifies the
	// situation
	require.True(t, blocks[1].GetHash().Equals(block.GetHash()))

	err = wal.Write(blocks[0])
	require.NoError(t, err)
	require.True(t, wal.Contains(blocks[0].GetHash()))
	block, err = wal.Read(blocks[0].GetHash())
	require.NoError(t, err)
	require.True(t, blocks[0].GetHash().Equals(block.GetHash()))
	require.True(t, blocks[0].Equals(block))
}

// Check if after restart wal is functioning correctly
func TestBlockWALRestart(t *testing.T) {
	log := testlogger.NewLogger(t)
	defer log.Sync()
	defer cleanupAfterTest(t)

	chainID, blocks, _, _ := GetBlocks(t, 4, 1)
	wal, err := NewBlockWAL(constTestFolder, chainID, log)
	require.NoError(t, err)
	for i := range blocks {
		err = wal.Write(blocks[i])
		require.NoError(t, err)
	}

	//Restart: WAL object is recreated
	wal, err = NewBlockWAL(constTestFolder, chainID, log)
	require.NoError(t, err)
	for i := range blocks {
		require.True(t, wal.Contains(blocks[i].GetHash()))
		block, err := wal.Read(blocks[i].GetHash())
		require.NoError(t, err)
		require.True(t, blocks[i].Equals(block))
	}
}

func cleanupAfterTest(t *testing.T) {
	err := os.RemoveAll(constTestFolder)
	require.NoError(t, err)
}
