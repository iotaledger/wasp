package tests

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/clients/apiclient"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/vm/core/governance"
	"github.com/iotaledger/wasp/v2/tools/cluster"
)

func TestWaspCLIExternalRotationGovAccessNodes(t *testing.T) {
	t.Skip("TODO: fix or remove test")
	addAccessNode := func(w *WaspCLITest, pubKey string) {
		out := w.MustRun("chain", "gov-change-access-nodes", "accept", pubKey, "--node=0")
		out = w.GetReceiptFromRunPostRequestOutput(out)
		require.Regexp(t, `.*Error: \(empty\).*`, strings.Join(out, ""))
	}
	testWaspCLIExternalRotation(t, addAccessNode)
}

func TestWaspCLIExternalRotationPermissionlessAccessNodes(t *testing.T) {
	t.Skip("TODO: fix or remove test")

	addAccessNode := func(w *WaspCLITest, pubKey string) {
		for _, idx := range w.Cluster.AllNodes() {
			w.MustRun("chain", "access-nodes", "add", "--peers=next-committee-member", fmt.Sprintf("--node=%d", idx))
		}
	}
	testWaspCLIExternalRotation(t, addAccessNode)
}

func testWaspCLIExternalRotation(t *testing.T, addAccessNode func(*WaspCLITest, string)) {
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
		out := wTest.MustRun("chain", "call-view", inccounterSCName, "getCounter", "--node=0")
		out = wTest.MustPipe(out, "decode", "string", "counter", "int")
		require.Regexp(t, fmt.Sprintf(`(?m)counter:\s+%d$`, n), out[0])
	}

	committee, quorum := w.ArgCommitteeConfig(0)
	out := w.MustRun(
		"chain",
		"deploy",
		"--chain=chain1",
		committee,
		quorum,
		fmt.Sprintf("--gov-controller=%s", w.WaspCliAddress.String()),
		"--node=0",
	)
	// TODO: FIX this test
	//matches := regexp.MustCompile(`.*ChainID:\s*([a-zA-Z0-9_]*)\s+.*`).FindStringSubmatch(strings.Join(out, " "))
	//chainID := matches[1]
	w.ActivateChainOnAllNodes("chain1", 0)

	// start a new wasp cluster
	w2 := newWaspCLITest(t, waspClusterOpts{
		nNodes:  4,
		dirName: "wasp-cluster-new-gov",
		modifyConfig: func(nodeIndex int, configParams cluster.WaspConfigParams) cluster.WaspConfigParams {
			// avoid port conflicts when running everything on localhost
			configParams.APIPort += 100
			configParams.MetricsPort += 100
			configParams.PeeringPort += 100
			configParams.ProfilingPort += 100
			return configParams
		},
	})

	// adds node #0 from cluster2 as access node of the chain
	{
		node0peerInfo, _, err := w2.Cluster.WaspClient(0).NodeAPI.
			GetPeeringIdentity(context.Background()).
			Execute()
		require.NoError(t, err)

		// set trust relations between node0 of cluster 2 and all nodes of cluster 1
		err = w.Cluster.AddTrustedNode(apiclient.PeeringTrustRequest{
			Name:       "next-committee-member",
			PublicKey:  node0peerInfo.PublicKey,
			PeeringURL: node0peerInfo.PeeringURL,
		})
		require.NoError(t, err)

		for _, nodeIndex := range w.Cluster.Config.AllNodes() {
			// equivalent of "wasp-cli peer info"
			peerInfo, _, err2 := w.Cluster.WaspClient(nodeIndex).NodeAPI.
				GetPeeringIdentity(context.Background()).
				Execute()
			require.NoError(t, err2)

			w2.MustRun("peering", "trust", fmt.Sprintf("old-committee-%d", nodeIndex), peerInfo.PublicKey, peerInfo.PeeringURL, "--node=0")
			require.NoError(t, err2)
		}

		// add node 0 from cluster 2 as an access node
		pubKey, err := cryptolib.PublicKeyFromString(node0peerInfo.PublicKey)
		require.NoError(t, err)

		addAccessNode(w, pubKey.String())
	}

	// activate the chain on the new nodes
	w2.MustRun("chain", "add", "chain1")
	for _, idx := range w2.Cluster.AllNodes() {
		w2.MustRun("chain", "activate", fmt.Sprintf("--node=%d", idx))
	}

	// deploy a contract, test its working
	{
		// TODO: rewrite to use a native contract?
		/*
				// deposit funds first to run offledger requests
				w.MustRun("chain", "deposit", "base:10000000", "--node=0")

				// test chain deploy-contract command
				w.MustRun("chain", "deploy-contract", vmtype, inccounterSCName, "inccounter SC", file,
				"string", "counter", "int64", "42",
				"--node=0",
			)
		*/

		checkCounter(w, 42)
	}

	// init maintenance
	out = w.PostRequestGetReceipt("governance", "startMaintenance", "--node=0")
	require.Regexp(t, `.*Error: \(empty\).*`, strings.Join(out, ""))

	// check that node0 from clust2 is synced and maintenance is on
	for i := 0; ; i++ {
		if i >= 30 {
			t.Fatalf("Timeout waiting access node to be synched, last out=%v", out)
		}
		time.Sleep(1 * time.Second)
		var err error
		out, err = w2.Run("chain", "call-view", governance.Contract.Name, governance.ViewGetMaintenanceStatus.Name, "--node=0")
		if err != nil {
			t.Logf("Warning: call failed to ViewGetMaintenanceStatus: %v", err)
			continue
		}

		// refactor me: Waspcli maintenance status
		/*
			out, err = w2.Pipe(out, "decode", "string", governance.ParamMaintenanceStatus, "bool")
			if err != nil {
				t.Logf("Warning: call failed to ViewGetMaintenanceStatus: %v", err)
				continue
			}
			if strings.Contains(out[0], "true") {
				break
			}*/
	}

	// stop the initial cluster
	w.Cluster.Stop()

	// run DKG on the new cluster, obtain the new state controller address
	out = w2.MustRun("chain", "rundkg", w2.ArgAllNodesExcept(0), "--node=0")
	var newStateControllerAddr string
	for _, line := range out {
		matches := regexp.MustCompile(`Address: ([a-zA-Z0-9_]+)`).FindStringSubmatch(line)
		if len(matches) > 1 {
			newStateControllerAddr = matches[1]
		}
	}
	t.Logf("DKG generated state controller address: %v", newStateControllerAddr)

	// issue a governance rotatation via CLI
	out = w.MustRun("chain", "rotate", newStateControllerAddr)
	require.Regexp(t, `.*successfully.*`, strings.Join(out, ""))

	// stop maintenance
	// set the new nodes as the default (so querying the receipt doesn't fail)
	w.MustRun("set", "wasp.0", w2.Cluster.Config.APIHost(0))
	out = w.PostRequestGetReceipt("governance", "stopMaintenance", "--node=0")
	require.Regexp(t, `.*Error: \(empty\).*`, strings.Join(out, ""))

	// chain still works
	w2.MustRun("chain", "post-request", "-s", inccounterSCName, "increment", "--node=0")
	checkCounter(w2, 43)
}

func TestRotateOnOrigin(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping cluster tests in short mode")
	}

	w := newWaspCLITest(t, waspClusterOpts{
		nNodes: 1,
	})
	// start a chain on node 0
	w.MustRun("chain", "deploy", "--chain=chain1", "--node=0")
	w.ActivateChainOnAllNodes("chain1", 0)
	dkg := w.MustRun("chain", "rundkg", "--peers", "me")
	require.Greater(t, len(dkg), 1)
	rotateAddress := strings.TrimPrefix(dkg[1], "Address: ")

	blockIndex1 := getBlockIndex(t, w)

	w.MustRun("chain", "rotate", rotateAddress, "--chain=chain1")

	blockIndex2 := getBlockIndex(t, w)

	// block index should be the same until deposit is called
	require.Equal(t, blockIndex2, blockIndex1)

	w.MustRun("chain", "deposit", "base|1000000", "--chain=chain1")

	blockIndex3 := getBlockIndex(t, w)
	require.Equal(t, blockIndex2+1, blockIndex3)

	chainInfo := w.MustRun("chain", "info", "--chain=chain1")
	stateAddress := extractStateAddress(chainInfo)
	require.Equal(t, rotateAddress, stateAddress)
}

func getBlockIndex(t *testing.T, w *WaspCLITest) uint32 {
	block, _, err := w.Cluster.WaspClient(0).CorecontractsAPI.BlocklogGetLatestBlockInfo(context.Background()).Execute()
	require.NoError(t, err)
	return block.BlockIndex
}

var stateAddressRegex = regexp.MustCompile(`State address: (0x[0-9a-fA-F]+)`)

func extractStateAddress(lines []string) string {
	for _, line := range lines {
		matches := stateAddressRegex.FindStringSubmatch(line)
		if len(matches) > 1 {
			return matches[1]
		}
	}
	return ""
}
