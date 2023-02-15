package tests

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/clients/apiclient"
	"github.com/iotaledger/wasp/clients/chainclient"
	"github.com/iotaledger/wasp/clients/scclient"
	"github.com/iotaledger/wasp/contracts/native/inccounter"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/utxodb"
	"github.com/iotaledger/wasp/tools/cluster/templates"
)

// executed in cluster_test.go
func testPermitionlessAccessNode(t *testing.T, env *ChainEnv) {
	// deploy the inccounter for the test to use
	env.deployNativeIncCounterSC(0)

	// deposit funds for offledger requests
	keyPair, _, err := env.Clu.NewKeyPairWithFunds()
	require.NoError(t, err)

	env.DepositFunds(utxodb.FundsFromFaucetAmount, keyPair)

	// spin a new node
	clu2 := newCluster(t, waspClusterOpts{
		nNodes:  1,
		dirName: "wasp-cluster-access-node",
		modifyConfig: func(nodeIndex int, configParams templates.WaspConfigParams) templates.WaspConfigParams {
			// avoid port conflicts when running everything on localhost
			configParams.APIPort += 100
			configParams.DashboardPort += 100
			configParams.MetricsPort += 100
			configParams.NanomsgPort += 100
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
	node0peerInfo, _, err := nodeClient.NodeApi.GetPeeringIdentity(context.Background()).Execute()
	require.NoError(t, err)

	err = clu2.AddTrustedNode(apiclient.PeeringTrustRequest{
		Name:       "node-from-other-cluster",
		PublicKey:  node0peerInfo.PublicKey,
		PeeringURL: node0peerInfo.PeeringURL,
	})
	require.NoError(t, err)

	accessNodePeerInfo, _, err := accessNodeClient.NodeApi.GetPeeringIdentity(context.Background()).Execute()
	require.NoError(t, err)

	err = env.Clu.AddTrustedNode(apiclient.PeeringTrustRequest{
		Name:       "node-from-other-cluster",
		PublicKey:  accessNodePeerInfo.PublicKey,
		PeeringURL: accessNodePeerInfo.PeeringURL,
	}, []int{0})
	require.NoError(t, err)

	// activate the chain on the access node
	_, err = accessNodeClient.ChainsApi.
		SetChainRecord(context.Background(), env.Chain.ChainID.String()).
		ChainRecord(apiclient.ChainRecord{
			IsActive:    true,
			AccessNodes: []string{},
		}).Execute()
	require.NoError(t, err)

	// add node 0 from cluster 2 as a *permitionless* access node
	_, err = nodeClient.ChainsApi.AddAccessNode(context.Background(), env.Chain.ChainID.String(), accessNodePeerInfo.PublicKey).Execute()
	require.NoError(t, err)

	// give some time for the access node to sync
	time.Sleep(2 * time.Second)

	// send a request to the access node
	myClient := scclient.New(
		chainclient.New(
			env.Clu.L1Client(),
			accessNodeClient,
			env.Chain.ChainID,
			keyPair,
		),
		isc.Hn(nativeIncCounterSCName),
	)
	req, err := myClient.PostOffLedgerRequest(inccounter.FuncIncCounter.Name)
	require.NoError(t, err)

	// request has been processed
	_, err = env.Chain.CommitteeMultiClient().WaitUntilRequestProcessedSuccessfully(env.Chain.ChainID, req.ID(), 1*time.Minute)
	require.NoError(t, err)

	// remove the access node from cluster1 node 0
	_, err = nodeClient.ChainsApi.RemoveAccessNode(context.Background(), env.Chain.ChainID.String(), accessNodePeerInfo.PublicKey).Execute()
	require.NoError(t, err)

	time.Sleep(1 * time.Second) // Access/Server node info is exchanged asynchronously.

	// try sending the request again
	req, err = myClient.PostOffLedgerRequest(inccounter.FuncIncCounter.Name)
	require.NoError(t, err)

	// request is not processed after a while
	time.Sleep(2 * time.Second)
	receipt, _, err := nodeClient.RequestsApi.GetReceipt(context.Background(), env.Chain.ChainID.String(), req.ID().String()).Execute()

	require.Error(t, err)
	require.Regexp(t, `404`, err.Error())
	require.Nil(t, receipt)
}
