package tests

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/client/chainclient"
	"github.com/iotaledger/wasp/client/scclient"
	"github.com/iotaledger/wasp/contracts/native/inccounter"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/utxodb"
	"github.com/iotaledger/wasp/tools/cluster/templates"
)

func TestPermitionlessAccessNode(t *testing.T) {
	// setup a test with a single node
	env := setupNativeInccounterTest(t, 1, []int{0})

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

	nodeClient := env.Clu.WaspClient(0)
	accessNodeClient := clu2.WaspClient(0)

	// adds node #0 from cluster2 as access node of node #0 from cluster1

	// trust setup between the two nodes
	node0peerInfo, err := nodeClient.GetPeeringSelf()
	require.NoError(t, err)
	err = clu2.AddTrustedNode(node0peerInfo)
	require.NoError(t, err)

	accessNodePeerInfo, err := accessNodeClient.GetPeeringSelf()
	require.NoError(t, err)
	err = env.Clu.AddTrustedNode(accessNodePeerInfo, []int{0})
	require.NoError(t, err)

	// activate the chain on the access node
	err = accessNodeClient.PutChainRecord(registry.NewChainRecord(env.Chain.ChainID, true, []*cryptolib.PublicKey{}))
	require.NoError(t, err)

	// add node 0 from cluster 2 as a *permitionless* access node
	err = nodeClient.AddAccessNode(env.Chain.ChainID, accessNodePeerInfo.PubKey)
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
	err = nodeClient.RemoveAccessNode(env.Chain.ChainID, accessNodePeerInfo.PubKey)
	require.NoError(t, err)

	// try sending the request again
	req, err = myClient.PostOffLedgerRequest(inccounter.FuncIncCounter.Name)
	require.NoError(t, err)

	// request is not processed after a while
	time.Sleep(2 * time.Second)
	rec, err := nodeClient.RequestReceipt(env.Chain.ChainID, req.ID())
	require.Error(t, err)
	require.Regexp(t, `"Code":404`, err.Error())
	require.Nil(t, rec)
}
