package tests

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClusterSingleNode(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping cluster tests in short mode")
	}

	// TODO could be interesting to experiment running these in parallel
	// t.Parallel()

	// setup a cluster with a single node
	run := createTestWrapper(t, 1, []int{0})

	t.Run("permitionless access node", func(t *testing.T) { run(t, testPermitionlessAccessNode) })

	t.Run("SDRUC", func(t *testing.T) { run(t, testSDRUC) })

	t.Run("publisher", func(t *testing.T) { run(t, testNanoPublisher) })

	t.Run("spam onledger", func(t *testing.T) { run(t, testSpamOnledger) })
	t.Run("spam offledger", func(t *testing.T) { run(t, testSpamOffLedger) })
	t.Run("spam call wasm views", func(t *testing.T) { run(t, testSpamCallViewWasm) })
}

func TestClusterMultiNodeCommittee(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping cluster tests in short mode")
	}
	// setup a cluster with 4 nodes
	run := createTestWrapper(t, 4, []int{0, 1, 2, 3})

	t.Run("deploy basic", func(t *testing.T) { run(t, testDeployChain) })
	t.Run("deploy contract", func(t *testing.T) { run(t, testDeployContractOnly) })
	t.Run("deploy contract and spawn", func(t *testing.T) { run(t, testDeployContractAndSpawn) })

	t.Run("accountsBasic", func(t *testing.T) { run(t, testBasicAccounts) })
	t.Run("2acccounts", func(t *testing.T) { run(t, testBasic2Accounts) })

	t.Run("small blob", func(t *testing.T) { run(t, testBlobStoreSmallBlob) })
	t.Run("many blobs", func(t *testing.T) { run(t, testBlobStoreManyBlobsNoEncoding) })

	t.Run("post deploy", func(t *testing.T) { run(t, testPostDeployInccounter) })
	t.Run("post 1", func(t *testing.T) { run(t, testPost1Request) })
	t.Run("post 3 recursive", func(t *testing.T) { run(t, testPost3Recursive) })
	t.Run("post 5", func(t *testing.T) { run(t, testPost5Requests) })
	t.Run("post 5 async", func(t *testing.T) { run(t, testPost5AsyncRequests) })

	t.Run("EVM jsonrpc", func(t *testing.T) { run(t, testEVMJsonRPCCluster) })

	t.Run("maintenance", func(t *testing.T) { run(t, testMaintenance) })

	t.Run("offledger basic", func(t *testing.T) { run(t, testOffledgerRequest) })
	t.Run("offledger 900KB", func(t *testing.T) { run(t, testOffledgerRequest900KB) })
	t.Run("offledger nonce", func(t *testing.T) { run(t, testOffledgerNonce) })

	t.Run("inccounter invalid entrypoint", func(t *testing.T) { run(t, testInvalidEntrypoint) })
	t.Run("inccounter increment", func(t *testing.T) { run(t, testIncrement) })
	t.Run("inccounter increment with transfer", func(t *testing.T) { run(t, testIncrementWithTransfer) })
	t.Run("inccounter increment call", func(t *testing.T) { run(t, testIncCallIncrement1) })
	t.Run("inccounter increment recursive", func(t *testing.T) { run(t, testIncCallIncrement2Recurse5x) })
	t.Run("inccounter increment post", func(t *testing.T) { run(t, testIncPostIncrement) })
	t.Run("inccounter increment repeatmany", func(t *testing.T) { run(t, testIncRepeatManyIncrement) })
	t.Run("inccounter local state internal call ", func(t *testing.T) { run(t, testIncLocalStateInternalCall) })
	t.Run("inccounter local state sandbox call", func(t *testing.T) { run(t, testIncLocalStateSandboxCall) })
	t.Run("inccounter local state post", func(t *testing.T) { run(t, testIncLocalStatePost) })
	t.Run("inccounter view counter", func(t *testing.T) { run(t, testIncViewCounter) })
	t.Run("inccounter timelock", func(t *testing.T) { run(t, testIncCounterTimelock) })
}

func createTestWrapper(tt *testing.T, clusterSize int, committee []int) func(t *testing.T, f func(*testing.T, *ChainEnv)) {
	dkgQuorum := uint16((2*len(committee))/3 + 1)
	clu := newCluster(tt, waspClusterOpts{nNodes: clusterSize})
	dkgAddr, err := clu.RunDKG(committee, dkgQuorum)
	require.NoError(tt, err)

	return func(t *testing.T, f func(*testing.T, *ChainEnv)) {
		// create a fresh new chain for the test
		allNodes := clu.Config.AllNodes()
		chain, err := clu.DeployChain("testChain", allNodes, allNodes, dkgQuorum, dkgAddr)
		require.NoError(t, err)
		env := newChainEnv(t, clu, chain)

		t.Cleanup(func() {
			clu.MultiClient().DeactivateChain(chain.ChainID)
		})
		f(t, env)
	}
}
