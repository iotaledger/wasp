package smGPAUtils

import (
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/maps"
	"pgregory.net/rapid"

	"github.com/iotaledger/wasp/packages/state"
)

type blockCacheTestSM struct { // State machine for block cache property based Rapid tests
	*blockCacheNoWALTestSM
	blocksNotInWAL []BlockKey
	wal            TestBlockWAL
}

var _ rapid.StateMachine = &blockCacheTestSM{}

func (bctsmT *blockCacheTestSM) Init(t *rapid.T) {
	bctsmT.blockCacheNoWALTestSM = &blockCacheNoWALTestSM{}
	bctsmT.blocksNotInWAL = []BlockKey{}
	bctsmT.wal = NewMockedTestBlockWAL()
	bctsmT.blockCacheNoWALTestSM.initStateMachine(t, 10, bctsmT.wal, func(block state.Block) {
		bctsmT.blocksNotInWAL = lo.Without(bctsmT.blocksNotInWAL, NewBlockKey(block.L1Commitment()))
	})
}

// Cleanup() // inherited from blockCacheNoWALTestSM

func (bctsmT *blockCacheTestSM) Check(t *rapid.T) {
	bctsmT.blockCacheNoWALTestSM.Check(t)
	bctsmT.invariantAllBlocksInWAL(t)
}

// AddNewBlock(t *rapid.T) // inherited from blockCacheNoWALTestSM
// AddExistingBlock(t *rapid.T) // inherited from blockCacheNoWALTestSM
// CleanCache(t *rapid.T) // inherited from blockCacheNoWALTestSM

// Highly unlikely
func (bctsmT *blockCacheTestSM) RemoveBlockFromWAL(t *rapid.T) {
	if len(bctsmT.blocksNotInWAL) == len(bctsmT.blocks) {
		t.Skip()
	}
	blocksToChoose := lo.Without(maps.Keys(bctsmT.blocks), bctsmT.blocksNotInWAL...)
	blockKey := rapid.SampledFrom(blocksToChoose).Example()
	bctsmT.wal.Delete(blockKey.AsBlockHash())
	bctsmT.blocksNotInWAL = append(bctsmT.blocksNotInWAL, blockKey)
	t.Logf("Block %s is removed from WAL", blockKey)
}

func (bctsmT *blockCacheTestSM) GetBlockFromCache(t *rapid.T) {
	blocksToChoose := lo.Intersect(bctsmT.blocksNotInWAL, bctsmT.blocksInCache)
	if len(blocksToChoose) == 0 {
		t.Skip()
	}
	blockKey := rapid.SampledFrom(blocksToChoose).Example()
	bctsmT.getAndCheckBlock(t, blockKey)
	t.Logf("Block %s is retrieved from cache", blockKey)
}

func (bctsmT *blockCacheTestSM) GetBlockFromWAL(t *rapid.T) {
	blocksToChoose := bctsmT.blocksNotInCache(t)
	blocksToChoose = lo.Without(blocksToChoose, bctsmT.blocksNotInWAL...)
	if len(blocksToChoose) == 0 {
		t.Skip()
	}
	blockKey := rapid.SampledFrom(blocksToChoose).Example()
	bctsmT.getAndCheckBlock(t, blockKey)
	t.Logf("Block %s is retrieved from WAL", blockKey)
}

func (bctsmT *blockCacheTestSM) GetBlockFromCacheOrWAL(t *rapid.T) {
	blocksToChoose := lo.Without(append([]BlockKey{}, bctsmT.blocksInCache...), bctsmT.blocksNotInWAL...)
	if len(blocksToChoose) == 0 {
		t.Skip()
	}
	blockKey := rapid.SampledFrom(blocksToChoose).Example()
	bctsmT.getAndCheckBlock(t, blockKey)
	t.Logf("Block %s is retrieved from cache or wal", blockKey)
}

func (bctsmT *blockCacheTestSM) GetBlockFromNowhere(t *rapid.T) { // Unsuccessfully
	blocksToChoose := lo.Intersect(bctsmT.blocksNotInWAL, bctsmT.blocksNotInCache(t))
	if len(blocksToChoose) == 0 {
		t.Skip()
	}
	blockKey := rapid.SampledFrom(blocksToChoose).Example()
	blockExpected, ok := bctsmT.blocks[blockKey]
	require.True(t, ok)
	commitment := blockExpected.L1Commitment()
	block := bctsmT.bc.GetBlock(commitment)
	require.Nil(t, block)
	t.Logf("Block %s is not retrieved, because it is neither in cache nor in WAL", commitment)
}

// Restart(t *rapid.T) // inherited from blockCacheNoWALTestSM

func (bctsmT *blockCacheTestSM) invariantAllBlocksInWAL(t *rapid.T) {
	for _, blockKey := range lo.Without(maps.Keys(bctsmT.blocks), bctsmT.blocksNotInWAL...) {
		require.True(t, bctsmT.wal.Contains(blockKey.AsBlockHash()))
	}
}

func TestBlockCachePropBasedFull(t *testing.T) {
	rapid.Check(t, rapid.Run[*blockCacheTestSM]())
}
