package tests

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/clients/apiclient"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotajsonrpc"
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

	newAddr, err := clu.RunDKG([]int{0, 1, 2, 3}, 3)
	require.NoError(t, err)
	newAddrStr := newAddr.String()

	client := clu.WaspClientFromHostName(clu.Config.APIHost(0))

	block, _, err := client.CorecontractsAPI.BlocklogGetLatestBlockInfo(context.Background()).Execute()
	require.NoError(t, err)

	// performing rotation for all nodes
	for i := 0; i < 3; i++ {
		c := clu.WaspClientFromHostName(clu.Config.APIHost(i))
		rotateRequest := c.ChainsAPI.RotateChain(context.Background()).RotateRequest(apiclient.RotateChainRequest{
			RotateToAddress: &newAddrStr,
		})
		_, err := rotateRequest.Execute()
		require.NoError(t, err)
		time.Sleep(1 * time.Second)
	}

	newKeyPair, _, err := clu.NewKeyPairWithFunds()
	require.NoError(t, err)
	balance, _, err := client.CorecontractsAPI.AccountsGetAccountBalance(context.Background(), newKeyPair.Address().String()).Execute()
	require.NoError(t, err)

	require.Equal(t, balance.BaseTokens, "0", "balance should be 0")

	_, err = chain.Client(newKeyPair).DepositFunds(1e9)
	require.NoError(t, err)

	time.Sleep(3 * time.Second)

	balance, _, err = client.CorecontractsAPI.AccountsGetAccountBalance(context.Background(), newKeyPair.Address().String()).Execute()
	require.NoError(t, err)

	require.Equal(t, balance.BaseTokens, "999900000", "balance should be 1e9 minus fee")

	info, _, err := client.ChainsAPI.GetCommitteeInfo(context.Background()).Execute()
	require.NoError(t, err)

	require.Equal(t, newAddrStr, info.StateAddress, "state address should be updated")

	newBlock, _, err := client.CorecontractsAPI.BlocklogGetLatestBlockInfo(context.Background()).Execute()
	require.NoError(t, err)

	require.Equal(t, block.BlockIndex+1, newBlock.BlockIndex, "block index should be incremented")

	chainObjId, err := iotago.ObjectIDFromHex(chain.ChainID.String())
	require.NoError(t, err)

	object, err := clu.L1Client().GetObject(context.Background(), iotaclient.GetObjectRequest{
		ObjectID: chainObjId,
		Options: &iotajsonrpc.IotaObjectDataOptions{
			ShowContent: true,
		},
	})
	require.NoError(t, err)

	var fieldMap map[string]interface{}
	err = json.Unmarshal(object.Data.Content.Data.MoveObject.Fields, &fieldMap)
	require.NoError(t, err)

	require.Equal(t, int(fieldMap["state_index"].(float64)), int(newBlock.BlockIndex), "state index in anchor should equal to state index in storage")
}
