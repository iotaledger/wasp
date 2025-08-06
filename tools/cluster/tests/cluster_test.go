package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/iotaledger/wasp/v2/clients/apiclient"
	"github.com/iotaledger/wasp/v2/packages/isc"
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

func TestClusterRotateChain(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping cluster tests in short mode")
	}

	clu := newCluster(t, waspClusterOpts{nNodes: 4})
	addr, err := clu.RunDKG([]int{0, 1, 2, 3}, 3)
	require.NoError(t, err)

	allNodes := clu.Config.AllNodes()
	chain, err := clu.DeployChain(allNodes, allNodes, 3, addr)
	require.NoError(t, err)

	stateAddressStr := chain.StateAddress.String()

	fmt.Printf("chain.StateAddress: %s\n", stateAddressStr)

	newAddr, err := clu.RunDKG([]int{0, 1, 2, 3}, 3)
	require.NoError(t, err)
	newAddrStr := newAddr.String()

	fmt.Printf("migrating to address: %s\n", newAddrStr)

	client := clu.WaspClientFromHostName(clu.Config.APIHost(0))

	block, _, err := client.CorecontractsAPI.BlocklogGetLatestBlockInfo(context.Background()).Execute()
	require.NoError(t, err)
	fmt.Printf("old block index: %+v\n", block.BlockIndex)

	rotateRequest := client.ChainsAPI.RotateChain(context.Background()).RotateRequest(apiclient.RotateChainRequest{
		RotateToAddress: &newAddrStr,
	})

	response, err := rotateRequest.Execute()
	require.NoError(t, err)
	fmt.Printf("response: %+v\n", response)

	newKeyPair, _, err := clu.NewKeyPairWithFunds()
	require.NoError(t, err)

	_, err = chain.Client(newKeyPair).DepositFunds(1 * isc.Million)
	require.NoError(t, err)

	info, _, err := client.ChainsAPI.GetCommitteeInfo(context.Background()).Execute()
	require.NoError(t, err)
	fmt.Printf("info: %+v\n", info.StateAddress)

	require.Equal(t, newAddrStr, info.StateAddress, "state address should be updated")

	block, _, err = client.CorecontractsAPI.BlocklogGetLatestBlockInfo(context.Background()).Execute()
	require.NoError(t, err)
	fmt.Printf("new block index: %+v\n", block.BlockIndex)

	t.FailNow()

}
