// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package smGPAUtils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/testutil/testlogger"
)

func TestBlockCacheSimple(t *testing.T) {
	log := testlogger.NewLogger(t)
	defer log.Sync()

	factory := NewBlockFactory(t)
	blocks := factory.GetBlocks(4, 1)
	blockCache, err := NewBlockCache(NewDefaultTimeProvider(), NewEmptyTestBlockWAL(), log)
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
	blockCache, err := NewBlockCache(NewDefaultTimeProvider(), NewEmptyTestBlockWAL(), log)
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
	blockCache, err := NewBlockCache(NewDefaultTimeProvider(), wal, log)
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
	blockCache, err := NewBlockCache(tp, wal, log)
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
