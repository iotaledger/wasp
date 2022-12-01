// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package smGPAUtils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/core/kvstore/mapdb"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
)

func TestBlockCacheSimple(t *testing.T) {
	log := testlogger.NewLogger(t)
	defer log.Sync()

	_, blocks, _ := GetBlocks(t, 4, 1)
	blockCache, err := NewBlockCache(mapdb.NewMapDB(), NewDefaultTimeProvider(), NewEmptyBlockWAL(), log)
	require.NoError(t, err)
	blockCache.AddBlock(blocks[0])
	blockCache.AddBlock(blocks[1])
	blockCache.AddBlock(blocks[2])
	require.NotNil(t, blockCache.GetBlock(1, blocks[0].GetHash()))
	require.NotNil(t, blockCache.GetBlock(3, blocks[2].GetHash()))
	require.NotNil(t, blockCache.GetBlock(2, blocks[1].GetHash()))
	require.Nil(t, blockCache.GetBlock(4, blocks[3].GetHash()))
}

func TestBlockCacheCleaning(t *testing.T) {
	log := testlogger.NewLogger(t)
	defer log.Sync()

	_, blocks, _ := GetBlocks(t, 6, 2)
	blockCache, err := NewBlockCache(mapdb.NewMapDB(), NewDefaultTimeProvider(), NewEmptyBlockWAL(), log)
	require.NoError(t, err)
	beforeTime := time.Now()
	blockCache.AddBlock(blocks[0])
	blockCache.AddBlock(blocks[1])
	require.NotNil(t, blockCache.GetBlock(1, blocks[0].GetHash()))
	require.NotNil(t, blockCache.GetBlock(2, blocks[1].GetHash()))
	blockCache.CleanOlderThan(beforeTime)
	require.NotNil(t, blockCache.GetBlock(1, blocks[0].GetHash()))
	require.NotNil(t, blockCache.GetBlock(2, blocks[1].GetHash()))
	blockCache.CleanOlderThan(time.Now())
	require.Nil(t, blockCache.GetBlock(1, blocks[0].GetHash()))
	require.Nil(t, blockCache.GetBlock(2, blocks[1].GetHash()))
	blockCache.AddBlock(blocks[2])
	blockCache.AddBlock(blocks[3])
	inTheMiddleTime := time.Now()
	blockCache.AddBlock(blocks[4])
	blockCache.AddBlock(blocks[5])
	require.NotNil(t, blockCache.GetBlock(2, blocks[2].GetHash()))
	require.NotNil(t, blockCache.GetBlock(3, blocks[3].GetHash()))
	require.NotNil(t, blockCache.GetBlock(3, blocks[4].GetHash()))
	require.NotNil(t, blockCache.GetBlock(3, blocks[5].GetHash()))
	blockCache.CleanOlderThan(inTheMiddleTime)
	require.Nil(t, blockCache.GetBlock(2, blocks[2].GetHash()))
	require.Nil(t, blockCache.GetBlock(3, blocks[3].GetHash()))
	require.NotNil(t, blockCache.GetBlock(3, blocks[4].GetHash()))
	require.NotNil(t, blockCache.GetBlock(3, blocks[5].GetHash()))
}

// Test if blocks are put/taken from WAL, if they are not available in cache
func TestBlockCacheWAL(t *testing.T) {
	log := testlogger.NewLogger(t)
	defer log.Sync()

	_, blocks, _ := GetBlocks(t, 3, 2)
	blockCache, err := NewBlockCache(mapdb.NewMapDB(), NewDefaultTimeProvider(), NewMockedBlockWAL(), log)
	require.NoError(t, err)
	blockCache.AddBlock(blocks[0])
	blockCache.AddBlock(blocks[1])
	require.NotNil(t, blockCache.GetBlock(1, blocks[0].GetHash()))
	require.NotNil(t, blockCache.GetBlock(2, blocks[1].GetHash()))
	require.Nil(t, blockCache.GetBlock(2, blocks[2].GetHash()))
	blockCache.CleanOlderThan(time.Now()) // Blocks are cleaned from cache, but should be accessible from WAL
	require.NotNil(t, blockCache.GetBlock(1, blocks[0].GetHash()))
	require.NotNil(t, blockCache.GetBlock(2, blocks[1].GetHash()))
	require.Nil(t, blockCache.GetBlock(2, blocks[2].GetHash()))
}
