//nolint:unused // false positives because of rapid.Check
package smGPAUtils

import (
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/exp/maps"
	"pgregory.net/rapid"

	"github.com/iotaledger/wasp/packages/util"
)

type blockCacheTestSM struct { // State machine for block cache property based Rapid tests
	*blockCacheNoWALTestSM
	wal BlockWAL
}

func (bctsmT *blockCacheTestSM) Init(t *rapid.T) {
	bctsmT.blockCacheNoWALTestSM = &blockCacheNoWALTestSM{}
	bctsmT.wal = NewMockedBlockWAL()
	bctsmT.blockCacheNoWALTestSM.initWAL(t, bctsmT.wal)
}

// Cleanup() // inheritted from blockCacheNoWALTestSM

func (bctsmT *blockCacheTestSM) Check(t *rapid.T) {
	bctsmT.blockCacheNoWALTestSM.Check(t)
	bctsmT.invariantAllBlocksInWAL(t)
}

// AddNewBlock(t *rapid.T) // inheritted from blockCacheNoWALTestSM
// AddExistingBlock(t *rapid.T) // inheritted from blockCacheNoWALTestSM
// CleanCache(t *rapid.T) // inheritted from blockCacheNoWALTestSM
// GetBlockFromCache(t *rapid.T) // inheritted from blockCacheNoWALTestSM

func (bctsmT *blockCacheTestSM) GetBlockFromWAL(t *rapid.T) { // Unsuccessfully
	if len(bctsmT.blocks) == 0 {
		t.Skip()
	}
	blockKey := rapid.SampledFrom(maps.Keys(bctsmT.blocks)).Example()
	if util.Contains(blockKey, bctsmT.blocksInCache) {
		t.Skip()
	}
	blockExpected, ok := bctsmT.blocks[blockKey]
	require.True(t, ok)
	commitment := blockExpected.L1Commitment()
	block := bctsmT.bc.GetBlock(commitment)
	require.Nil(t, block)
	require.True(t, bctsmT.wal.Contains(blockExpected.Hash()))
	t.Logf("Block %s is not retrieved from WAL", commitment)
}

// Restart(t *rapid.T) // inheritted from blockCacheNoWALTestSM

func (bctsmT *blockCacheTestSM) invariantAllBlocksInWAL(t *rapid.T) {
	for _, block := range bctsmT.blocks {
		require.True(t, bctsmT.wal.Contains(block.Hash()))
	}
}

func TestBlockCacheFullPropBased(t *testing.T) {
	rapid.Check(t, rapid.Run[*blockCacheTestSM]())
}
