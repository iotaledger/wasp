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
	blocks, _ := factory.GetBlocks(4, 1)
	blockCache, err := NewBlockCache(NewDefaultTimeProvider(), NewEmptyBlockWAL(), log)
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
	blocks, _ := factory.GetBlocks(6, 2)
	blockCache, err := NewBlockCache(NewDefaultTimeProvider(), NewEmptyBlockWAL(), log)
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
	blocks, _ := factory.GetBlocks(3, 2)
	wal := NewMockedBlockWAL()
	blockCache, err := NewBlockCache(NewDefaultTimeProvider(), wal, log)
	require.NoError(t, err)
	blockCache.AddBlock(blocks[0])
	blockCache.AddBlock(blocks[1])
	require.NotNil(t, blockCache.GetBlock(blocks[0].L1Commitment()))
	require.NotNil(t, blockCache.GetBlock(blocks[1].L1Commitment()))
	require.Nil(t, blockCache.GetBlock(blocks[2].L1Commitment()))
	blockCache.CleanOlderThan(time.Now()) // Blocks are cleaned from cache, are accessible from WAL, but block cache does not retrieve them from there
	require.True(t, wal.Contains(blocks[0].Hash()))
	require.True(t, wal.Contains(blocks[1].Hash()))
	require.False(t, wal.Contains(blocks[2].Hash()))
	require.Nil(t, blockCache.GetBlock(blocks[0].L1Commitment()))
	require.Nil(t, blockCache.GetBlock(blocks[1].L1Commitment()))
	require.Nil(t, blockCache.GetBlock(blocks[2].L1Commitment()))
}
