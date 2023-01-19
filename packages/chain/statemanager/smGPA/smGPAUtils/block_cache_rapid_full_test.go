//nolint:unused // false positives because of rapid.Check
package smGPAUtils

import (
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/exp/maps"
	"pgregory.net/rapid"

	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util"
)

type blockCacheTestSM struct { // State machine for block cache property based Rapid tests
	*blockCacheNoWALTestSM
	blocksNotInWAL []BlockKey
	wal            TestBlockWAL
}

func (bctsmT *blockCacheTestSM) Init(t *rapid.T) {
	bctsmT.blockCacheNoWALTestSM = &blockCacheNoWALTestSM{}
	bctsmT.blocksNotInWAL = []BlockKey{}
	bctsmT.wal = NewMockedTestBlockWAL()
	bctsmT.blockCacheNoWALTestSM.initStateMachine(t, bctsmT.wal, func(block state.Block) {
		bctsmT.blocksNotInWAL = util.Remove(NewBlockKey(block.L1Commitment()), bctsmT.blocksNotInWAL)
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
	blocksToChoose := util.RemoveAll(bctsmT.blocksNotInWAL, maps.Keys(bctsmT.blocks))
	blockKey := rapid.SampledFrom(blocksToChoose).Example()
	bctsmT.wal.Delete(blockKey.AsBlockHash())
	bctsmT.blocksNotInWAL = append(bctsmT.blocksNotInWAL, blockKey)
	t.Logf("Block %s is removed from WAL", blockKey)
}

func (bctsmT *blockCacheTestSM) GetBlockFromCache(t *rapid.T) {
	blocksToChoose := util.Intersection(bctsmT.blocksNotInWAL, bctsmT.blocksInCache)
	if len(blocksToChoose) == 0 {
		t.Skip()
	}
	blockKey := rapid.SampledFrom(blocksToChoose).Example()
	bctsmT.getAndCheckBlock(t, blockKey)
	t.Logf("Block %s is retrieved from cache", blockKey)
}

func (bctsmT *blockCacheTestSM) GetBlockFromWAL(t *rapid.T) {
	blocksToChoose := bctsmT.blocksNotInCache(t)
	blocksToChoose = util.RemoveAll(bctsmT.blocksNotInWAL, blocksToChoose)
	if len(blocksToChoose) == 0 {
		t.Skip()
	}
	blockKey := rapid.SampledFrom(blocksToChoose).Example()
	bctsmT.getAndCheckBlock(t, blockKey)
	t.Logf("Block %s is retrieved from WAL", blockKey)
}

func (bctsmT *blockCacheTestSM) GetBlockFromCacheOrWAL(t *rapid.T) {
	blocksToChoose := util.RemoveAll(bctsmT.blocksNotInWAL, append([]BlockKey{}, bctsmT.blocksInCache...))
	if len(blocksToChoose) == 0 {
		t.Skip()
	}
	blockKey := rapid.SampledFrom(blocksToChoose).Example()
	bctsmT.getAndCheckBlock(t, blockKey)
	t.Logf("Block %s is retrieved from cache or wal", blockKey)
}

func (bctsmT *blockCacheTestSM) GetBlockFromNowhere(t *rapid.T) { // Unsuccessfully
	blocksToChoose := util.Intersection(bctsmT.blocksNotInWAL, bctsmT.blocksNotInCache(t))
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
	for _, blockKey := range util.RemoveAll(bctsmT.blocksNotInWAL, maps.Keys(bctsmT.blocks)) {
		require.True(t, bctsmT.wal.Contains(blockKey.AsBlockHash()))
	}
}

func TestBlockCachePropBasedFull(t *testing.T) {
	rapid.Check(t, rapid.Run[*blockCacheTestSM]())
}
