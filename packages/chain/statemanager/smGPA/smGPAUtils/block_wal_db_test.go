package smGPAUtils

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/core/kvstore/mapdb"
	"github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
)

// const constTestFolder = "basicWALTest" // TODO: define or remove

func TestRecoverDBFromWALBasic(t *testing.T) {
	log := testlogger.NewLogger(t)
	defer log.Sync()
	defer cleanupAfterTest(t)

	factory := NewBlockFactory(t)
	blocks := factory.GetBlocks(5, 1)
	testRecoverDBFromWALBasic(t, log, factory.GetChainID(), blocks)
}

func TestRecoverDBFromWALFull(t *testing.T) {
	log := testlogger.NewLogger(t)
	defer log.Sync()
	defer cleanupAfterTest(t)

	factory := NewBlockFactory(t)
	blocks0 := factory.GetBlocks(10, 1)
	blocks1 := factory.GetBlocksFrom(27, 3, blocks0[6].L1Commitment(), 2)
	blocks2 := factory.GetBlocksFrom(15, 1, state.OriginL1Commitment(), 3)
	allBlocks := blocks0
	allBlocks = append(allBlocks, blocks1...)
	allBlocks = append(allBlocks, blocks2...)
	testRecoverDBFromWALBasic(t, log, factory.GetChainID(), allBlocks)
}

func testRecoverDBFromWALBasic(t *testing.T, log *logger.Logger, chainID isc.ChainID, blocks []state.Block) {
	wal, err := NewBlockWAL(log, constTestFolder, chainID, NewBlockWALMetrics())
	require.NoError(t, err)
	for i := range blocks {
		err := wal.Write(blocks[i])
		require.NoError(t, err)
	}

	db := state.InitChainStore(mapdb.NewMapDB())
	for i := range blocks {
		require.False(t, db.HasTrieRoot(blocks[i].L1Commitment().TrieRoot()))
	}
	FillDBFromBlockWAL(db, wal)
	for i := range blocks {
		require.True(t, db.HasTrieRoot(blocks[i].L1Commitment().TrieRoot()))
	}
}

/*func cleanupAfterTest(t *testing.T) {		// TODO: define or remove
	err := os.RemoveAll(constTestFolder)
	require.NoError(t, err)
}*/
