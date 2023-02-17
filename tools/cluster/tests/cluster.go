package tests

import (
	"flag"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/util/l1starter"
	"github.com/iotaledger/wasp/tools/cluster"
	"github.com/iotaledger/wasp/tools/cluster/templates"
)

type waspClusterOpts struct {
	nNodes       int
	modifyConfig templates.ModifyNodesConfigFn
	dirName      string
}

// by default, when running the cluster tests we will automatically setup a private tangle,
// however its possible to run the tests on any compatible network, by providing the L1 node configuration:
// example:
// go test -timeout 30m github.com/iotaledger/wasp/tools/cluster/tests -layer1-api="http://1.1.1.123:3000" -layer1-faucet="http://1.1.1.123:5000"
var l1 = l1starter.New(flag.CommandLine, flag.CommandLine)

// newCluster starts a new cluster environment for tests.
// It is a private function because cluster tests cannot be run in parallel,
// so all cluster tests MUST be in this same package.
func newCluster(t *testing.T, opt ...waspClusterOpts) *cluster.Cluster {
	if testing.Short() {
		t.Skip("Skipping cluster test in short mode")
	}
	l1.StartPrivtangleIfNecessary(t.Logf)

	dirname := "wasp-cluster"
	var modifyNodesConfig templates.ModifyNodesConfigFn

	waspConfig := cluster.DefaultWaspConfig()

	if len(opt) > 0 {
		if opt[0].dirName != "" {
			dirname = opt[0].dirName
		}
		if opt[0].nNodes != 0 {
			waspConfig.NumNodes = opt[0].nNodes
		}
		modifyNodesConfig = opt[0].modifyConfig
	}

	clusterConfig := cluster.NewConfig(
		waspConfig,
		l1.Config,
		modifyNodesConfig,
	)

	dataPath := path.Join(os.TempDir(), dirname)
	clu := cluster.New(t.Name(), clusterConfig, dataPath, t, nil)

	err := clu.InitDataPath(".", true)
	require.NoError(t, err)

	err = clu.StartAndTrustAll(dataPath)
	require.NoError(t, err)

	t.Cleanup(clu.Stop)

	return clu
}
