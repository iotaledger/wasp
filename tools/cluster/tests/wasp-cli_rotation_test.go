package tests

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
	"github.com/iotaledger/wasp/tools/cluster"
	"github.com/stretchr/testify/require"
)

func TestWaspCLIExternalRotation(t *testing.T) {
	// this test starts a chain on acluster of 6 nodes,
	// then puts the chain on maintenance mode, stops the cluster
	// starts a new 4 nodes cluster, runs the DKG on the new nodes,
	// adds 1 of the previous cluster nodes as peer, so that they can sync state
	// rotates the chain state controller to the new cluster
	// stops the maintenance and ensure the chain is up-and-running

	w := newWaspCLITest(t, waspClusterOpts{
		nNodes:  6,
		dirName: "wasp-cluster-initial",
	})

	committee, quorum := w.CommitteeConfig()
	out := w.Run(
		"chain",
		"deploy",
		"--chain=chain1",
		committee,
		quorum,
		fmt.Sprintf("--gov-controller=%s", w.CliAddress.Bech32(parameters.L1.Protocol.Bech32HRP)),
	)
	chainID := regexp.MustCompile(`(.*)ChainID:\s*([a-zA-Z0-9_]*),`).FindStringSubmatch(out[len(out)-1])[2]

	vmtype := vmtypes.WasmTime
	name := "inccounter"
	w.CopyFile(srcFile)

	// test chain deploy-contract command
	w.Run("chain", "deploy-contract", vmtype, name, "inccounter SC", file,
		"string", "counter", "int64", "42",
	)

	checkCounter := func(wTest *WaspCLITest, n int) {
		// test chain call-view command
		out = wTest.Run("chain", "call-view", name, "getCounter")
		out = wTest.Pipe(out, "decode", "string", "counter", "int")
		require.Regexp(t, fmt.Sprintf(`(?m)counter:\s+%d$`, n), out[0])
	}

	checkCounter(w, 42)

	// init maintenance
	out = w.PostRequestGetReceipt("governance", "startMaintenance")
	require.Regexp(t, `.*Error: \(empty\).*`, strings.Join(out, ""))

	// stop the initial cluster
	w.Cluster.Stop()

	// start a new wasp cluster
	w2 := newWaspCLITest(t, waspClusterOpts{
		dirName: "wasp-cluster-new-gov",
		nNodes:  4,
	})

	// keep add a node from the old cluster as a peer of the new cluster, so that the new nodes can sync the chain state
	// we're chosing 5, so that the default ports don't conflict with the new cluster
	initialClusterNodeIndex := 5
	{
		initOk := make(chan bool, 1)
		apiURL := fmt.Sprintf("http://localhost:%s", strconv.Itoa(w.Cluster.Config.APIPort(initialClusterNodeIndex)))
		_, err := cluster.DoStartWaspNode(
			w.Cluster.NodeDataPath(initialClusterNodeIndex),
			initialClusterNodeIndex,
			apiURL,
			initOk,
			t,
		)
		require.NoError(t, err)
		select {
		case <-initOk:
		case <-time.After(5 * time.Second):
			t.Fatal("timeout re-starting node from initial cluster")
		}
	}

	// establish trust connections between the node from the old cluster and the new cluster
	{
		nodeFromInitialClusterPeerInfo, err := w.Cluster.WaspClient(initialClusterNodeIndex).GetPeeringSelf()
		require.NoError(t, err)
		w2.Cluster.AddTrustedNode(nodeFromInitialClusterPeerInfo)
		for _, nodeIndex := range w2.Cluster.Config.AllNodes() {
			peerInfo, err := w2.Cluster.WaspClient(nodeIndex).GetPeeringSelf()
			require.NoError(t, err)
			w.Cluster.AddTrustedNode(peerInfo, []int{initialClusterNodeIndex})
		}
	}

	// run DKG on the new cluster, obtain the new state controller address
	out = w2.Run("chain", "rundkg", "--committee=0,1,2,3")
	newStateControllerAddr := regexp.MustCompile(`(.*):\s*([a-zA-Z0-9_]*)$`).FindStringSubmatch(out[0])[2]

	w2.Run("chain", "add", "chain1", chainID)
	w2.Run("set", "chain", "chain1")

	// issue a governance rotatation via CLI
	out = w.Run("chain", "rotate", newStateControllerAddr)
	require.Regexp(t, `.*successfully.*`, strings.Join(out, ""))

	// activate the chain on the new nodes
	w2.Run("chain", "activate", "--nodes=0,1,2,3")

	// stop maintenance
	out = w2.PostRequestGetReceipt("governance", "stopMaintenance")
	require.Regexp(t, `.*Error: \(empty\).*`, strings.Join(out, ""))

	// chain still works
	w2.Run("chain", "post-request", name, "increment")
	checkCounter(w2, 43)
}
