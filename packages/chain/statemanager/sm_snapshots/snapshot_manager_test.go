package sm_snapshots

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	//"github.com/iotaledger/hive.go/kvstore/mapdb"
	"github.com/iotaledger/wasp/packages/chain/statemanager/sm_gpa/sm_gpa_utils"
	//"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
)

const localSnapshotsPathConst = "testSnapshots"

func TestBlockCommitted(t *testing.T) {
	log := testlogger.NewLogger(t)
	defer log.Sync()

	var err error
	numberOfBlocks := 10
	factory := sm_gpa_utils.NewBlockFactory(t)
	blocks := factory.GetBlocks(numberOfBlocks, 1)
	store := factory.GetStore()
	snapshotManager, err := NewSnapshotManager(context.Background(), nil, factory.GetChainID(), 2, localSnapshotsPathConst, []string{}, store, log)
	defer cleanupAfterTest(t)
	require.NoError(t, err)
	for _, block := range blocks {
		snapshotManager.BlockCommittedAsync(NewSnapshotInfo(block.StateIndex(), block.L1Commitment()))
	}
	time.Sleep(5 * time.Second)
}

func cleanupAfterTest(t *testing.T) {
	err := os.RemoveAll(localSnapshotsPathConst)
	require.NoError(t, err)
}
