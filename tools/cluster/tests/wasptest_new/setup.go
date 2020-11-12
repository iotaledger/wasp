package wasptest

import (
	"errors"
	"flag"
	"fmt"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/root"
	"github.com/iotaledger/wasp/packages/vm/examples"
	"github.com/iotaledger/wasp/packages/vm/examples/donatewithfeedback/dwfimpl"
	"github.com/iotaledger/wasp/packages/vm/examples/fairauction"
	"github.com/iotaledger/wasp/packages/vm/examples/fairroulette"
	"github.com/iotaledger/wasp/packages/vm/examples/inccounter"
	"github.com/iotaledger/wasp/packages/vm/examples/tokenregistry"
	"github.com/iotaledger/wasp/plugins/wasmtimevm"
	"github.com/iotaledger/wasp/tools/cluster"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
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

var builtinProgramHash  = map[string]string{
	"donatewithfeedback" : dwfimpl.ProgramHash,
	"fairauction": fairauction.ProgramHash,
	"fairroulette": fairroulette.ProgramHash,
	"increment": inccounter.ProgramHash,
	"tokenregistry": tokenregistry.ProgramHash,
}

func check(err error, t *testing.T) {
	t.Helper()
	assert.NoError(t, err)
	if err != nil {
		t.FailNow()
	}
}

func setup(t *testing.T, configPath string) *cluster.Cluster {
	_, filename, _, _ := runtime.Caller(0)

	clu, err := cluster.New(path.Join(path.Dir(filename), "..", configPath), "cluster-data")
	check(err, t)

	err = clu.Init(true, t.Name())
	check(err, t)

	err = clu.Start()
	check(err, t)

	t.Cleanup(clu.Stop)

	return clu
}

func setupAndLoad(t *testing.T, name string, description string, nrOfRequests int, expectedMessages map[string]int) (*cluster.Cluster, *cluster.Chain){
	clu := setup(t, "test_cluster")

	expectations := map[string]int{
		"chainrec":            2,
		"active_committee":    1,
		"dismissed_committee": 0,
		"state":               3 + nrOfRequests,
		"request_in":          2 + nrOfRequests,
		"request_out":         3 + nrOfRequests,
	}
	for k,v := range expectedMessages {
		expectations[k] = v
	}
	err := clu.ListenToMessages(expectations)
	check(err, t)

	chain, err := clu.DeployDefaultChain()
	check(err, t)

	err = deployContract(chain, name, description, nil)
	check(err, t)

	return clu, chain
}

func deployContract(chain *cluster.Chain, wasmName string, scDescription string, initParams map[string]interface{}) error {
	wasmLoaded = true
	wasmPath := wasmName + "_bg.wasm"
	if *useGo {
		fmt.Println("Using Go Wasm SC instead of Rust Wasm SC")
		time.Sleep(time.Second)
		wasmPath = wasmName + "_go.wasm"
	}

	if !*useWasp {
		wasm, err := ioutil.ReadFile("../wasmtest/wasm/" + wasmPath)
		if err != nil {
			return err
		}
		_, err = chain.DeployExternalContract(wasmtimevm.PluginName, wasmName, scDescription, wasm, initParams)
		return err
	}

	fmt.Println("Using Wasp built-in SC instead of Rust Wasm SC")
	time.Sleep(time.Second)
	hash,ok := builtinProgramHash[wasmName]
	if !ok {
		return errors.New("Unknown built-in SC: " + wasmName)
	}

	// default name is wasmName, but caller can override that
	params := make(map[string]interface{})
	params[root.ParamName] = wasmName
	for k, v := range initParams {
		params[k] = v
	}
	_, err := chain.DeployBuiltinContract(examples.VMType, hash, scDescription, params)
	return err
}
