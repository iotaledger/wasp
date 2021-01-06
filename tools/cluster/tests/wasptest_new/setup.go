package wasptest

import (
	"flag"
	"fmt"
	"io/ioutil"
	"path"
	"runtime"
	"testing"
	"time"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/client/chainclient"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/tools/cluster"
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
		time.Sleep(time.Second)
		wasmPath = wasmName + "_go.wasm"
	}

	if !*useWasp {
		wasm, err := ioutil.ReadFile("wasm/" + wasmPath)
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
		Transfer: transfer,
		Args:     codec.MakeDict(params),
	})
	check(err, t)
	err = chain.CommitteeMultiClient().WaitUntilAllRequestsProcessed(tx, 30*time.Second)
	check(err, t)
	if !clu.WaitUntilExpectationsMet() {
		t.Fail()
	}
}

func setup(t *testing.T, configPath string) {
	if testing.Short() {
		t.Skip("Skipping cluster test in short mode")
	}

	_, filename, _, _ := runtime.Caller(0)

	clu, err = cluster.New(path.Join(path.Dir(filename), "..", configPath), "cluster-data")
	check(err, t)

	err = clu.Init(true, t.Name())
	check(err, t)

	err = clu.Start()
	check(err, t)

	t.Cleanup(clu.Stop)
}

func setupAndLoad(t *testing.T, name string, description string, nrOfRequests int, expectedMessages map[string]int) {
	setup(t, "test_cluster")

	expectations := map[string]int{
		"chainrec":            2,
		"active_committee":    1,
		"dismissed_committee": 0,
		//"state":               3 + nrOfRequests,
		"request_in":  2 + nrOfRequests,
		"request_out": 3 + nrOfRequests,
	}
	if nrOfRequests == 1 {
		expectations["state"] = 4
	}
	for k, v := range expectedMessages {
		expectations[k] = v
	}
	err := clu.ListenToMessages(expectations)
	check(err, t)

	chain, err = clu.DeployDefaultChain()
	check(err, t)

	err = deployContract(name, description, nil)
	check(err, t)

	err = requestFunds(clu, scOwnerAddr, "client")
	check(err, t)

	client = chainclient.New(clu.Level1Client, clu.WaspClient(0), chain.ChainID, scOwner.SigScheme())
}
