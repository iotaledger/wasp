package tests

import (
	"flag"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/clients/iota-go/iotaconn"
	"github.com/iotaledger/wasp/packages/testutil/l1starter"
	"github.com/iotaledger/wasp/tools/cluster"
)

type waspClusterOpts struct {
	nNodes       int
	modifyConfig cluster.ModifyNodesConfigFn
	dirName      string
}

var l1 l1starter.IotaNodeEndpoint

// by default, when running the cluster tests we will automatically setup a private tangle,
// however it's possible to run the tests on any compatible network, by providing the L1 node configuration.
// example:
// go test -timeout 30m github.com/iotaledger/wasp/tools/cluster/tests -layer1-api="http://1.1.1.123:3000" -layer1-faucet="http://1.1.1.123:5000"

func parseConfig() l1starter.L1EndpointConfig {
	config := l1starter.L1EndpointConfig{}

	args := flag.CommandLine
	args.StringVar(&config.APIURL, "layer1-api", "", "layer1 API address")
	args.StringVar(&config.FaucetURL, "layer1-faucet", "", "layer1 faucet port")

	if len(config.FaucetURL) > 0 || len(config.APIURL) > 0 {
		config.IsLocal = false
	}

	return config
}

// newCluster starts a new cluster environment (both L1 and L2) for tests.
// It is a private function because cluster tests cannot be run in parallel,
// so all cluster tests MUST be in this same package.
func newCluster(t *testing.T, opt ...waspClusterOpts) *cluster.Cluster {
	if testing.Short() {
		t.Skip("Skipping cluster test in short mode")
	}

	dirname := "wasp-cluster"
	var modifyNodesConfig cluster.ModifyNodesConfigFn

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

	l1 = l1starter.ClusterStart(l1starter.L1EndpointConfig{
		IsLocal:       false,
		RandomizeSeed: true,
		APIURL:        iotaconn.AlphanetEndpointURL,
		FaucetURL:     iotaconn.AlphanetFaucetURL,
	})

	clusterConfig := cluster.NewConfig(
		waspConfig,
		l1,
		modifyNodesConfig,
	)
	l1PackageID := l1.ISCPackageID()
	dataPath := path.Join(os.TempDir(), dirname)
	clu := cluster.New(t.Name(), clusterConfig, dataPath, t, nil, &l1PackageID)

	err := clu.InitDataPath(".", true)
	require.NoError(t, err)

	err = clu.StartAndTrustAll(dataPath)
	require.NoError(t, err)

	t.Cleanup(clu.Stop)

	return clu
}
