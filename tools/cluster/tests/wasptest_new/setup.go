package wasptest

import (
	"flag"
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/plugins/wasmtimevm"
	"github.com/iotaledger/wasp/tools/cluster"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path"
	"runtime"
	"testing"
	"time"
)

var (
	useGo       = flag.Bool("go", false, "use Go instead of Rust")
	useWasp     = flag.Bool("wasp", false, "use Wasp built-in instead of Rust")
	wasmLoaded  = false
	seed        = "C6hPhCS2E2dKUGS3qj4264itKXohwgL3Lm2fNxayAKr"
	wallet      = testutil.NewWallet(seed)
	scOwner     = wallet.WithIndex(0)
	programHash hashing.HashValue
)

func check(err error, t *testing.T) {
	t.Helper()
	assert.NoError(t, err)
	if err != nil {
		t.FailNow()
	}
}

func setup(t *testing.T, configPath string) *cluster.Cluster {
	_, filename, _, _ := runtime.Caller(0)

	wasps, err := cluster.New(path.Join(path.Dir(filename), "..", configPath), "cluster-data")
	check(err, t)

	err = wasps.Init(true, t.Name())
	check(err, t)

	err = wasps.Start()
	check(err, t)

	t.Cleanup(wasps.Stop)

	return wasps
}

func loadWasmIntoWasps(chain *cluster.Chain, wasmName string, scDescription string, initParams map[string]interface{}) error {
	wasmLoaded = true
	wasmPath := wasmName + "_bg.wasm"
	if *useGo {
		fmt.Println("Using Go Wasm instead of Rust Wasm")
		time.Sleep(time.Second)
		wasmPath = wasmName + "_go.wasm"
	}
	wasm, err := ioutil.ReadFile("../wasmtest/wasm/" + wasmPath)
	if err != nil {
		return err
	}
	_, err = chain.DeployExternalContract(wasmtimevm.PluginName, wasmName, scDescription, wasm, initParams)
	return err
}

func startSmartContract(wasps *cluster.Cluster, scProgramHash string, scDescription string) (*coretypes.ChainID, *address.Address, *balance.Color, error) {
	var err error
	if *useWasp || !wasmLoaded {
		fmt.Println("Using Wasp built-in instead of Rust Wasm")
		time.Sleep(time.Second)
		programHash, err = hashing.HashValueFromBase58(scProgramHash)
		if err != nil {
			return nil, nil, nil, err
		}
	} else {
		scProgramHash = programHash.String()
	}
	scChain, scAddr, scColor, err := apilib.DeployChain(apilib.CreateChainParams{
		Node:                  wasps.NodeClient,
		CommitteeApiHosts:     wasps.ApiHosts(),
		CommitteePeeringHosts: wasps.PeeringHosts(),
		N:                     4,
		T:                     3,
		OwnerSigScheme:        scOwner.SigScheme(),
		Description:           scDescription,
		Textout:               os.Stdout,
		Prefix:                "[deploy " + scProgramHash + "]",
	})
	if err != nil {
		return nil, nil, nil, err
	}

	err = apilib.ActivateChain(apilib.ActivateChainParams{
		ChainID:           *scChain,
		ApiHosts:          wasps.ApiHosts(),
		WaitForCompletion: true,
		PublisherHosts:    wasps.PublisherHosts(),
		Timeout:           30 * time.Second,
	})
	return scChain, scAddr, scColor, err
}
