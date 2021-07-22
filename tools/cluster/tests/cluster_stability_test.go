/**
This test will test the recovery capabilities of the clusterized nodes, where some nodes can break/restart/lay dead at any time.
*/

package tests

import (
	"testing"
	"time"

	"github.com/iotaledger/wasp/contracts/native/inccounter"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/tools/cluster"
	clutest "github.com/iotaledger/wasp/tools/cluster/testutil"
	"github.com/stretchr/testify/require"
)

func setupClusterWreckTest(t *testing.T, clusterSize int, committee []int, quorum uint16) (*cluster.Cluster, *cluster.Chain) {
	t.Log(quorum)

	clu1 := clutest.NewCluster(t, clusterSize)

	addr, err := clu1.RunDKG(committee, quorum)
	require.NoError(t, err)

	t.Logf("generated state address: %s", addr.Base58())

	chain1, err := clu1.DeployChain("chain", clu1.Config.AllNodes(), committee, quorum, addr)
	require.NoError(t, err)
	t.Logf("deployed chainID: %s", chain1.ChainID.Base58())

	description := "testing with inccounter"
	progHash := inccounter.Contract.ProgramHash

	_, err = chain1.DeployContract(incCounterSCName, progHash.String(), description, nil)
	require.NoError(t, err)

	waitUntil(t, contractIsDeployed(chain1, incCounterSCName), clu1.Config.AllNodes(), 50*time.Second, "contract to be deployed")
	return clu1, chain1
}

func spawnClientsAndSendRequests(t *testing.T, numRequests int, quorum uint16, clusterSize int) {
	// Inversion of (2*len(committee))/3 + 1 = quorum
	numCommitteeNodes := 1.5*float64(quorum) - 1.5

	cmt := util.MakeRange(0, int(numCommitteeNodes))
	clu1, chain1 := setupClusterWreckTest(t, clusterSize, cmt, quorum)

	for i := 0; i < numRequests; i++ {
		client := createNewClient(t, clu1, chain1)

		_, err = client.PostRequest(inccounter.FuncIncCounter.Name)
		require.NoError(t, err)
	}

	waitUntil(t, counterEquals(chain1, int64(numRequests)), util.MakeRange(0, clusterSize), 40*time.Second)

	printBlocks(t, chain1, numRequests+3)
}

func TestOngoingFailureWithoutRecovery(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	t.Run("cluster=6,N=4,req=8", func(t *testing.T) {
		const numRequests = 8
		const quorum = 3
		const clusterSize = 6

		spawnClientsAndSendRequests(t, numRequests, quorum, clusterSize)
	})
}
