package tests

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
	"github.com/iotaledger/wasp/tools/cluster/templates"
	"github.com/stretchr/testify/require"
)

func TestWaspCLIExternalRotation(t *testing.T) {
	// this test starts a chain on cluster of 4 nodes,
	// adds 1 new node as an access node (this node will be part of the new committee, this way it is synced)
	// then puts the chain on maintenance mode, stops the cluster
	// starts a new 4 nodes cluster (including the previous access node), runs the DKG on the new nodes,
	// rotates the chain state controller to the new cluster
	// stops the maintenance and ensure the chain is up-and-running

	w := newWaspCLITest(t, waspClusterOpts{
		nNodes:  4,
		dirName: "wasp-cluster-initial",
	})

	inccounterSCName := "inccounter"
	checkCounter := func(wTest *WaspCLITest, n int) {
		// test chain call-view command
		out := wTest.Run("chain", "call-view", inccounterSCName, "getCounter")
		out = wTest.Pipe(out, "decode", "string", "counter", "int")
		require.Regexp(t, fmt.Sprintf(`(?m)counter:\s+%d$`, n), out[0])
	}

	committee, quorum := w.CommitteeConfig()
	out := w.Run(
		"chain",
		"deploy",
		"--chain=chain1",
		committee,
		quorum,
		fmt.Sprintf("--gov-controller=%s", w.WaspCliAddress.Bech32(parameters.L1.Protocol.Bech32HRP)),
	)
	chainID := regexp.MustCompile(`(.*)ChainID:\s*([a-zA-Z0-9_]*),`).FindStringSubmatch(out[len(out)-1])[2]

	// start a new wasp cluster
	w2 := newWaspCLITest(t, waspClusterOpts{
		nNodes:  4,
		dirName: "wasp-cluster-new-gov",
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

	// adds node #0 from cluster2 as access node of the chain
	{
		// TODO I guess it would make sense to do these steps via CLI
		node0peerInfo, err := w2.Cluster.WaspClient(0).GetPeeringSelf()
		require.NoError(t, err)

		/// TODO HMM, should nodes be added as trusted peers first, or access nodes in gov contract?
		//- It really shouldnt' matter..

		// set trust relations between node0 of cluster 2 and all nodes of cluster 1
		w.Cluster.AddTrustedNode(node0peerInfo)
		for _, nodeIndex := range w.Cluster.Config.AllNodes() {
			peerInfo, err := w.Cluster.WaspClient(nodeIndex).GetPeeringSelf()
			require.NoError(t, err)
			w2.Cluster.AddTrustedNode(peerInfo, []int{0})
		}

		// add node 0 from cluster 2 as an access node in the governance contract
		pubKey, err := cryptolib.NewPublicKeyFromBase58String(node0peerInfo.PubKey)
		require.NoError(t, err)

		out = w.Run("chain", "change-access-nodes", "accept", pubKey.String())
		out = w.GetReceiptFromRunPostRequestOutput(out)
		require.Regexp(t, `.*Error: \(empty\).*`, strings.Join(out, ""))
	}

	// activate the chain on the new nodes
	w2.Run("chain", "add", "chain1", chainID)
	w2.Run("set", "chain", "chain1")
	w2.Run("chain", "activate")

	// deploy a contract, test its working
	{
		vmtype := vmtypes.WasmTime
		w.CopyFile(srcFile)

		// test chain deploy-contract command
		w.Run("chain", "deploy-contract", vmtype, inccounterSCName, "inccounter SC", file,
			"string", "counter", "int64", "42",
		)

		checkCounter(w, 42)
	}

	// init maintenance
	out = w.PostRequestGetReceipt("governance", "startMaintenance")
	require.Regexp(t, `.*Error: \(empty\).*`, strings.Join(out, ""))

	// stop the initial cluster
	// TODO is it needed to check that node0 from clust2 is synced at this point?
	w.Cluster.Stop()

	// for the approach below to work, we would need "permitionless access nodes",
	// instead we need to add the node to the access nodes list in the gov contract as it is done above
	// // keep add a node from the old cluster as a peer of the new cluster, so that the new nodes can sync the chain state
	// // we're chosing 5, so that the default ports don't conflict with the new cluster
	// initialClusterNodeIndex := 5
	// {
	// 	initOk := make(chan bool, 1)
	// 	apiURL := fmt.Sprintf("http://localhost:%s", strconv.Itoa(w.Cluster.Config.APIPort(initialClusterNodeIndex)))
	// 	_, err := cluster.DoStartWaspNode(
	// 		w.Cluster.NodeDataPath(initialClusterNodeIndex),
	// 		initialClusterNodeIndex,
	// 		apiURL,
	// 		initOk,
	// 		t,
	// 	)
	// 	require.NoError(t, err)
	// 	select {
	// 	case <-initOk:
	// 	case <-time.After(5 * time.Second):
	// 		t.Fatal("timeout re-starting node from initial cluster")
	// 	}
	// }
	//
	// // establish trust connections between the node from the old cluster and the new cluster
	// {
	// 	nodeFromInitialClusterPeerInfo, err := w.Cluster.WaspClient(initialClusterNodeIndex).GetPeeringSelf()
	// 	require.NoError(t, err)
	// 	w2.Cluster.AddTrustedNode(nodeFromInitialClusterPeerInfo)
	// 	for _, nodeIndex := range w2.Cluster.Config.AllNodes() {
	// 		peerInfo, err := w2.Cluster.WaspClient(nodeIndex).GetPeeringSelf()
	// 		require.NoError(t, err)
	// 		w.Cluster.AddTrustedNode(peerInfo, []int{initialClusterNodeIndex})
	// 	}
	// }

	// run DKG on the new cluster, obtain the new state controller address
	out = w2.Run("chain", "rundkg")
	newStateControllerAddr := regexp.MustCompile(`(.*):\s*([a-zA-Z0-9_]*)$`).FindStringSubmatch(out[0])[2]

	// issue a governance rotatation via CLI
	out = w.Run("chain", "rotate", newStateControllerAddr)
	require.Regexp(t, `.*successfully.*`, strings.Join(out, ""))

	// stop maintenance
	// set the new nodes as the default (so querying the receipt doesn't fail)
	w.Run("set", "wasp.0.api", w2.Cluster.Config.APIHost(0))
	out = w.PostRequestGetReceipt("governance", "stopMaintenance")
	require.Regexp(t, `.*Error: \(empty\).*`, strings.Join(out, ""))

	// chain still works
	w2.Run("chain", "post-request", inccounterSCName, "increment")
	checkCounter(w2, 43)
}
