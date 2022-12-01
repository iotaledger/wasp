package smGPAUtils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/exp/maps"
	"pgregory.net/rapid"

	"github.com/iotaledger/hive.go/core/kvstore"
	"github.com/iotaledger/hive.go/core/kvstore/mapdb"
	"github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/wasp/packages/common"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/packages/util"
)

type blockChangeCallbackFun func(*rapid.T, state.BlockHash)

type blockCacheNoWALTestSM struct { // State machine for block cache no WAL property based Rapid tests
	bc            BlockCache
	chainID       *isc.ChainID
	ao            *isc.AliasOutputWithID
	vs            state.VirtualStateAccess
	store         kvstore.KVStore
	blocks        map[state.BlockHash]state.Block
	blockTimes    []*blockTime
	blocksInCache []state.BlockHash
	blocksInDB    []state.BlockHash
	onAddBlockFun blockChangeCallbackFun
	log           *logger.Logger
}

var emptyBlockChangeCallbackFun blockChangeCallbackFun = func(*rapid.T, state.BlockHash) {}

func (bcnwtsmT *blockCacheNoWALTestSM) initWAL(t *rapid.T, wal BlockWAL, onAddBlockFun blockChangeCallbackFun) {
	var err error
	bcnwtsmT.store = mapdb.NewMapDB()
	bcnwtsmT.log = testlogger.NewLogger(t)
	bcnwtsmT.bc, err = NewBlockCache(bcnwtsmT.store, NewDefaultTimeProvider(), wal, bcnwtsmT.log)
	require.NoError(t, err)
	bcnwtsmT.chainID, bcnwtsmT.ao, bcnwtsmT.vs = GetOriginState(t)
	bcnwtsmT.blockTimes = make([]*blockTime, 0)
	bcnwtsmT.blocks = make(map[state.BlockHash]state.Block)
	bcnwtsmT.blocksInCache = make([]state.BlockHash, 0)
	bcnwtsmT.blocksInDB = make([]state.BlockHash, 0)
	bcnwtsmT.onAddBlockFun = onAddBlockFun
}

func (bcnwtsmT *blockCacheNoWALTestSM) Init(t *rapid.T) {
	bcnwtsmT.initWAL(t, NewEmptyBlockWAL(), emptyBlockChangeCallbackFun)
}

func (bcnwtsmT *blockCacheNoWALTestSM) Cleanup() {
	bcnwtsmT.log.Sync()
}

func (bcnwtsmT *blockCacheNoWALTestSM) Check(t *rapid.T) {
	bcnwtsmT.invariantAllBlocksInCacheDifferent(t)
	bcnwtsmT.invariantAllBlocksInDBDifferent(t)
	bcnwtsmT.invariantBlocksInCacheBijectionToBlockTimes(t)
}

func (bcnwtsmT *blockCacheNoWALTestSM) AddNewBlock(t *rapid.T) {
	block, aliasOutput, virtualState := GetNextState(t, bcnwtsmT.vs, bcnwtsmT.ao)
	bcnwtsmT.ao = aliasOutput
	bcnwtsmT.vs = virtualState
	bcnwtsmT.addBlock(t, block)
	t.Logf("New block %s added to cache", block.GetHash())
}

func (bcnwtsmT *blockCacheNoWALTestSM) AddExistingBlock(t *rapid.T) {
	if len(bcnwtsmT.blocksInCache) == len(bcnwtsmT.blocks) {
		t.Skip()
	}
	blockHash := rapid.SampledFrom(bcnwtsmT.blocksNotInCache(t)).Example()
	block, ok := bcnwtsmT.blocks[blockHash]
	require.True(t, ok)
	bcnwtsmT.addBlock(t, block)
	t.Logf("Block %s added to cache again", block.GetHash())
}

func (bcnwtsmT *blockCacheNoWALTestSM) WriteBlockToDb(t *rapid.T) {
	if len(bcnwtsmT.blocks) == 0 {
		t.Skip()
	}
	blockHash := rapid.SampledFrom(maps.Keys(bcnwtsmT.blocks)).Example()
	if ContainsBlockHash(blockHash, bcnwtsmT.blocksInDB) {
		t.Skip()
	}
	block, ok := bcnwtsmT.blocks[blockHash]
	require.True(t, ok)
	batch, err := bcnwtsmT.store.Batched()
	require.NoError(t, err)
	key := common.MakeKey(common.ObjectTypeBlock, util.Uint32To4Bytes(block.BlockIndex()))
	err = batch.Set(key, block.Bytes())
	require.NoError(t, err)
	err = batch.Commit()
	require.NoError(t, err)
	err = bcnwtsmT.store.Flush()
	require.NoError(t, err)
	bcnwtsmT.blocksInDB = append(bcnwtsmT.blocksInDB, blockHash)
	t.Logf("Block %s written to DB", block.GetHash())
}

func (bcnwtsmT *blockCacheNoWALTestSM) CleanCache(t *rapid.T) {
	if len(bcnwtsmT.blockTimes) == 0 {
		t.Skip()
	}
	index := rapid.Uint32Range(0, uint32(len(bcnwtsmT.blockTimes)-1)).Example()
	time := bcnwtsmT.blockTimes[index].time
	bcnwtsmT.bc.CleanOlderThan(time)
	for i := uint32(0); i <= index; i++ {
		blockHash := bcnwtsmT.blockTimes[i].blockHash
		bcnwtsmT.blocksInCache = DeleteBlockHash(blockHash, bcnwtsmT.blocksInCache)
		t.Logf("Block %s deleted from cache", blockHash)
	}
	bcnwtsmT.blockTimes = bcnwtsmT.blockTimes[index+1:]
	t.Logf("Cache cleaned until %v", time)
}

func (bcnwtsmT *blockCacheNoWALTestSM) GetBlockFromCache(t *rapid.T) {
	if len(bcnwtsmT.blocksInCache) == 0 {
		t.Skip()
	}
	blockHash := rapid.SampledFrom(bcnwtsmT.blocksInCache).Example()
	if ContainsBlockHash(blockHash, bcnwtsmT.blocksInDB) {
		t.Skip()
	}
	bcnwtsmT.tstGetBlockFromCache(t, blockHash)
}

func (bcnwtsmT *blockCacheNoWALTestSM) tstGetBlockFromCache(t *rapid.T, blockHash state.BlockHash) {
	bcnwtsmT.getAndCheckBlock(t, blockHash)
	t.Logf("Block from cache %s obtained", blockHash)
}

func (bcnwtsmT *blockCacheNoWALTestSM) GetBlockFromDB(t *rapid.T) {
	if len(bcnwtsmT.blocksInDB) == 0 {
		t.Skip()
	}
	blockHash := rapid.SampledFrom(bcnwtsmT.blocksInDB).Example()
	if ContainsBlockHash(blockHash, bcnwtsmT.blocksInCache) {
		t.Skip()
	}
	bcnwtsmT.tstGetBlockFromDB(t, blockHash)
}

func (bcnwtsmT *blockCacheNoWALTestSM) tstGetBlockFromDB(t *rapid.T, blockHash state.BlockHash) {
	bcnwtsmT.tstGetBlockNoCache(t, blockHash)
	t.Logf("Block from DB %s obtained", blockHash)
}

func (bcnwtsmT *blockCacheNoWALTestSM) tstGetBlockNoCache(t *rapid.T, blockHash state.BlockHash) {
	bcnwtsmT.getAndCheckBlock(t, blockHash)
	block, ok := bcnwtsmT.blocks[blockHash]
	require.True(t, ok)
	bcnwtsmT.addBlock(t, block)
}

func (bcnwtsmT *blockCacheNoWALTestSM) GetBlockFromCacheAndDB(t *rapid.T) {
	if (len(bcnwtsmT.blocksInCache) == 0) || len(bcnwtsmT.blocksInDB) == 0 {
		t.Skip()
	}
	blockHash := rapid.SampledFrom(bcnwtsmT.blocksInDB).Example()
	if !ContainsBlockHash(blockHash, bcnwtsmT.blocksInCache) {
		t.Skip()
	}
	bcnwtsmT.tstGetBlockFromCacheAndDB(t, blockHash)
}

func (bcnwtsmT *blockCacheNoWALTestSM) tstGetBlockFromCacheAndDB(t *rapid.T, blockHash state.BlockHash) {
	bcnwtsmT.getAndCheckBlock(t, blockHash)
	t.Logf("Block from cache and DB %s obtained", blockHash)
}

func (bcnwtsmT *blockCacheNoWALTestSM) GetLostBlock(t *rapid.T) {
	if len(bcnwtsmT.blocksInCache) == len(bcnwtsmT.blocks) {
		t.Skip()
	}
	blockHash := rapid.SampledFrom(bcnwtsmT.blocksNotInCache(t)).Example()
	if ContainsBlockHash(blockHash, bcnwtsmT.blocksInDB) {
		t.Skip()
	}
	bcnwtsmT.tstGetLostBlock(t, blockHash)
}

func (bcnwtsmT *blockCacheNoWALTestSM) tstGetLostBlock(t *rapid.T, blockHash state.BlockHash) {
	blockExpected, ok := bcnwtsmT.blocks[blockHash]
	require.True(t, ok)
	block := bcnwtsmT.bc.GetBlock(blockExpected.BlockIndex(), blockHash)
	require.Nil(t, block)
	t.Logf("Lost block %s is unobtainable", blockHash)
}

func (bcnwtsmT *blockCacheNoWALTestSM) Restart(t *rapid.T) {
	var err error
	bcnwtsmT.bc, err = NewBlockCache(bcnwtsmT.store, NewDefaultTimeProvider(), bcnwtsmT.bc.(*blockCache).wal, bcnwtsmT.log)
	require.NoError(t, err)
	bcnwtsmT.blocksInCache = make([]state.BlockHash, 0)
	bcnwtsmT.blockTimes = make([]*blockTime, 0)
	t.Logf("Block cache was restarted")
}

func (bcnwtsmT *blockCacheNoWALTestSM) invariantAllBlocksInCacheDifferent(t *rapid.T) {
	require.True(t, AllDifferentBlockHashes(bcnwtsmT.blocksInCache))
}

func (bcnwtsmT *blockCacheNoWALTestSM) invariantAllBlocksInDBDifferent(t *rapid.T) {
	require.True(t, AllDifferentBlockHashes(bcnwtsmT.blocksInDB))
}

func (bcnwtsmT *blockCacheNoWALTestSM) invariantBlocksInCacheBijectionToBlockTimes(t *rapid.T) {
	blockTimeHashes := make([]state.BlockHash, len(bcnwtsmT.blockTimes))
	for i := range bcnwtsmT.blockTimes {
		blockTimeHashes[i] = bcnwtsmT.blockTimes[i].blockHash
	}
	require.Equal(t, len(bcnwtsmT.blocksInCache), len(blockTimeHashes))
	for i := range bcnwtsmT.blocksInCache {
		require.True(t, ContainsBlockHash(bcnwtsmT.blocksInCache[i], blockTimeHashes))
	}
}

func (bcnwtsmT *blockCacheNoWALTestSM) addBlock(t *rapid.T, block state.Block) {
	blockHash := block.GetHash()
	bcnwtsmT.blocks[blockHash] = block
	err := bcnwtsmT.bc.AddBlock(block)
	require.NoError(t, err)
	require.False(t, ContainsBlockHash(blockHash, bcnwtsmT.blocksInCache))
	bcnwtsmT.blocksInCache = append(bcnwtsmT.blocksInCache, blockHash)
	bcnwtsmT.blockTimes = append(bcnwtsmT.blockTimes, &blockTime{
		time:      time.Now(),
		blockHash: blockHash,
	})
	bcnwtsmT.onAddBlockFun(t, blockHash)
}

func (bcnwtsmT *blockCacheNoWALTestSM) blocksNotInCache(t *rapid.T) []state.BlockHash {
	return RemoveAllBlockHashes(bcnwtsmT.blocksInCache, maps.Keys(bcnwtsmT.blocks))
}

func (bcnwtsmT *blockCacheNoWALTestSM) getAndCheckBlock(t *rapid.T, blockHash state.BlockHash) {
	blockExpected, ok := bcnwtsmT.blocks[blockHash]
	require.True(t, ok)
	block := bcnwtsmT.bc.GetBlock(blockExpected.BlockIndex(), blockHash)
	require.NotNil(t, block)
	require.True(t, blockExpected.Equals(block))
}

func TestBlockCacheNoWALPropBased(t *testing.T) {
	rapid.Check(t, rapid.Run[*blockCacheNoWALTestSM]())
}
