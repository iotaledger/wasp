package tests

import (
	"flag"
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	"github.com/iotaledger/goshimmer/client/wallet/packages/seed"
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/client/chainclient"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/requestargs"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/tools/cluster"
	clutest "github.com/iotaledger/wasp/tools/cluster/testutil"
	"github.com/mr-tron/base58"
	"github.com/stretchr/testify/require"
)

var (
	useGo       = flag.Bool("go", false, "use Go instead of Rust")
	useWasp     = flag.Bool("wasp", false, "use Wasp built-in instead of Rust")
	wallet      = initSeed()
	scOwner     = wallet.KeyPair(0)
	scOwnerAddr = ledgerstate.NewED25519Address(scOwner.PublicKey)
	chain       *cluster.Chain
	clu         *cluster.Cluster
	client      *chainclient.Client
	counter     *cluster.MessageCounter
	programHash hashing.HashValue
	err         error
)

func initSeed() *seed.Seed {
	b, err := base58.Decode("C6hPhCS2E2dKUGS3qj4264itKXohwgL3Lm2fNxayAKr")
	if err != nil {
		panic(err)
	}
	return seed.NewSeed(b)
}

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

func deployContract(wasmName, scDescription string, initParams map[string]interface{}) error {
	wasmPath := "wasm/" + wasmName + "_bg.wasm"
	if *useGo {
		wasmPath = "wasm/" + wasmName + "_go.wasm"
	}

	if !*useWasp {
		wasm, err := ioutil.ReadFile(wasmPath)
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
	// return nil
}

func postRequest(t *testing.T, contract, entryPoint coretypes.Hname, tokens int, params map[string]interface{}) {
	var transfer map[ledgerstate.Color]uint64
	if tokens != 0 {
		transfer = map[ledgerstate.Color]uint64{
			ledgerstate.ColorIOTA: uint64(tokens),
		}
	}
	postRequestFull(t, contract, entryPoint, transfer, params)
}

func postRequestFull(t *testing.T, contract, entryPoint coretypes.Hname, transfer map[ledgerstate.Color]uint64, params map[string]interface{}) {
	var b *ledgerstate.ColoredBalances
	if transfer != nil {
		b = ledgerstate.NewColoredBalances(transfer)
	}
	tx, err := client.Post1Request(contract, entryPoint, chainclient.PostRequestParams{
		Transfer: b,
		Args:     requestargs.New().AddEncodeSimpleMany(codec.MakeDict(params)),
	})
	check(err, t)
	err = chain.CommitteeMultiClient().WaitUntilAllRequestsProcessed(client.ChainID, tx, 60*time.Second)
	check(err, t)
	if !counter.WaitUntilExpectationsMet() {
		t.Fail()
	}
}

func setup(t *testing.T, configPath string) { //nolint:unparam
	clu = clutest.NewCluster(t)
	chain, err = clu.DeployDefaultChain()
	check(err, t)
}

func setupAndLoad(t *testing.T, name, description string, nrOfRequests int) {
	setup(t, "test_cluster")

	expectations := map[string]int{
		"dismissed_committee": 0,
		"state":               3 + nrOfRequests,
		//"request_out":         3 + nrOfRequests,    // not always coming from all nodes, but from quorum only
	}

	counter, err = clu.StartMessageCounter(expectations)
	check(err, t)

	chain, err = clu.DeployDefaultChain()
	check(err, t)

	err = deployContract(name, description, nil)
	check(err, t)

	err = requestFunds(clu, scOwnerAddr, "client")
	check(err, t)

	client = chainclient.New(clu.GoshimmerClient(), clu.WaspClient(0), chain.ChainID, scOwner)
}
