//nolint:unused // false positives because of rapid.Check
package smGPAUtils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/exp/maps"
	"pgregory.net/rapid"

	"github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/packages/util"
)

type blockCacheNoWALTestSM struct { // State machine for block cache no WAL property based Rapid tests
	bc                  BlockCache
	factory             *BlockFactory
	lastBlockCommitment *state.L1Commitment
	blocks              map[BlockKey]state.Block
	blockTimes          []*blockTime
	blocksInCache       []BlockKey
	log                 *logger.Logger
}

func (bcnwtsmT *blockCacheNoWALTestSM) initWAL(t *rapid.T, wal BlockWAL) {
	var err error
	bcnwtsmT.factory = NewBlockFactory(t)
	bcnwtsmT.lastBlockCommitment = state.OriginL1Commitment()
	bcnwtsmT.log = testlogger.NewLogger(t)
	bcnwtsmT.bc, err = NewBlockCache(NewDefaultTimeProvider(), wal, bcnwtsmT.log)
	require.NoError(t, err)
	bcnwtsmT.blockTimes = make([]*blockTime, 0)
	bcnwtsmT.blocks = make(map[BlockKey]state.Block)
	bcnwtsmT.blocksInCache = make([]BlockKey, 0)
}

func (bcnwtsmT *blockCacheNoWALTestSM) Init(t *rapid.T) {
	bcnwtsmT.initWAL(t, NewEmptyBlockWAL())
}

func (bcnwtsmT *blockCacheNoWALTestSM) Cleanup() {
	bcnwtsmT.log.Sync()
}

func (bcnwtsmT *blockCacheNoWALTestSM) Check(t *rapid.T) {
	bcnwtsmT.invariantAllBlocksInCacheDifferent(t)
	bcnwtsmT.invariantBlocksInCacheBijectionToBlockTimes(t)
}

func (bcnwtsmT *blockCacheNoWALTestSM) AddNewBlock(t *rapid.T) {
	block := bcnwtsmT.factory.GetNextBlock(bcnwtsmT.lastBlockCommitment)
	bcnwtsmT.lastBlockCommitment = block.L1Commitment()
	bcnwtsmT.addBlock(t, block)
	t.Logf("New block %s added to cache", bcnwtsmT.lastBlockCommitment)
}

func (bcnwtsmT *blockCacheNoWALTestSM) AddExistingBlock(t *rapid.T) {
	if len(bcnwtsmT.blocksInCache) == len(bcnwtsmT.blocks) {
		t.Skip()
	}
	blockKey := rapid.SampledFrom(bcnwtsmT.blocksNotInCache(t)).Example()
	block, ok := bcnwtsmT.blocks[blockKey]
	require.True(t, ok)
	bcnwtsmT.addBlock(t, block)
	t.Logf("Block %s added to cache again", block.L1Commitment())
}

func (bcnwtsmT *blockCacheNoWALTestSM) CleanCache(t *rapid.T) {
	if len(bcnwtsmT.blockTimes) == 0 {
		t.Skip()
	}
	index := rapid.Uint32Range(0, uint32(len(bcnwtsmT.blockTimes)-1)).Example()
	time := bcnwtsmT.blockTimes[index].time
	bcnwtsmT.bc.CleanOlderThan(time)
	for i := uint32(0); i <= index; i++ {
		blockKey := bcnwtsmT.blockTimes[i].blockKey
		bcnwtsmT.blocksInCache = util.Remove(blockKey, bcnwtsmT.blocksInCache)
		t.Logf("Block %s deleted from cache", blockKey)
	}
	bcnwtsmT.blockTimes = bcnwtsmT.blockTimes[index+1:]
	t.Logf("Cache cleaned until %v", time)
}

func (bcnwtsmT *blockCacheNoWALTestSM) GetBlockFromCache(t *rapid.T) {
	if len(bcnwtsmT.blocksInCache) == 0 {
		t.Skip()
	}
	blockKey := rapid.SampledFrom(bcnwtsmT.blocksInCache).Example()
	bcnwtsmT.tstGetBlockFromCache(t, blockKey)
}

func (bcnwtsmT *blockCacheNoWALTestSM) tstGetBlockFromCache(t *rapid.T, blockKey BlockKey) {
	bcnwtsmT.getAndCheckBlock(t, blockKey)
	t.Logf("Block from cache %s obtained", blockKey)
}

func (bcnwtsmT *blockCacheNoWALTestSM) Restart(t *rapid.T) {
	var err error
	bcnwtsmT.bc, err = NewBlockCache(NewDefaultTimeProvider(), bcnwtsmT.bc.(*blockCache).wal, bcnwtsmT.log)
	require.NoError(t, err)
	bcnwtsmT.blocksInCache = make([]BlockKey, 0)
	bcnwtsmT.blockTimes = make([]*blockTime, 0)
	t.Log("Block cache was restarted")
}

func (bcnwtsmT *blockCacheNoWALTestSM) invariantAllBlocksInCacheDifferent(t *rapid.T) {
	require.True(t, util.AllDifferent(bcnwtsmT.blocksInCache))
}

func (bcnwtsmT *blockCacheNoWALTestSM) invariantBlocksInCacheBijectionToBlockTimes(t *rapid.T) {
	blockTimeKeys := make([]BlockKey, len(bcnwtsmT.blockTimes))
	for i := range bcnwtsmT.blockTimes {
		blockTimeKeys[i] = bcnwtsmT.blockTimes[i].blockKey
	}
	require.Equal(t, len(bcnwtsmT.blocksInCache), len(blockTimeKeys))
	for i := range bcnwtsmT.blocksInCache {
		require.True(t, util.Contains(bcnwtsmT.blocksInCache[i], blockTimeKeys))
	}
}

func (bcnwtsmT *blockCacheNoWALTestSM) addBlock(t *rapid.T, block state.Block) {
	blockKey := NewBlockKey(block.L1Commitment())
	bcnwtsmT.blocks[blockKey] = block
	bcnwtsmT.bc.AddBlock(block)
	require.False(t, util.Contains(blockKey, bcnwtsmT.blocksInCache))
	bcnwtsmT.blocksInCache = append(bcnwtsmT.blocksInCache, blockKey)
	bcnwtsmT.blockTimes = append(bcnwtsmT.blockTimes, &blockTime{
		time:     time.Now(),
		blockKey: blockKey,
	})
}

func (bcnwtsmT *blockCacheNoWALTestSM) blocksNotInCache(t *rapid.T) []BlockKey {
	return util.RemoveAll(bcnwtsmT.blocksInCache, maps.Keys(bcnwtsmT.blocks))
}

func (bcnwtsmT *blockCacheNoWALTestSM) getAndCheckBlock(t *rapid.T, blockKey BlockKey) {
	blockExpected, ok := bcnwtsmT.blocks[blockKey]
	require.True(t, ok)
	block := bcnwtsmT.bc.GetBlock(blockExpected.L1Commitment())
	require.NotNil(t, block)
	require.True(t, blockExpected.Hash().Equals(block.Hash())) // Should be Equals instead of Hash().Equals(); bwtsmT.blocks[blockHash]
}

func TestBlockCacheNoWALPropBased(t *testing.T) {
	rapid.Check(t, rapid.Run[*blockCacheNoWALTestSM]())
}
