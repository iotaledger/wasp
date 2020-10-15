package wasmtest

import (
	"errors"
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/client"
	"github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/plugins/wasmtimevm"
	"github.com/iotaledger/wasp/tools/cluster"
	"github.com/mr-tron/base58"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"runtime"
	"testing"
	"time"
)

var (
	wallet        = testutil.NewWallet("C6hPhCS2E2dKUGS3qj4264itKXohwgL3Lm2fNxayAKr")
	scOwner       = wallet.WithIndex(0)
	scAddr        *address.Address
	scColor       *balance.Color
	scOwnerAddr   *address.Address
	scProgramHash *hashing.HashValue
	wasps         *cluster.Cluster
)

func check(err error, t *testing.T) {
	t.Helper()
	assert.NoError(t, err)
	if err != nil {
		t.FailNow()
	}
}

func checkSuccess(err error, t *testing.T, success string) {
	t.Helper()
	assert.NoError(t, err)
	if err != nil {
		t.FailNow()
	} else {
		fmt.Printf("[test] SUCCESS: %s\n", success)
	}
}

func loadWasmIntoWasps(t *testing.T, wasmPath string, scDescription string) error {
	wasm, err := ioutil.ReadFile(wasmPath)
	if err != nil {
		return err
	}
	scProgramHash = nil
	return wasps.MultiClient().Do(func(i int, w *client.WaspClient) error {
		var err error
		hashValue, err := w.PutProgram(wasmtimevm.PluginName, scDescription, wasm)
		if err != nil {
			return err
		}
		if scProgramHash == nil {
			scProgramHash = hashValue
			return nil
		}
		if *scProgramHash != *hashValue {
			return errors.New("code hash mismatch")
		}
		return nil
	})
}

func preamble(t *testing.T, wasmPath string, scDescription string, testName string) {
	var seed [32]byte
	rand.Read(seed[:])
	seed58 := base58.Encode(seed[:])
	wallet = testutil.NewWallet(seed58)
	scOwner = wallet.WithIndex(0)
	scOwnerAddr = scOwner.Address()

	// start wasp nodes
	startWasps(t, "test_cluster2", testName)

	// load sc code into wasps, save hash for later use
	err := loadWasmIntoWasps(t, wasmPath, scDescription)
	check(err, t)

	err = requestFunds(wasps, scOwnerAddr, "sc owner")
	check(err, t)
}

func requestFunds(wasps *cluster.Cluster, addr *address.Address, who string) error {
	err := wasps.NodeClient.RequestFunds(addr)
	if err != nil {
		return err
	}
	if !wasps.VerifyAddressBalances(addr, testutil.RequestFundsAmount, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount,
	}, "requested funds for " + who) {
		return errors.New("unexpected requested amount")
	}
	return nil
}

func startSmartContract(t *testing.T, scProgramHash *hashing.HashValue, scDescription string) {
	var err error
	scAddr, scColor, err = apilib.CreateSC(apilib.CreateSCParams{
		Node:                  wasps.NodeClient,
		CommitteeApiHosts:     wasps.ApiHosts(),
		CommitteePeeringHosts: wasps.PeeringHosts(),
		N:                     4,
		T:                     3,
		OwnerSigScheme:        scOwner.SigScheme(),
		ProgramHash:           *scProgramHash,
		Description:           scDescription,
		Textout:               os.Stdout,
		Prefix:                "[deploy] ",
	})
	checkSuccess(err, t, "smart contract has been created")

	err = apilib.ActivateSCMulti(apilib.ActivateSCParams{
		Addresses:         []*address.Address{scAddr},
		ApiHosts:          wasps.ApiHosts(),
		WaitForCompletion: true,
		PublisherHosts:    wasps.PublisherHosts(),
		Timeout:           30 * time.Second,
	})
	checkSuccess(err, t, "smart contract has been activated and initialized")
}

func startWasps(t *testing.T, configPath string, testName string) {
	var err error
	_, filename, _, _ := runtime.Caller(0)

	wasps, err = cluster.New(path.Join(path.Dir(filename), "..", configPath), "cluster-data")
	check(err, t)

	err = wasps.Init(true, testName)
	check(err, t)

	err = wasps.Start()
	check(err, t)

	t.Cleanup(wasps.Stop)
}
