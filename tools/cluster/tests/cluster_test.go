package tests

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClusterSingleNode(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping cluster tests in short mode")
	}

	// setup a cluster with a single node
	env := createTestWrapper(t, 1, []int{0})

	t.Run("permissionless access node", env.testPermissionlessAccessNode)

	t.Run("spam onledger", env.testSpamOnledger)
	t.Run("spam offledger", env.testSpamOffLedger)
	t.Run("spam EVM", env.testSpamEVM)
	t.Run("accounts dump", env.testDumpAccounts)
}

func TestClusterMultiNodeCommittee(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping cluster tests in short mode")
	}

	// setup a cluster with 4 nodes
	env := createTestWrapper(t, 4, []int{0, 1, 2, 3})

	t.Run("deploy basic", env.testDeployChain)

	t.Run("accountsBasic", env.testBasicAccounts)
	t.Run("2accounts", env.testBasic2Accounts)

	t.Run("post 1", env.testPost1Request)
	t.Run("post 3", env.testPost3Requests)
	t.Run("post 5 async", env.testPost5AsyncRequests)

	t.Run("EVM jsonrpc", env.testEVMJsonRPCCluster)

	t.Run("offledger basic", env.testOffledgerRequest)
	t.Run("offledger nonce", env.testOffledgerNonce)

	t.Run("webapi ISC estimategas onledger", env.testEstimateGasOnLedger)
	t.Run("webapi ISC estimategas offledger", env.testEstimateGasOffLedger)
}

func createTestWrapper(t *testing.T, clusterSize int, committee []int) *ChainEnv {
	dkgQuorum := uint16((2*len(committee))/3 + 1)
	clu := newCluster(t, waspClusterOpts{nNodes: clusterSize})
	dkgAddr, err := clu.RunDKG(committee, dkgQuorum)
	require.NoError(t, err)

	// create a fresh new chain for the test
	allNodes := clu.Config.AllNodes()
	chain, err := clu.DeployChain(allNodes, allNodes, dkgQuorum, dkgAddr)
	require.NoError(t, err)
	env := newChainEnv(t, clu, chain)

	t.Cleanup(func() {
		clu.MultiClient().DeactivateChain()
	})
	return env
}
