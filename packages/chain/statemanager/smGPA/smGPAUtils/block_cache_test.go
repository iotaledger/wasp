// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package smGPAUtils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
)

func TestBlockCacheSimple(t *testing.T) {
	log := testlogger.NewLogger(t)
	defer log.Sync()

	factory := NewBlockFactory(t)
	blocks := factory.GetBlocks(4, 1)
	blockCache, err := NewBlockCache(NewDefaultTimeProvider(), 100, NewEmptyTestBlockWAL(), log)
	require.NoError(t, err)
	blockCache.AddBlock(blocks[0])
	blockCache.AddBlock(blocks[1])
	blockCache.AddBlock(blocks[2])
	require.NotNil(t, blockCache.GetBlock(blocks[0].L1Commitment()))
	require.NotNil(t, blockCache.GetBlock(blocks[2].L1Commitment()))
	require.NotNil(t, blockCache.GetBlock(blocks[1].L1Commitment()))
	require.Nil(t, blockCache.GetBlock(blocks[3].L1Commitment()))
}

func TestBlockCacheCleaning(t *testing.T) {
	log := testlogger.NewLogger(t)
	defer log.Sync()

	factory := NewBlockFactory(t)
	blocks := factory.GetBlocks(6, 2)
	blockCache, err := NewBlockCache(NewDefaultTimeProvider(), 100, NewEmptyTestBlockWAL(), log)
	require.NoError(t, err)
	beforeTime := time.Now()
	blockCache.AddBlock(blocks[0])
	blockCache.AddBlock(blocks[1])
	require.NotNil(t, blockCache.GetBlock(blocks[0].L1Commitment()))
	require.NotNil(t, blockCache.GetBlock(blocks[1].L1Commitment()))
	blockCache.CleanOlderThan(beforeTime)
	require.NotNil(t, blockCache.GetBlock(blocks[0].L1Commitment()))
	require.NotNil(t, blockCache.GetBlock(blocks[1].L1Commitment()))
	blockCache.CleanOlderThan(time.Now())
	require.Nil(t, blockCache.GetBlock(blocks[0].L1Commitment()))
	require.Nil(t, blockCache.GetBlock(blocks[1].L1Commitment()))
	blockCache.AddBlock(blocks[2])
	blockCache.AddBlock(blocks[3])
	inTheMiddleTime := time.Now()
	blockCache.AddBlock(blocks[4])
	blockCache.AddBlock(blocks[5])
	require.NotNil(t, blockCache.GetBlock(blocks[2].L1Commitment()))
	require.NotNil(t, blockCache.GetBlock(blocks[3].L1Commitment()))
	require.NotNil(t, blockCache.GetBlock(blocks[4].L1Commitment()))
	require.NotNil(t, blockCache.GetBlock(blocks[5].L1Commitment()))
	blockCache.CleanOlderThan(inTheMiddleTime)
	require.Nil(t, blockCache.GetBlock(blocks[2].L1Commitment()))
	require.Nil(t, blockCache.GetBlock(blocks[3].L1Commitment()))
	require.NotNil(t, blockCache.GetBlock(blocks[4].L1Commitment()))
	require.NotNil(t, blockCache.GetBlock(blocks[5].L1Commitment()))
}

// Test if blocks are put/taken from WAL, if they are not available in cache
func TestBlockCacheWAL(t *testing.T) {
	log := testlogger.NewLogger(t)
	defer log.Sync()

	factory := NewBlockFactory(t)
	blocks := factory.GetBlocks(3, 2)
	wal := NewMockedTestBlockWAL()
	blockCache, err := NewBlockCache(NewDefaultTimeProvider(), 100, wal, log)
	require.NoError(t, err)
	blockCache.AddBlock(blocks[0])
	blockCache.AddBlock(blocks[1])
	require.NotNil(t, blockCache.GetBlock(blocks[0].L1Commitment()))
	require.NotNil(t, blockCache.GetBlock(blocks[1].L1Commitment()))
	require.Nil(t, blockCache.GetBlock(blocks[2].L1Commitment()))
	blockCache.CleanOlderThan(time.Now())
	require.True(t, wal.Contains(blocks[0].Hash()))
	require.True(t, wal.Contains(blocks[1].Hash()))
	require.False(t, wal.Contains(blocks[2].Hash()))
	require.NotNil(t, blockCache.GetBlock(blocks[0].L1Commitment()))
	require.NotNil(t, blockCache.GetBlock(blocks[1].L1Commitment()))
	require.Nil(t, blockCache.GetBlock(blocks[2].L1Commitment()))
}

// Test if blocks are cleaned from cache on timer as well as on cache size.
func TestBlockCacheCleanUp(t *testing.T) {
	log := testlogger.NewLogger(t)
	defer log.Sync()

	factory := NewBlockFactory(t)
	blocks := factory.GetBlocks(16, 2)
	now := time.Now()
	wal := NewEmptyTestBlockWAL()
	tp := NewArtifficialTimeProvider(now)
	blockCache, err := NewBlockCache(tp, 5, wal, log)
	require.NoError(t, err)
	setNowAddBlockFun := func(d time.Duration, b state.Block) {
		tp.SetNow(now.Add(d))
		blockCache.AddBlock(b)
	}
	blockCache.AddBlock(blocks[0])
	setNowAddBlockFun(1*time.Second, blocks[1])
	setNowAddBlockFun(2*time.Second, blocks[2])
	setNowAddBlockFun(4*time.Second, blocks[3])
	setNowAddBlockFun(5*time.Second, blocks[4])
	require.Equal(t, 5, blockCache.Size())
	require.NotNil(t, blockCache.GetBlock(blocks[0].L1Commitment()))
	require.NotNil(t, blockCache.GetBlock(blocks[1].L1Commitment()))
	require.NotNil(t, blockCache.GetBlock(blocks[2].L1Commitment()))
	require.NotNil(t, blockCache.GetBlock(blocks[3].L1Commitment()))
	require.NotNil(t, blockCache.GetBlock(blocks[4].L1Commitment()))
	setNowAddBlockFun(6*time.Second, blocks[5]) // cache overflows
	require.Equal(t, 5, blockCache.Size())
	require.Nil(t, blockCache.GetBlock(blocks[0].L1Commitment()))
	require.NotNil(t, blockCache.GetBlock(blocks[1].L1Commitment()))
	require.NotNil(t, blockCache.GetBlock(blocks[2].L1Commitment()))
	require.NotNil(t, blockCache.GetBlock(blocks[3].L1Commitment()))
	require.NotNil(t, blockCache.GetBlock(blocks[4].L1Commitment()))
	require.NotNil(t, blockCache.GetBlock(blocks[5].L1Commitment()))
	blockCache.CleanOlderThan(now.Add(3 * time.Second))
	require.Equal(t, 3, blockCache.Size())
	require.Nil(t, blockCache.GetBlock(blocks[1].L1Commitment()))
	require.Nil(t, blockCache.GetBlock(blocks[2].L1Commitment()))
	require.NotNil(t, blockCache.GetBlock(blocks[3].L1Commitment()))
	require.NotNil(t, blockCache.GetBlock(blocks[4].L1Commitment()))
	require.NotNil(t, blockCache.GetBlock(blocks[5].L1Commitment()))
	setNowAddBlockFun(7*time.Second, blocks[6])
	setNowAddBlockFun(8*time.Second, blocks[7])
	setNowAddBlockFun(9*time.Second, blocks[8]) // cache overflows
	setNowAddBlockFun(10*time.Second, blocks[9])
	require.Equal(t, 5, blockCache.Size())
	require.Nil(t, blockCache.GetBlock(blocks[3].L1Commitment()))
	require.Nil(t, blockCache.GetBlock(blocks[4].L1Commitment()))
	require.NotNil(t, blockCache.GetBlock(blocks[5].L1Commitment()))
	require.NotNil(t, blockCache.GetBlock(blocks[6].L1Commitment()))
	require.NotNil(t, blockCache.GetBlock(blocks[7].L1Commitment()))
	require.NotNil(t, blockCache.GetBlock(blocks[8].L1Commitment()))
	require.NotNil(t, blockCache.GetBlock(blocks[9].L1Commitment()))
	setNowAddBlockFun(11*time.Second, blocks[10])
	setNowAddBlockFun(12*time.Second, blocks[11])
	setNowAddBlockFun(13*time.Second, blocks[12])
	setNowAddBlockFun(14*time.Second, blocks[13])
	setNowAddBlockFun(15*time.Second, blocks[14])
	setNowAddBlockFun(16*time.Second, blocks[15])
	require.Equal(t, 5, blockCache.Size())
	require.Nil(t, blockCache.GetBlock(blocks[5].L1Commitment()))
	require.Nil(t, blockCache.GetBlock(blocks[6].L1Commitment()))
	require.Nil(t, blockCache.GetBlock(blocks[7].L1Commitment()))
	require.Nil(t, blockCache.GetBlock(blocks[8].L1Commitment()))
	require.Nil(t, blockCache.GetBlock(blocks[9].L1Commitment()))
	require.Nil(t, blockCache.GetBlock(blocks[10].L1Commitment()))
	require.NotNil(t, blockCache.GetBlock(blocks[11].L1Commitment()))
	require.NotNil(t, blockCache.GetBlock(blocks[12].L1Commitment()))
	require.NotNil(t, blockCache.GetBlock(blocks[13].L1Commitment()))
	require.NotNil(t, blockCache.GetBlock(blocks[14].L1Commitment()))
	require.NotNil(t, blockCache.GetBlock(blocks[15].L1Commitment()))
}

// Test if blocks are put into cache AND WAL and taken from cache OR WAL.
// NOTE: the situation, when a block is available in cache but not in WAL is
// unlikely. It may happen only if WAL gets corrupted outside Wasp.
func TestBlockCacheFull(t *testing.T) {
	log := testlogger.NewLogger(t)
	defer log.Sync()

	factory := NewBlockFactory(t)
	blocks := factory.GetBlocks(5, 2)
	now := time.Now()
	wal := NewMockedTestBlockWAL()
	tp := NewArtifficialTimeProvider(now)
	blockCache, err := NewBlockCache(tp, 100, wal, log)
	require.NoError(t, err)
	blockCache.AddBlock(blocks[0]) // Will be dropped from cache AND WAL
	blockCache.AddBlock(blocks[1]) // Will be dropped from cache
	tp.SetNow(now.Add(2 * time.Second))
	blockCache.AddBlock(blocks[2]) // Will be dropped from WAL
	blockCache.AddBlock(blocks[3]) // Will remain in both cache AND WAL
	require.NotNil(t, blockCache.GetBlock(blocks[0].L1Commitment()))
	require.NotNil(t, blockCache.GetBlock(blocks[1].L1Commitment()))
	require.NotNil(t, blockCache.GetBlock(blocks[2].L1Commitment()))
	require.NotNil(t, blockCache.GetBlock(blocks[3].L1Commitment()))
	require.Nil(t, blockCache.GetBlock(blocks[4].L1Commitment()))
	blockCache.CleanOlderThan(now.Add(1 * time.Second)) // Removing blocks 0 and 1
	wal.Delete(blocks[0].L1Commitment().BlockHash())
	wal.Delete(blocks[2].L1Commitment().BlockHash())
	require.False(t, wal.Contains(blocks[0].Hash()))
	require.True(t, wal.Contains(blocks[1].Hash()))
	require.False(t, wal.Contains(blocks[2].Hash()))
	require.True(t, wal.Contains(blocks[3].Hash()))
	require.Nil(t, blockCache.GetBlock(blocks[0].L1Commitment()))
	require.NotNil(t, blockCache.GetBlock(blocks[1].L1Commitment()))
	require.NotNil(t, blockCache.GetBlock(blocks[2].L1Commitment()))
	require.NotNil(t, blockCache.GetBlock(blocks[3].L1Commitment()))
	require.Nil(t, blockCache.GetBlock(blocks[4].L1Commitment()))
}
