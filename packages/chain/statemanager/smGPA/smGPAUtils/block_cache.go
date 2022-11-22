package smGPAUtils

import (
	"time"

	"github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/wasp/packages/state"
)

type blockTime struct {
	time     time.Time
	blockKey BlockKey
}

type blockCache struct {
	log          *logger.Logger
	blocks       map[BlockKey]state.Block
	wal          BlockWAL
	times        []*blockTime
	timeProvider TimeProvider
}

var _ BlockCache = &blockCache{}

func NewBlockCache(tp TimeProvider, wal BlockWAL, log *logger.Logger) (BlockCache, error) {
	return &blockCache{
		log:          log.Named("bc"),
		blocks:       make(map[BlockKey]state.Block),
		wal:          wal,
		times:        make([]*blockTime, 0),
		timeProvider: tp,
	}, nil
}

// Adds block to cache and WAL
func (bcT *blockCache) AddBlock(block state.Block) error {
	blockKey := NewBlockKey(block.L1Commitment())
	err := bcT.wal.Write(block)
	if err != nil {
		bcT.log.Errorf("Failed writing block %s to WAL: %v", blockKey, err)
		return err
	}
	bcT.log.Debugf("Block %s written to WAL", blockKey)

	bcT.blocks[blockKey] = block
	bcT.times = append(bcT.times, &blockTime{
		time:     bcT.timeProvider.GetNow(),
		blockKey: blockKey,
	})
	bcT.log.Debugf("Block %s added to cache", blockKey)
	return nil
}

func (bcT *blockCache) GetBlock(commitment *state.L1Commitment) state.Block {
	blockKey := NewBlockKey(commitment)
	// Check in cache
	block, ok := bcT.blocks[blockKey]

	if ok {
		bcT.log.Debugf("Block %s retrieved from cache", commitment)
		return block
	}
	bcT.log.Debugf("Block %s is not in cache", commitment)
	return nil
}

func (bcT *blockCache) CleanOlderThan(limit time.Time) {
	for i, bt := range bcT.times {
		if bt.time.After(limit) {
			bcT.times = bcT.times[i:]
			return
		}
		delete(bcT.blocks, bt.blockKey)
		bcT.log.Debugf("Block %s deleted from cache", bt.blockKey)
	}
	bcT.times = make([]*blockTime, 0) // All the blocks were deleted
}
