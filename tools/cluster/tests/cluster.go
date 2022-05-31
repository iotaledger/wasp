package tests

import (
	"flag"
	"os"
	"path"
	"testing"

	"github.com/iotaledger/wasp/packages/util/l1starter"
	"github.com/iotaledger/wasp/tools/cluster"
	"github.com/stretchr/testify/require"
)

type waspClusterOpts struct {
	nNodes       int
	modifyConfig cluster.ModifyNodesConfigFn
}

// by default, when running the cluster tests we will automatically setup a private tangle,
// however its possible to run the tests on any compatible network, by providing the L1 node configuration:
// example:
// go test -timeout 30m github.com/iotaledger/wasp/tools/cluster/tests -layer1-host="1.1.1.123" -layer1-api-port=4000 -layer1-faucet-port=5000
var l1 = l1starter.New(flag.CommandLine)

// newCluster starts a new cluster environment for tests.
// It is a private function because cluster tests cannot be run in parallel,
// so all cluster tests MUST be in this same package.
func newCluster(t *testing.T, opt ...waspClusterOpts) *cluster.Cluster {
	if testing.Short() {
		t.Skip("Skipping cluster test in short mode")
	}

	dataPath := path.Join(os.TempDir(), "wasp-cluster")

	l1.StartPrivtangleIfNecessary(t.Logf)

	clusterConfig := cluster.NewConfig(
		cluster.DefaultWaspConfig(),
		l1.Config,
	)

	var modifyNodesConfig cluster.ModifyNodesConfigFn
	if len(opt) > 0 {
		clusterConfig.Wasp.NumNodes = opt[0].nNodes
		modifyNodesConfig = opt[0].modifyConfig
	}

	clu := cluster.New(t.Name(), clusterConfig, t)

	err := clu.InitDataPath(".", dataPath, true, modifyNodesConfig)
	require.NoError(t, err)

	err = clu.Start(dataPath)
	require.NoError(t, err)

	t.Cleanup(clu.Stop)

	return clu
}
