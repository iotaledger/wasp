package sm_snapshots

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/kvstore/mapdb"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/runtime/ioutils"
	"github.com/iotaledger/wasp/packages/chain/statemanager/sm_gpa/sm_gpa_utils"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
)

const localSnapshotsPathConst = "testSnapshots"

func TestSnapshotManagerLocal(t *testing.T) {
	createFun := func(chainID isc.ChainID, store state.Store, log *logger.Logger) SnapshotManager {
		snapshotManager, err := NewSnapshotManager(context.Background(), nil, chainID, 0, localSnapshotsPathConst, []string{}, store, log)
		require.NoError(t, err)
		return snapshotManager
	}
	defer cleanupAfterTest(t)

	testSnapshotManagerSimple(t, createFun, func(isc.ChainID, []SnapshotInfo) {})
}

func TestSnapshotManagerNetwork(t *testing.T) {
	log := testlogger.NewLogger(t)
	defer log.Sync()

	err := ioutils.CreateDirectory(localSnapshotsPathConst, 0o777)
	require.NoError(t, err)

	port := ":9999"
	handler := http.FileServer(http.Dir(localSnapshotsPathConst))
	go http.ListenAndServe(port, handler)

	createFun := func(chainID isc.ChainID, store state.Store, log *logger.Logger) SnapshotManager {
		networkPaths := []string{"http://localhost" + port + "/"}
		snapshotManager, err := NewSnapshotManager(context.Background(), nil, chainID, 0, "nonexistent", networkPaths, store, log)
		require.NoError(t, err)
		return snapshotManager
	}
	defer cleanupAfterTest(t)

	createIndexFileFun := func(chainID isc.ChainID, snapshotInfos []SnapshotInfo) {
		indexFilePath := filepath.Join(localSnapshotsPathConst, chainID.String(), constIndexFileName)
		f, err := os.OpenFile(indexFilePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o666)
		require.NoError(t, err)
		defer f.Close()
		w := bufio.NewWriter(f)
		for _, snapshotInfo := range snapshotInfos {
			w.WriteString(snapshotFileName(snapshotInfo.StateIndex(), snapshotInfo.BlockHash()) + "\n")
		}
		w.Flush()
	}
	testSnapshotManagerSimple(t, createFun, createIndexFileFun)
}

func testSnapshotManagerSimple(
	t *testing.T,
	createNewNodeFun func(isc.ChainID, state.Store, *logger.Logger) SnapshotManager,
	snapshotsAvailableFun func(isc.ChainID, []SnapshotInfo),
) {
	log := testlogger.NewLogger(t)
	defer log.Sync()

	numberOfBlocks := 10
	snapshotCreatePeriod := 2

	var err error
	factory := sm_gpa_utils.NewBlockFactory(t)
	blocks := factory.GetBlocks(numberOfBlocks, 1)
	storeOrig := factory.GetStore()
	snapshotManagerOrig, err := NewSnapshotManager(context.Background(), nil, factory.GetChainID(), uint32(snapshotCreatePeriod), localSnapshotsPathConst, []string{}, storeOrig, log)
	require.NoError(t, err)

	// "Running" node, making snapshots
	for _, block := range blocks {
		snapshotManagerOrig.BlockCommittedAsync(NewSnapshotInfo(block.StateIndex(), block.L1Commitment()))
	}
	for i := snapshotCreatePeriod - 1; i < numberOfBlocks; i += snapshotCreatePeriod {
		require.True(t, waitForBlock(t, snapshotManagerOrig, blocks[i], 10, 50*time.Millisecond))
	}
	createdSnapshots := make([]SnapshotInfo, 0)
	for _, block := range blocks {
		exists := snapshotManagerOrig.SnapshotExists(block.StateIndex(), block.L1Commitment())
		if block.StateIndex()%uint32(snapshotCreatePeriod) == 0 {
			require.True(t, exists)
			createdSnapshots = append(createdSnapshots, NewSnapshotInfo(block.StateIndex(), block.L1Commitment()))
		} else {
			require.False(t, exists)
		}
	}
	snapshotsAvailableFun(factory.GetChainID(), createdSnapshots)

	// Node is restarted
	storeNew := state.NewStore(mapdb.NewMapDB())
	snapshotManagerNew := createNewNodeFun(factory.GetChainID(), storeNew, log)

	// Wait for node to read the list of snapshots
	lastBlock := blocks[len(blocks)-1]
	require.True(t, waitForBlock(t, snapshotManagerNew, lastBlock, 10, 50*time.Millisecond))
	require.True(t, loadAndWaitLoaded(t, snapshotManagerNew, NewSnapshotInfo(lastBlock.StateIndex(), lastBlock.L1Commitment()), 10, 50*time.Millisecond))

	// Check the loaded snapshot
	for i := 0; i < len(blocks)-1; i++ {
		require.False(t, storeNew.HasTrieRoot(blocks[i].TrieRoot()))
	}
	require.True(t, storeNew.HasTrieRoot(lastBlock.TrieRoot()))

	sm_gpa_utils.CheckBlockInStore(t, storeNew, lastBlock)
	sm_gpa_utils.CheckStateInStores(t, storeOrig, storeNew, lastBlock.L1Commitment())
}

func waitForBlock(t *testing.T, snapshotManager SnapshotManager, block state.Block, maxIterations int, sleep time.Duration) bool {
	updateAndWaitFun := func() {
		snapshotManager.UpdateAsync()
		time.Sleep(sleep)
	}
	snapshotExistsFun := func() bool { return snapshotManager.SnapshotExists(block.StateIndex(), block.L1Commitment()) }
	return ensureTrue(t, fmt.Sprintf("block %v to be committed", block.StateIndex()), snapshotExistsFun, maxIterations, updateAndWaitFun)
}

func loadAndWaitLoaded(t *testing.T, snapshotManager SnapshotManager, snapshotInfo SnapshotInfo, maxIterations int, sleep time.Duration) bool {
	respChan := snapshotManager.LoadSnapshotAsync(snapshotInfo)
	loadCompletedFun := func() bool {
		select {
		case result := <-respChan:
			require.NoError(t, result)
			return true
		default:
			return false
		}
	}
	waitFun := func() { time.Sleep(sleep) }
	return ensureTrue(t, fmt.Sprintf("state %v to be loaded", snapshotInfo.StateIndex()), loadCompletedFun, maxIterations, waitFun)
}

func ensureTrue(t *testing.T, title string, predicate func() bool, maxIterations int, step func()) bool {
	if predicate() {
		return true
	}
	for i := 1; i < maxIterations; i++ {
		t.Logf("Waiting for %s iteration %v", title, i)
		step()
		if predicate() {
			return true
		}
	}
	return false
}

func cleanupAfterTest(t *testing.T) {
	err := os.RemoveAll(localSnapshotsPathConst)
	require.NoError(t, err)
}
