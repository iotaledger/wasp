package testutil

import (
	"flag"
	"os"
	"path"
	"testing"

	"github.com/iotaledger/wasp/tools/cluster"
	"github.com/stretchr/testify/require"
)

var defaultConfig = cluster.DefaultConfig()

var (
	numNodes          = flag.Int("num-nodes", 4, "amount of wasp nodes") //nolint:gomnd
	goShimmerUseNode  = flag.Bool("goshimmer-use-node", defaultConfig.Goshimmer.UseNode, "If false (default), a mocked version of Goshimmer will be used")
	goShimmerHostname = flag.String("goshimmer-hostname", defaultConfig.Goshimmer.Hostname, "Goshimmer hostname")
	goShimmerPort     = flag.Int("goshimmer-txport", defaultConfig.Goshimmer.TxStreamPort, "Goshimmer port")
)

// opt: [n nodes, custom cluster config, modifyNodesConfigFn]
func NewCluster(t *testing.T, opt ...interface{}) *cluster.Cluster {
	if testing.Short() {
		t.Skip("Skipping cluster test in short mode")
	}

	config := cluster.DefaultConfig()

	config.Goshimmer.Hostname = *goShimmerHostname
	config.Goshimmer.UseNode = *goShimmerUseNode
	config.Goshimmer.TxStreamPort = *goShimmerPort

	nNodes := *numNodes
	if len(opt) > 0 {
		n, ok := opt[0].(int)
		if ok {
			nNodes = n
		}
	}

	if len(opt) > 1 {
		customConfig, ok := opt[1].(*cluster.ClusterConfig)
		if ok {
			config = customConfig
		}
	}

	var modifyNodesConfig cluster.ModifyNodesConfigFn

	if len(opt) > 2 {
		fn, ok := opt[2].(cluster.ModifyNodesConfigFn)
		if ok {
			modifyNodesConfig = fn
		}
	}

	config.Wasp.NumNodes = nNodes

	clu := cluster.New(t.Name(), config)

	dataPath := path.Join(os.TempDir(), "wasp-cluster")
	err := clu.InitDataPath(".", dataPath, true, modifyNodesConfig)
	require.NoError(t, err)

	err = clu.Start(dataPath)
	require.NoError(t, err)

	t.Cleanup(clu.Stop)

	return clu
}
