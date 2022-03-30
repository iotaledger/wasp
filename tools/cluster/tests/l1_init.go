package tests

import (
	"context"
	"flag"
	"os"
	"path"

	"github.com/iotaledger/wasp/packages/nodeconn"
	"github.com/iotaledger/wasp/packages/testutil/privtangle"
)

// by default, when running the cluster tests we will automatically setup a private tangle,
// however its possible to run the tests on any compatible network, by providing the L1 node configuration:
// example:
// go test -timeout 30m github.com/iotaledger/wasp/tools/cluster/tests -layer1-host="1.1.1.123" -layer1-api-port=4000 -layer1-faucet-port=5000

var (
	layer1Host       = flag.String("layer1-host", "", "layer1 host")
	layer1APIPort    = flag.Int("layer1-api-port", 0, "layer1 API port")
	layer1FaucetPort = flag.Int("layer1-faucet-port", 0, "layer1 faucet port")
	pvtTangleNnodes  = flag.Int("priv-tangle-n-nodes", 2, "number of hornet nodes to be spawned in the private tangle")
)
var ClustL1Config nodeconn.L1Config

const pvtTangleAPIPort = 16500

// init sets up a private tangle to run the cluster tests, in case no L1 host was provided via cli
func init() {
	if *layer1Host != "" {
		ClustL1Config.Hostname = *layer1Host
		ClustL1Config.APIPort = *layer1APIPort
		ClustL1Config.FaucetPort = *layer1FaucetPort
		return
	}

	// start private tangle if no L1 parameters were provided
	l1DirPath := path.Join(os.TempDir(), "l1")
	ctx := context.Background()
	pt := privtangle.Start(ctx, l1DirPath, pvtTangleAPIPort, *pvtTangleNnodes, nil)
	ClustL1Config.Hostname = "localhost"
	ClustL1Config.APIPort = pt.NodePortRestAPI(0)
	ClustL1Config.FaucetPort = pt.NodePortFaucet(0)
	ClustL1Config.FaucetKey = pt.FaucetKeyPair
}
