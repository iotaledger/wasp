package snapshots

import (
	"bufio"
	"context"
	"fmt"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/log"
	"github.com/iotaledger/hive.go/runtime/ioutils"
	"github.com/iotaledger/wasp/v2/packages/chain/statemanager/gpa/utils"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/kvstore/mapdb"
	"github.com/iotaledger/wasp/v2/packages/metrics"
	"github.com/iotaledger/wasp/v2/packages/state"
	"github.com/iotaledger/wasp/v2/packages/state/statetest"
	"github.com/iotaledger/wasp/v2/packages/testutil/testlogger"
)

const localSnapshotsPathConst = "testSnapshots"

type (
	createNewNodeFun      func(isc.ChainID, *state.BlockHash, state.Store, log.Logger) SnapshotManager
	snapshotsAvailableFun func(isc.ChainID, []SnapshotInfo)
)

var (
	localSnapshotsCreatePathConst   = filepath.Join(localSnapshotsPathConst, "create")
	localSnapshotsDownloadPathConst = filepath.Join(localSnapshotsPathConst, "download")
)

func TestSnapshotManagerLocal(t *testing.T) {
	testSnapshotManager(t, getLocalFuns, testSnapshotManagerLast)
}

func TestSnapshotManagerNetworkHTTP(t *testing.T) {
	testSnapshotManager(t, getNetworkHTTPFuns, testSnapshotManagerLast)
}

func TestSnapshotManagerNetworkFile(t *testing.T) {
	testSnapshotManager(t, getNetworkFileFuns, testSnapshotManagerLast)
}

func TestSnapshotManagerLoadMiddleLocal(t *testing.T) {
	testSnapshotManager(t, getLocalFuns, testSnapshotManagerMiddle)
}

func TestSnapshotManagerLoadMiddleNetworkHTTP(t *testing.T) {
	testSnapshotManager(t, getNetworkHTTPFuns, testSnapshotManagerMiddle)
}

func TestSnapshotManagerLoadMiddleNetworkFile(t *testing.T) {
	testSnapshotManager(t, getNetworkFileFuns, testSnapshotManagerMiddle)
}

func testSnapshotManager(
	t *testing.T,
	getFunsFun func(*testing.T) (createNewNodeFun, snapshotsAvailableFun),
	runTestFun func(*testing.T, createNewNodeFun, snapshotsAvailableFun),
) {
	createFun, snapshotAvailableFun := getFunsFun(t)
	runTestFun(t, createFun, snapshotAvailableFun)
}

func getLocalFuns(t *testing.T) (createNewNodeFun, snapshotsAvailableFun) {
	return func(chainID isc.ChainID, snapshotToLoad *state.BlockHash, store state.Store, log log.Logger) SnapshotManager {
			snapshotManager, err := NewSnapshotManager(
				context.Background(),
				nil,
				chainID,
				snapshotToLoad,
				0,
				0,
				localSnapshotsCreatePathConst,
				[]string{},
				store,
				mockSnapshotsMetrics(),
				log,
			)
			require.NoError(t, err)
			return snapshotManager
		},
		func(isc.ChainID, []SnapshotInfo) {}
}

func getNetworkHTTPFuns(t *testing.T) (createNewNodeFun, snapshotsAvailableFun) {
	err := ioutils.CreateDirectory(localSnapshotsCreatePathConst, 0o777)
	require.NoError(t, err)

	port := ":9999"
	startServer(t, port, http.FileServer(http.Dir(localSnapshotsCreatePathConst)))

	return getNetworkFuns(t, []string{"http://localhost" + port + "/"})
}

func getNetworkFileFuns(t *testing.T) (createNewNodeFun, snapshotsAvailableFun) {
	return getNetworkFuns(t, []string{"file://" + localSnapshotsCreatePathConst + "/"})
}

func getNetworkFuns(t *testing.T, networkPaths []string) (createNewNodeFun, snapshotsAvailableFun) {
	return func(chainID isc.ChainID, snapshotToLoad *state.BlockHash, store state.Store, log log.Logger) SnapshotManager {
			snapshotManager, err := NewSnapshotManager(
				context.Background(),
				nil,
				chainID,
				snapshotToLoad,
				0,
				0,
				localSnapshotsDownloadPathConst,
				networkPaths,
				store,
				mockSnapshotsMetrics(),
				log,
			)
			require.NoError(t, err)
			return snapshotManager
		},
		func(chainID isc.ChainID, snapshotInfos []SnapshotInfo) {
			indexFilePath := filepath.Join(localSnapshotsCreatePathConst, chainID.String(), constIndexFileName)
			f, err := os.OpenFile(indexFilePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o666)
			require.NoError(t, err)
			defer f.Close()
			w := bufio.NewWriter(f)
			for _, snapshotInfo := range snapshotInfos {
				w.WriteString(snapshotFileName(snapshotInfo.StateIndex(), snapshotInfo.BlockHash()) + "\n")
			}
			w.Flush()
		}
}

func testSnapshotManagerLast(
	t *testing.T,
	createNewNodeFun createNewNodeFun,
	snapshotsAvailableFun snapshotsAvailableFun,
) {
	testSnapshotManagerAny(t, createNewNodeFun, snapshotsAvailableFun, 0)
}

func testSnapshotManagerMiddle(
	t *testing.T,
	createNewNodeFun createNewNodeFun,
	snapshotsAvailableFun snapshotsAvailableFun,
) {
	testSnapshotManagerAny(t, createNewNodeFun, snapshotsAvailableFun, 2)
}

func testSnapshotManagerAny(
	t *testing.T,
	createNewNodeFun createNewNodeFun,
	snapshotsAvailableFun snapshotsAvailableFun,
	numberBeforeLast int,
) {
	log := testlogger.NewLogger(t)
	defer log.Shutdown()
	defer cleanupAfterSnapshotManagerTest(t)

	numberOfBlocks := 20
	snapshotCreatePeriod := 4
	snapshotDelayPeriod := 3
	snapshotToLoadStateIndex := numberOfBlocks - snapshotCreatePeriod*(numberBeforeLast+int(math.Ceil(float64(snapshotDelayPeriod)/float64(snapshotCreatePeriod))))

	var err error
	factory := utils.NewBlockFactory(t)
	blocks := factory.GetBlocks(numberOfBlocks, 1)
	storeOrig := factory.GetStore()
	snapshotManagerOrig, err := NewSnapshotManager(
		context.Background(),
		nil,
		factory.GetChainID(),
		nil,
		uint32(snapshotCreatePeriod),
		uint32(snapshotDelayPeriod),
		localSnapshotsCreatePathConst,
		[]string{},
		storeOrig,
		mockSnapshotsMetrics(),
		log,
	)
	require.NoError(t, err)
	require.Equal(t, uint32(0), snapshotManagerOrig.GetLoadedSnapshotStateIndex())

	// "Running" node, making snapshots
	for _, block := range blocks {
		snapshotManagerOrig.BlockCommittedAsync(NewSnapshotInfo(block.StateIndex(), block.L1Commitment()))
	}
	for i := snapshotCreatePeriod - 1; i < numberOfBlocks-snapshotDelayPeriod; i += snapshotCreatePeriod {
		require.True(t, waitForBlock(t, factory.GetChainID(), blocks[i], 10, 50*time.Millisecond))
	}
	createdSnapshots := make([]SnapshotInfo, 0)
	for _, block := range blocks {
		exists := snapshotExists(t, factory.GetChainID(), block.StateIndex(), block.L1Commitment())
		if block.StateIndex()%uint32(snapshotCreatePeriod) == 0 && block.StateIndex() <= uint32(numberOfBlocks-snapshotDelayPeriod) {
			require.True(t, exists)
			createdSnapshots = append(createdSnapshots, NewSnapshotInfo(block.StateIndex(), block.L1Commitment()))
		} else {
			require.False(t, exists)
		}
	}
	snapshotsAvailableFun(factory.GetChainID(), createdSnapshots)

	// Node is restarted
	var snapshotToLoad *state.BlockHash
	if numberBeforeLast > 0 {
		blockHash := blocks[snapshotToLoadStateIndex-1].Hash()
		snapshotToLoad = &blockHash
	} else {
		snapshotToLoad = nil
	}
	storeNew := statetest.NewStoreWithUniqueWriteMutex(mapdb.NewMapDB())
	snapshotManagerNew := createNewNodeFun(factory.GetChainID(), snapshotToLoad, storeNew, log)
	require.Equal(t, uint32(snapshotToLoadStateIndex), snapshotManagerNew.GetLoadedSnapshotStateIndex())

	// Check the loaded snapshot
	for i := 0; i < len(blocks); i++ {
		if i == snapshotToLoadStateIndex-1 {
			require.True(t, storeNew.HasTrieRoot(blocks[i].TrieRoot()))
			utils.CheckBlockInStore(t, storeNew, blocks[i])
			utils.CheckStateInStores(t, storeOrig, storeNew, blocks[i].L1Commitment())
		} else {
			require.False(t, storeNew.HasTrieRoot(blocks[i].TrieRoot()))
		}
	}
}

func snapshotExists(t *testing.T, chainID isc.ChainID, stateIndex uint32, commitment *state.L1Commitment) bool {
	path := filepath.Join(localSnapshotsCreatePathConst, chainID.String(), snapshotFileName(stateIndex, commitment.BlockHash()))
	exists, isDir, err := ioutils.PathExists(path)
	require.False(t, isDir)
	require.NoError(t, err)
	return exists
}

func waitForBlock(t *testing.T, chainID isc.ChainID, block state.Block, maxIterations int, sleep time.Duration) bool {
	updateAndWaitFun := func() {
		time.Sleep(sleep)
	}
	snapshotExistsFun := func() bool { return snapshotExists(t, chainID, block.StateIndex(), block.L1Commitment()) }
	return ensureTrue(t, fmt.Sprintf("block %v to be committed", block.StateIndex()), snapshotExistsFun, maxIterations, updateAndWaitFun)
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

func cleanupAfterSnapshotManagerTest(t *testing.T) {
	err := os.RemoveAll(localSnapshotsPathConst)
	require.NoError(t, err)
}

func mockSnapshotsMetrics() *metrics.ChainSnapshotsMetrics {
	return metrics.NewChainMetricsProvider().GetChainMetrics(isc.EmptyChainID()).Snapshots
}
