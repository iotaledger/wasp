package tests

import (
	"context"
	"flag"
	"os"
	"path"

	"github.com/iotaledger/wasp/packages/testutil/privtangle"
	"github.com/iotaledger/wasp/tools/cluster"
)

var (
	layer1Hostname   = flag.String("layer1-hostname", "", "layer1 hostname")
	layer1APIPort    = flag.Int("layer1-api-port", 0, "layer1 API port")
	layer1FaucetPort = flag.Int("layer1-faucet-port", 0, "layer1 faucet port")
	pvtTangleNnodes  = flag.Int("priv-tangle-n-nodes", 2, "number of hornet nodes to be spawned in the private tangle")
)
var L1Config cluster.L1Config

// init sets up a private tangle to run the cluster tests, in case no L1 host was provided via cli
func init() {
	if *layer1Hostname != "" {
		L1Config.Hostname = *layer1Hostname
		L1Config.APIPort = *layer1APIPort
		L1Config.FaucetPort = *layer1FaucetPort
		return
	}

	// start private tangle if no L1 parameters were provided
	l1DirPath := path.Join(os.TempDir(), "l1")
	ctx := context.Background()
	pt := privtangle.Start(ctx, l1DirPath, L1Config.APIPort, *pvtTangleNnodes, nil)
	L1Config.Hostname = "localhost"
	L1Config.APIPort = pt.NodePortRestAPI(0)
	L1Config.FaucetPort = pt.NodePortFaucet(0)
}
