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
	run := createTestWrapper(t, 1, []int{0})

	t.Run("permissionless access node", func(t *testing.T) { run(t, testPermissionlessAccessNode) }) // passed

	t.Run("spam onledger", func(t *testing.T) { run(t, testSpamOnledger) })
	t.Run("spam offledger", func(t *testing.T) { run(t, testSpamOffLedger) })
	t.Run("spam EVM", func(t *testing.T) { run(t, testSpamEVM) })
	t.Run("accounts dump", func(t *testing.T) { run(t, testDumpAccounts) }) // passed
}

func TestClusterMultiNodeCommittee(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping cluster tests in short mode")
	}

	// setup a cluster with 4 nodes
	run := createTestWrapper(t, 4, []int{0, 1, 2, 3})

	t.Run("deploy basic", func(t *testing.T) { run(t, testDeployChain) }) // passed

	t.Run("accountsBasic", func(t *testing.T) { run(t, testBasicAccounts) }) // passed
	t.Run("2accounts", func(t *testing.T) { run(t, testBasic2Accounts) })    // passed

	t.Run("post 1", func(t *testing.T) { run(t, testPost1Request) })             // passed
	t.Run("post 3", func(t *testing.T) { run(t, testPost3Requests) })            // passed
	t.Run("post 5 async", func(t *testing.T) { run(t, testPost5AsyncRequests) }) // passed

	t.Run("EVM jsonrpc", func(t *testing.T) { run(t, testEVMJsonRPCCluster) }) // passed

	t.Run("offledger basic", func(t *testing.T) { run(t, testOffledgerRequest) }) // passed
	t.Run("offledger nonce", func(t *testing.T) { run(t, testOffledgerNonce) })   // passed

	t.Run("webapi ISC estimategas onledger", func(t *testing.T) { run(t, testEstimateGasOnLedger) })   // passed
	t.Run("webapi ISC estimategas offledger", func(t *testing.T) { run(t, testEstimateGasOffLedger) }) // passed
}

func createTestWrapper(tt *testing.T, clusterSize int, committee []int) func(t *testing.T, f func(*testing.T, *ChainEnv)) {
	dkgQuorum := uint16((2*len(committee))/3 + 1)
	clu := newCluster(tt, waspClusterOpts{nNodes: clusterSize})
	dkgAddr, err := clu.RunDKG(committee, dkgQuorum)
	require.NoError(tt, err)

	return func(t *testing.T, f func(*testing.T, *ChainEnv)) {
		// create a fresh new chain for the test
		allNodes := clu.Config.AllNodes()
		chain, err := clu.DeployChain(allNodes, allNodes, dkgQuorum, dkgAddr)
		require.NoError(t, err)
		env := newChainEnv(t, clu, chain)

		t.Cleanup(func() {
			clu.MultiClient().DeactivateChain()
		})
		f(t, env)
	}
}
