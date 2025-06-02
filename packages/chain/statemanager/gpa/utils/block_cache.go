// Package utils provides utility functions and structures for the state manager's GPA implementation.
package utils

import (
	"time"

	"github.com/samber/lo"

	"github.com/iotaledger/hive.go/ds/shrinkingmap"
	"github.com/iotaledger/hive.go/log"

	"github.com/iotaledger/wasp/packages/metrics"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util/timeutil"
)

type blockTime struct {
	time     time.Time
	blockKey BlockKey
}

type blockCache struct {
	log          log.Logger
	blocks       *shrinkingmap.ShrinkingMap[BlockKey, state.Block]
	maxCacheSize int
	wal          BlockWAL
	times        []*blockTime
	timeProvider timeutil.TimeProvider
	metrics      *metrics.ChainStateManagerMetrics
}

var _ BlockCache = &blockCache{}

func NewBlockCache(tp timeutil.TimeProvider, maxCacheSize int, wal BlockWAL, metrics *metrics.ChainStateManagerMetrics, log log.Logger) (BlockCache, error) {
	return &blockCache{
		log:          log.NewChildLogger("BC"),
		blocks:       shrinkingmap.New[BlockKey, state.Block](),
		maxCacheSize: maxCacheSize,
		wal:          wal,
		times:        make([]*blockTime, 0),
		timeProvider: tp,
		metrics:      metrics,
	}, nil
}

// Adds block to cache and WAL
func (bcT *blockCache) AddBlock(block state.Block) {
	err := bcT.wal.Write(block)
	if err != nil {
		bcT.log.LogErrorf("Failed writing block index %v %s to WAL: %v", block.StateIndex(), block.L1Commitment(), err)
	}
	bcT.addBlockToCache(block)
}

// Adds block to cache only
func (bcT *blockCache) addBlockToCache(block state.Block) {
	commitment := block.L1Commitment()
	blockKey := NewBlockKey(commitment)
	_, exists := bcT.blocks.Get(blockKey)
	if exists {
		bcT.times = lo.Filter(bcT.times, func(bt *blockTime, _ int) bool {
			return !bt.blockKey.Equals(blockKey)
		})
	}
	bcT.blocks.Set(blockKey, block)
	bcT.times = append(bcT.times, &blockTime{
		time:     bcT.timeProvider.GetNow(),
		blockKey: blockKey,
	})
	bcT.log.LogDebugf("Block index %v %s added to cache", block.StateIndex(), commitment)

	if bcT.Size() > bcT.maxCacheSize {
		blockKey := bcT.times[0].blockKey
		bcT.times[0] = nil // Freeing up memory
		bcT.times = bcT.times[1:]
		bcT.blocks.Delete(blockKey)
		bcT.log.LogDebugf("Block %s deleted from cache, because cache is too large", blockKey)
	}
	bcT.metrics.SetCacheSize(bcT.Size())
}

func (bcT *blockCache) GetBlock(commitment *state.L1Commitment) state.Block {
	blockKey := NewBlockKey(commitment)
	// Check in cache
	block, exists := bcT.blocks.Get(blockKey)
	if exists {
		bcT.log.LogDebugf("Block index %v %s retrieved from cache", block.StateIndex(), commitment)
		return block
	}

	// Check in WAL
	if bcT.wal.Contains(commitment.BlockHash()) {
		block, err := bcT.wal.Read(commitment.BlockHash())
		if err != nil {
			bcT.log.LogErrorf("Error reading block index %v %s from WAL: %w", block.StateIndex(), commitment, err)
			return nil
		}
		bcT.addBlockToCache(block)
		bcT.log.LogDebugf("Block index %v %s retrieved from WAL", block.StateIndex(), commitment)
		return block
	}

	return nil
}

func (bcT *blockCache) CleanOlderThan(limit time.Time) {
	defer bcT.metrics.SetCacheSize(bcT.Size())
	for i, bt := range bcT.times {
		if bt.time.After(limit) {
			bcT.times = bcT.times[i:]
			return
		}
		bcT.blocks.Delete(bt.blockKey)
		bcT.times[i] = nil // Freeing up memory
		bcT.log.LogDebugf("Block %s deleted from cache, because it is too old", bt.blockKey)
	}
	bcT.times = make([]*blockTime, 0) // All the blocks were deleted
}

func (bcT *blockCache) Size() int {
	return bcT.blocks.Size()
}
