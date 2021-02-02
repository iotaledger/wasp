package wasptest

import (
	"flag"
	"fmt"
	"github.com/iotaledger/wasp/packages/coretypes/cbalances"
	"github.com/iotaledger/wasp/packages/coretypes/requestargs"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/vm/wasmhost"
	"io/ioutil"
	"testing"
	"time"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/client/chainclient"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/tools/cluster"
	clutest "github.com/iotaledger/wasp/tools/cluster/testutil"
	"github.com/stretchr/testify/require"
)

var (
	useGo       = flag.Bool("go", false, "use Go instead of Rust")
	useWasp     = flag.Bool("wasp", false, "use Wasp built-in instead of Rust")
	seed        = "C6hPhCS2E2dKUGS3qj4264itKXohwgL3Lm2fNxayAKr"
	wallet      = testutil.NewWallet(seed)
	scOwner     = wallet.WithIndex(0)
	scOwnerAddr = scOwner.Address()
	chain       *cluster.Chain
	clu         *cluster.Cluster
	client      *chainclient.Client
	counter     *cluster.MessageCounter
	programHash hashing.HashValue
	err         error
)

// TODO detached example code
//var builtinProgramHash = map[string]string{
//	"donatewithfeedback": dwfimpl.ProgramHash,
//	"fairauction":        fairauction.ProgramHash,
//	"fairroulette":       fairroulette.ProgramHash,
//	"inccounter":         inccounter.ProgramHash,
//	"tokenregistry":      tokenregistry.ProgramHash,
//}

func check(err error, t *testing.T) {
	t.Helper()
	require.NoError(t, err)
}

func deployContract(wasmName string, scDescription string, initParams map[string]interface{}) error {
	wasmPath := wasmName + "_bg.wasm"
	if *useGo {
		wasmPath = wasmName + "_go.wasm"
	}

	if !*useWasp {
		wasm, err := ioutil.ReadFile(wasmhost.WasmPath(wasmPath))
		if err != nil {
			return err
		}
		_, ph, err := chain.DeployWasmContract(wasmName, scDescription, wasm, initParams)
		programHash = ph
		fmt.Printf("--- deployContract err = %v proghash = %s\n", err, programHash.String())
		return err
	}
	panic("example contract disabled")
	//fmt.Println("Using Wasp built-in SC instead of Rust Wasm SC")
	//time.Sleep(time.Second)
	//hash, ok := builtinProgramHash[wasmName]
	//if !ok {
	//	return errors.New("Unknown built-in SC: " + wasmName)
	//}

	// TODO detached example contract code
	//_, err := chain.DeployContract(wasmName, examples.VMType, hash, scDescription, initParams)
	//return err
	return nil
}

func postRequest(t *testing.T, contract coretypes.Hname, entryPoint coretypes.Hname, tokens int, params map[string]interface{}) {
	var transfer map[balance.Color]int64
	if tokens != 0 {
		transfer = map[balance.Color]int64{
			balance.ColorIOTA: int64(tokens),
		}
	}
	postRequestFull(t, contract, entryPoint, transfer, params)
}

func postRequestFull(t *testing.T, contract coretypes.Hname, entryPoint coretypes.Hname, transfer map[balance.Color]int64, params map[string]interface{}) {
	tx, err := client.PostRequest(contract, entryPoint, chainclient.PostRequestParams{
		Transfer: cbalances.NewFromMap(transfer),
		Args:     requestargs.New().AddEncodeSimpleMany(codec.MakeDict(params)),
	})
	check(err, t)
	err = chain.CommitteeMultiClient().WaitUntilAllRequestsProcessed(tx, 30*time.Second)
	check(err, t)
	if !counter.WaitUntilExpectationsMet() {
		t.Fail()
	}
}

func setup(t *testing.T, configPath string) {
	clu = clutest.NewCluster(t)
}

func setupAndLoad(t *testing.T, name string, description string, nrOfRequests int, expectedMessages map[string]int) {
	setup(t, "test_cluster")

	expectations := map[string]int{
		"chainrec":            2,
		"active_committee":    1,
		"dismissed_committee": 0,
		//"state":               3 + nrOfRequests,
		"request_in": 2 + nrOfRequests,
		//"request_out": 3 + nrOfRequests,    // not always coming from all nodes, but from quorum only
	}
	if nrOfRequests == 1 {
		expectations["state"] = 4
	}
	for k, v := range expectedMessages {
		expectations[k] = v
	}

	var err error

	counter, err = clu.StartMessageCounter(expectations)
	check(err, t)

	chain, err = clu.DeployDefaultChain()
	check(err, t)

	err = deployContract(name, description, nil)
	check(err, t)

	err = requestFunds(clu, scOwnerAddr, "client")
	check(err, t)

	client = chainclient.New(clu.Level1Client(), clu.WaspClient(0), chain.ChainID, scOwner.SigScheme())
}
