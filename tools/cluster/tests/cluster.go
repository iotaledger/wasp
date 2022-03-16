package tests

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
	numNodes       = flag.Int("num-nodes", 4, "amount of wasp nodes")
	layer1Hostname = flag.String("layer1-hostname", defaultConfig.L1.Hostname, "layer1 hostname")
	layer1APIPort  = flag.Int("layer1-api-port", defaultConfig.L1.APIPort, "layer1 API port")
)

// newCluster starts a new cluster environment for tests.
// It is a private function because cluster tests cannot be run in parallel,
// so all cluster tests MUST be in this same package.
// opt: [n nodes, custom cluster config, modifyNodesConfigFn]
func newCluster(t *testing.T, opt ...interface{}) *cluster.Cluster {
	if testing.Short() {
		t.Skip("Skipping cluster test in short mode")
	}

	config := cluster.DefaultConfig()

	config.L1.Hostname = *layer1Hostname
	config.L1.APIPort = *layer1APIPort

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

	clu := cluster.New(t.Name(), config, t)

	dataPath := path.Join(os.TempDir(), "wasp-cluster")
	err := clu.InitDataPath(".", dataPath, true, modifyNodesConfig)
	require.NoError(t, err)

	err = clu.Start(dataPath)
	require.NoError(t, err)

	t.Cleanup(clu.Stop)

	return clu
}
