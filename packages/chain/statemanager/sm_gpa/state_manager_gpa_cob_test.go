package sm_gpa

import (
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/kvstore/mapdb"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/chain/statemanager/sm_gpa/sm_gpa_utils"
	"github.com/iotaledger/wasp/packages/origin"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/packages/util/pipe"
)

func initTestChainOfBlocks(t *testing.T) (
	*logger.Logger,
	*sm_gpa_utils.BlockFactory,
	state.Store,
	*stateManagerGPA,
) {
	bf := sm_gpa_utils.NewBlockFactory(t, nil)
	log := testlogger.NewLogger(t)
	store := state.NewStoreWithUniqueWriteMutex(mapdb.NewMapDB())
	smGPA, err := New(bf.GetChainID(), 0, nil, nil, store, mockStateManagerMetrics(), log, NewStateManagerParameters())
	require.NoError(t, err)
	sm, ok := smGPA.(*stateManagerGPA)
	require.True(t, ok)
	origin.InitChain(store, bf.GetChainInitParameters(), 0)
	return log, bf, store, sm
}

func prependOriginBlock(bf *sm_gpa_utils.BlockFactory, blocks []state.Block) []state.Block {
	originBlock := bf.GetOriginBlock()
	return append([]state.Block{originBlock}, blocks...)
}

func blocksToBlockInfos(blocks []state.Block) []*blockInfo {
	return lo.Map(blocks, func(block state.Block, _ int) *blockInfo {
		return &blockInfo{
			trieRoot:   block.TrieRoot(),
			blockIndex: block.StateIndex(),
		}
	})
}

func runTestChainOfBlocks(
	t *testing.T,
	log *logger.Logger,
	bf *sm_gpa_utils.BlockFactory,
	store state.Store,
	sm *stateManagerGPA,
	blocksToCommit []state.Block,
	blocksToPrune []state.Block,
	blocksInChain []state.Block,
	blocksExpected []state.Block,
) {
	defer log.Sync()

	for _, block := range blocksToCommit {
		sd := bf.GetStateDraft(block)
		block2 := store.Commit(sd)
		sm_gpa_utils.CheckBlocksEqual(t, block, block2)
		log.Debugf("Committed block: %v %s", block.StateIndex(), block.L1Commitment())
	}
	for _, block := range blocksToPrune {
		_, err := store.Prune(block.TrieRoot())
		require.NoError(t, err)
		log.Debugf("Pruned block: %v %s", block.StateIndex(), block.L1Commitment())
	}
	if blocksInChain == nil {
		require.Nil(t, sm.chainOfBlocks)
	} else {
		sm.chainOfBlocks = pipe.NewDeque[*blockInfo]()
		for _, bi := range blocksToBlockInfos(blocksInChain) {
			sm.chainOfBlocks.AddEnd(bi)
			log.Debugf("Added block to currently known blocks chain: %v %s", bi.blockIndex, bi.trieRoot)
		}
	}

	lastBlock := blocksToCommit[len(blocksToCommit)-1]
	sm.updateChainOfBlocks(lastBlock.L1Commitment(), lastBlock.StateIndex())
	bisExpected := blocksToBlockInfos(blocksExpected)
	bisActual := sm.chainOfBlocks.PeekAll()
	require.Equal(t, len(bisExpected), len(bisActual))
	for i := range bisExpected {
		log.Debugf("Expecting block: %v %s", bisExpected[i].blockIndex, bisExpected[i].trieRoot)
		require.True(t, bisExpected[i].trieRoot.Equals(bisActual[i].trieRoot))
		require.Equal(t, bisExpected[i].blockIndex, bisActual[i].blockIndex)
	}
}

func TestChainOfBlocksNewChainFullHistory(t *testing.T) {
	totalBlocks := 10
	log, bf, store, sm := initTestChainOfBlocks(t)
	blocksToCommit := bf.GetBlocks(totalBlocks, 1)
	runTestChainOfBlocks(t, log, bf, store, sm, blocksToCommit, nil, nil, prependOriginBlock(bf, blocksToCommit))
}

func TestChainOfBlocksNewChainSomeHistory(t *testing.T) {
	totalBlocks := 10
	prunedBlocks := 5
	log, bf, store, sm := initTestChainOfBlocks(t)
	blocksToCommit := bf.GetBlocks(totalBlocks, 1)
	blocksToPrune := prependOriginBlock(bf, blocksToCommit[:prunedBlocks])
	blocksExpected := blocksToCommit[prunedBlocks:]
	runTestChainOfBlocks(t, log, bf, store, sm, blocksToCommit, blocksToPrune, nil, blocksExpected)
}

func TestChainOfBlocksMergeAllFullHistory(t *testing.T) {
	totalBlocks := 15
	chainEnd := 10
	log, bf, store, sm := initTestChainOfBlocks(t)
	blocksToCommit := bf.GetBlocks(totalBlocks, 1)
	blocksInChain := blocksToCommit[:chainEnd]
	runTestChainOfBlocks(t, log, bf, store, sm, blocksToCommit, nil, blocksInChain, prependOriginBlock(bf, blocksToCommit))
}

func TestChainOfBlocksMergeAllSomeHistory(t *testing.T) {
	totalBlocks := 15
	prunedBlocks := 5
	chainEnd := 10
	log, bf, store, sm := initTestChainOfBlocks(t)
	blocksToCommit := bf.GetBlocks(totalBlocks, 1)
	blocksToPrune := prependOriginBlock(bf, blocksToCommit[:prunedBlocks])
	blocksInChain := blocksToCommit[prunedBlocks:chainEnd]
	blocksExpected := blocksToCommit[prunedBlocks:]
	runTestChainOfBlocks(t, log, bf, store, sm, blocksToCommit, blocksToPrune, blocksInChain, blocksExpected)
}

func testChainOfBlocksMergeMiddleFullHistory(t *testing.T, totalBlocks, branchFrom, branchLength int) {
	log, bf, store, sm := initTestChainOfBlocks(t)
	originalBlocks := bf.GetBlocks(totalBlocks, 1)
	branchBlocks := bf.GetBlocksFrom(branchLength, 1, originalBlocks[branchFrom].L1Commitment(), 2)
	blocksToCommit := append([]state.Block{}, originalBlocks[:branchFrom+1]...)
	blocksToCommit = append(blocksToCommit, branchBlocks...)
	blocksInChain := prependOriginBlock(bf, originalBlocks)
	runTestChainOfBlocks(t, log, bf, store, sm, blocksToCommit, nil, blocksInChain, prependOriginBlock(bf, blocksToCommit))
}

func TestChainOfBlocksMergeMiddleFullHistory(t *testing.T) {
	totalBlocks := 15
	branchFrom := 9
	branchLength := 5
	testChainOfBlocksMergeMiddleFullHistory(t, totalBlocks, branchFrom, branchLength)
}

func TestChainOfBlocksMergeMiddleFullHistoryLonger(t *testing.T) {
	totalBlocks := 15
	branchFrom := 9
	branchLength := 10
	testChainOfBlocksMergeMiddleFullHistory(t, totalBlocks, branchFrom, branchLength)
}

func TestChainOfBlocksMergeMiddleFullHistoryShorter(t *testing.T) {
	totalBlocks := 15
	branchFrom := 9
	branchLength := 3
	testChainOfBlocksMergeMiddleFullHistory(t, totalBlocks, branchFrom, branchLength)
}

func TestChainOfBlocksMergeMiddleSomeHistory(t *testing.T) {
	totalBlocks := 15
	branchFrom := 9
	branchLength := 5
	prunedBlocks := 5
	log, bf, store, sm := initTestChainOfBlocks(t)
	originalBlocks := bf.GetBlocks(totalBlocks, 1)
	branchBlocks := bf.GetBlocksFrom(branchLength, 1, originalBlocks[branchFrom].L1Commitment(), 2)
	blocksToCommit := append([]state.Block{}, originalBlocks[:branchFrom+1]...)
	blocksToCommit = append(blocksToCommit, branchBlocks...)
	blocksToPrune := prependOriginBlock(bf, blocksToCommit[:prunedBlocks])
	blocksInChain := originalBlocks[prunedBlocks:]
	blocksExpected := blocksToCommit[prunedBlocks:]
	runTestChainOfBlocks(t, log, bf, store, sm, blocksToCommit, blocksToPrune, blocksInChain, blocksExpected)
}

func TestChainOfBlocksNoMerge(t *testing.T) {
	totalBlocks := 15
	branchFrom := 9
	branchLength := 5
	log, bf, store, sm := initTestChainOfBlocks(t)
	originalBlocks := bf.GetBlocks(totalBlocks, 1)
	branchBlocks := bf.GetBlocksFrom(branchLength, 1, originalBlocks[branchFrom].L1Commitment(), 2)
	blocksToCommit := append([]state.Block{}, originalBlocks[:branchFrom+1]...)
	blocksToCommit = append(blocksToCommit, branchBlocks...)
	blocksToPrune := prependOriginBlock(bf, originalBlocks[:branchFrom+1])
	blocksInChain := originalBlocks[branchFrom+1:]
	runTestChainOfBlocks(t, log, bf, store, sm, blocksToCommit, blocksToPrune, blocksInChain, branchBlocks)
}

func TestChainOfBlocksMergeAtOnceTooSmallHistory1(t *testing.T) {
	totalBlocks := 10
	chainStart := 5
	log, bf, store, sm := initTestChainOfBlocks(t)
	blocksToCommit := bf.GetBlocks(totalBlocks, 1)
	blocksInChain := blocksToCommit[chainStart:]
	runTestChainOfBlocks(t, log, bf, store, sm, blocksToCommit, nil, blocksInChain, prependOriginBlock(bf, blocksToCommit))
}

func TestChainOfBlocksMergeAtOnceTooSmallHistory2(t *testing.T) {
	totalBlocks := 15
	chainStart := 10
	prunedBlocks := 5
	log, bf, store, sm := initTestChainOfBlocks(t)
	blocksToCommit := bf.GetBlocks(totalBlocks, 1)
	blocksToPrune := prependOriginBlock(bf, blocksToCommit[:prunedBlocks])
	blocksInChain := blocksToCommit[chainStart:]
	blocksExpected := blocksToCommit[prunedBlocks:]
	runTestChainOfBlocks(t, log, bf, store, sm, blocksToCommit, blocksToPrune, blocksInChain, blocksExpected)
}

func TestChainOfBlocksMergeAtOnceTooLargeHistory(t *testing.T) {
	totalBlocks := 15
	chainStart := 5
	prunedBlocks := 10
	log, bf, store, sm := initTestChainOfBlocks(t)
	blocksToCommit := bf.GetBlocks(totalBlocks, 1)
	blocksToPrune := prependOriginBlock(bf, blocksToCommit[:prunedBlocks])
	blocksInChain := blocksToCommit[chainStart:]
	blocksExpected := blocksToCommit[prunedBlocks:]
	runTestChainOfBlocks(t, log, bf, store, sm, blocksToCommit, blocksToPrune, blocksInChain, blocksExpected)
}
