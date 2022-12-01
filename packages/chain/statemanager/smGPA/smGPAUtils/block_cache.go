package smGPAUtils

import (
	"time"

	"github.com/iotaledger/hive.go/core/kvstore"
	"github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/wasp/packages/state"
)

type blockTime struct {
	time      time.Time
	blockHash state.BlockHash
}

type blockCache struct {
	log          *logger.Logger
	store        kvstore.KVStore
	blocks       map[state.BlockHash]state.Block
	wal          BlockWAL
	times        []*blockTime
	timeProvider TimeProvider
}

var _ BlockCache = &blockCache{}

func NewBlockCache(store kvstore.KVStore, tp TimeProvider, wal BlockWAL, log *logger.Logger) (BlockCache, error) {
	return &blockCache{
		log:          log.Named("bc"),
		store:        store,
		blocks:       make(map[state.BlockHash]state.Block),
		wal:          wal,
		times:        make([]*blockTime, 0),
		timeProvider: tp,
	}, nil
}

func (bcT *blockCache) AddBlock(block state.Block) error {
	blockHash := block.GetHash()
	err := bcT.wal.Write(block)
	if err != nil {
		bcT.log.Debugf("Failed writing block %s to WAL: %v", blockHash, err)
		return err
	}
	bcT.log.Debugf("Block %s written to WAL", blockHash)
	bcT.addBlockToCache(blockHash, block)
	return nil
}

func (bcT *blockCache) addBlockToCache(blockHash state.BlockHash, block state.Block) {
	bcT.blocks[blockHash] = block
	bcT.times = append(bcT.times, &blockTime{
		time:      bcT.timeProvider.GetNow(),
		blockHash: blockHash,
	})
	bcT.log.Debugf("Block %v added to cache", blockHash)
}

func (bcT *blockCache) GetBlock(blockIndex uint32, blockHash state.BlockHash) state.Block {
	// Check in cache
	block, ok := bcT.blocks[blockHash]

	if ok {
		return block
	}
	bcT.log.Debugf("Block %s is not in cache", blockHash)

	// Check in WAL
	if bcT.wal.Contains(blockHash) {
		block, err := bcT.wal.Read(blockHash)
		if err != nil {
			bcT.log.Errorf("Error reading block %s from WAL: %v", blockHash, err)
			return nil
		}
		bcT.addBlockToCache(blockHash, block)
		return block
	}
	bcT.log.Debugf("Block %s is not in WAL", blockHash)

	// Check in DB. TODO: DB should also search by hash instead of block index
	blockBytes, err := state.LoadBlockBytes(bcT.store, blockIndex)
	if err != nil {
		bcT.log.Errorf("Loading block index #%d (%s) from the DB failed: %v", blockIndex, blockHash, err)
		return nil
	}
	if blockBytes == nil {
		bcT.log.Debugf("Block index #%d (%s) not found in the DB", blockIndex, blockHash)
		return nil
	}
	block, err = state.BlockFromBytes(blockBytes)
	if err != nil {
		bcT.log.Errorf("Failed unmarshalling block index #%d (%s), which was loaded from the DB: %v",
			blockIndex, blockHash, err)
		return nil
	}
	if !blockHash.Equals(block.GetHash()) {
		bcT.log.Errorf("Block index #%d loaded from the database does not have the expected hash %s",
			blockIndex, blockHash)
		return nil
	}

	bcT.addBlockToCache(blockHash, block)
	return block
}

func (bcT *blockCache) CleanOlderThan(limit time.Time) {
	for i, bt := range bcT.times {
		if bt.time.After(limit) {
			bcT.times = bcT.times[i:]
			return
		}
		delete(bcT.blocks, bt.blockHash)
		bcT.log.Debugf("Block %v deleted from cache", bt.blockHash)
	}
	bcT.times = make([]*blockTime, 0) // All the blocks were deleted
}
