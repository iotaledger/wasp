package tests

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/clients/apiclient"
	"github.com/iotaledger/wasp/clients/chainclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/testcore/contracts/inccounter"
	"github.com/iotaledger/wasp/tools/cluster/templates"
)

// executed in cluster_test.go
func testPermissionlessAccessNode(t *testing.T, env *ChainEnv) {
	// deposit funds for offledger requests
	keyPair, _, err := env.Clu.NewKeyPairWithFunds()
	require.NoError(t, err)

	env.DepositFunds(iotaclient.DefaultGasBudget, keyPair)

	// spin a new node
	clu2 := newCluster(t, waspClusterOpts{
		nNodes:  1,
		dirName: "wasp-cluster-access-node",
		modifyConfig: func(nodeIndex int, configParams templates.WaspConfigParams) templates.WaspConfigParams {
			// avoid port conflicts when running everything on localhost
			configParams.APIPort += 100
			configParams.MetricsPort += 100
			configParams.PeeringPort += 100
			configParams.ProfilingPort += 100
			return configParams
		},
	})
	// remove this cluster when the test ends
	t.Cleanup(clu2.Stop)

	nodeClient := env.Clu.WaspClient(0)
	accessNodeClient := clu2.WaspClient(0)

	// adds node #0 from cluster2 as access node of node #0 from cluster1

	// trust setup between the two nodes
	node0peerInfo, _, err := nodeClient.NodeAPI.GetPeeringIdentity(context.Background()).Execute()
	require.NoError(t, err)

	err = clu2.AddTrustedNode(apiclient.PeeringTrustRequest{
		Name:       "node-from-other-cluster",
		PublicKey:  node0peerInfo.PublicKey,
		PeeringURL: node0peerInfo.PeeringURL,
	})
	require.NoError(t, err)

	accessNodePeerInfo, _, err := accessNodeClient.NodeAPI.GetPeeringIdentity(context.Background()).Execute()
	require.NoError(t, err)

	err = env.Clu.AddTrustedNode(apiclient.PeeringTrustRequest{
		Name:       "node-from-other-cluster",
		PublicKey:  accessNodePeerInfo.PublicKey,
		PeeringURL: accessNodePeerInfo.PeeringURL,
	}, []int{0})
	require.NoError(t, err)

	// activate the chain on the access node
	_, err = accessNodeClient.ChainsAPI.
		SetChainRecord(context.Background(), env.Chain.ChainID.String()).
		ChainRecord(apiclient.ChainRecord{
			IsActive:    true,
			AccessNodes: []string{},
		}).Execute()
	require.NoError(t, err)

	// add node 0 from cluster 2 as a *permissionless* access node
	_, err = nodeClient.ChainsAPI.AddAccessNode(context.Background(), env.Chain.ChainID.String(), accessNodePeerInfo.PublicKey).Execute()
	require.NoError(t, err)

	// give some time for the access node to sync
	time.Sleep(2 * time.Second)

	// send a request to the access node
	myClient := chainclient.New(
		env.Clu.L1Client(),
		accessNodeClient,
		env.Chain.ChainID,
		env.Clu.Config.ISCPackageID(),
		keyPair,
	)
	req, err := myClient.PostOffLedgerRequest(context.Background(), accounts.FuncWithdraw.Message(),
		chainclient.PostRequestParams{
			Allowance: isc.NewAssets(10),
		},
	)
	require.NoError(t, err)

	// request has been processed
	_, err = env.Chain.CommitteeMultiClient().WaitUntilRequestProcessedSuccessfully(context.Background(), env.Chain.ChainID, req.ID(), false, 1*time.Minute)
	require.NoError(t, err)

	// remove the access node from cluster1 node 0
	_, err = nodeClient.ChainsAPI.RemoveAccessNode(context.Background(), env.Chain.ChainID.String(), accessNodePeerInfo.PublicKey).Execute()
	require.NoError(t, err)

	time.Sleep(1 * time.Second) // Access/Server node info is exchanged asynchronously.

	// try sending the request again
	req, err = myClient.PostOffLedgerRequest(context.Background(), inccounter.FuncIncCounter.Message(nil))
	require.NoError(t, err)

	// request is not processed after a while
	time.Sleep(2 * time.Second)
	receipt, _, err := nodeClient.ChainsAPI.GetReceipt(context.Background(), env.Chain.ChainID.String(), req.ID().String()).Execute()

	require.Error(t, err)
	require.Regexp(t, `404`, err.Error())
	require.Nil(t, receipt)
}
