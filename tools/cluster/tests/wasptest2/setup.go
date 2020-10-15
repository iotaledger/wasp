package wasptest2

import (
	"errors"
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/tools/cluster"
	"github.com/mr-tron/base58"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"os"
	"path"
	"runtime"
	"testing"
	"time"
)

var (
	seed = "C6hPhCS2E2dKUGS3qj4264itKXohwgL3Lm2fNxayAKr"
	wallet  = testutil.NewWallet(seed)
	scOwner = wallet.WithIndex(0)
	scOwnerAddr = scOwner.Address()
	programHash hashing.HashValue
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

func startSmartContract(wasps *cluster.Cluster, scProgramHash string, scDescription string)(*address.Address, *balance.Color, error) {
	var err error
	programHash, err = hashing.HashValueFromBase58(scProgramHash)
	if err != nil { return nil, nil, err}
	scAddr, scColor, err := apilib.CreateSC(apilib.CreateSCParams{
		Node:                  wasps.NodeClient,
		CommitteeApiHosts:     wasps.ApiHosts(),
		CommitteePeeringHosts: wasps.PeeringHosts(),
		N:                     4,
		T:                     3,
		OwnerSigScheme:        scOwner.SigScheme(),
		ProgramHash:           programHash,
		Description:           scDescription,
		Textout:               os.Stdout,
		Prefix:                "[deploy " + scProgramHash + "]",
	})
	if err != nil { return nil, nil, err}

	err = apilib.ActivateSCMulti(apilib.ActivateSCParams{
		Addresses:         []*address.Address{scAddr},
		ApiHosts:          wasps.ApiHosts(),
		WaitForCompletion: true,
		PublisherHosts:    wasps.PublisherHosts(),
		Timeout:           30 * time.Second,
	})
    return scAddr, scColor, err
}

func setup(t *testing.T, testName string) *cluster.Cluster {
	var seedBytes [32]byte
	rand.Read(seedBytes[:])
	seed = base58.Encode(seedBytes[:])
	wallet = testutil.NewWallet(seed)
	scOwner = wallet.WithIndex(0)
	scOwnerAddr = scOwner.Address()
	_, filename, _, _ := runtime.Caller(0)
	wasps, err := cluster.New(path.Join(path.Dir(filename), "..", "test_cluster2"), "cluster-data")
	check(err, t)
	err = wasps.Init(true, testName)
	check(err, t)
	err = wasps.Start()
	check(err, t)
	t.Cleanup(wasps.Stop)
	return wasps
}
