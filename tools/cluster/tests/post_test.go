package tests

import (
	"github.com/iotaledger/wasp/packages/coretypes/cbalances"
	"github.com/iotaledger/wasp/packages/coretypes/requestargs"
	"testing"
	"time"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/client/chainclient"
	"github.com/iotaledger/wasp/contracts/native/inccounter"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/stretchr/testify/require"
)

func deployInccounter42(t *testing.T, name string, counter int64) coretypes.ContractID {
	hname := coretypes.Hn(name)
	description := "testing contract deployment with inccounter"
	programHash = inccounter.Interface.ProgramHash

	_, err = chain.DeployContract(name, programHash.String(), description, map[string]interface{}{
		inccounter.VarCounter: counter,
		root.ParamName:        name,
	})
	check(err, t)

	chain.WithSCState(root.Interface.Hname(), func(host string, blockIndex uint32, state dict.Dict) bool {
		require.EqualValues(t, 2, blockIndex)
		checkRoots(t, chain)

		contractRegistry := collections.NewMapReadOnly(state, root.VarContractRegistry)
		crBytes := contractRegistry.MustGetAt(hname.Bytes())
		require.NotNil(t, crBytes)
		cr, err := root.DecodeContractRecord(crBytes)
		check(err, t)

		require.EqualValues(t, programHash, cr.ProgramHash)
		require.EqualValues(t, description, cr.Description)
		require.EqualValues(t, 0, cr.OwnerFee)
		require.EqualValues(t, cr.Name, name)

		return true
	})

	chain.WithSCState(hname, func(host string, blockIndex uint32, state dict.Dict) bool {
		counterValue, _, _ := codec.DecodeInt64(state.MustGet(inccounter.VarCounter))
		require.EqualValues(t, 42, counterValue)
		return true
	})

	// test calling root.FuncFindContractByName view function using client
	ret, err := chain.Cluster.WaspClient(0).CallView(
		chain.ContractID(root.Interface.Hname()),
		root.FuncFindContract,
		dict.FromGoMap(map[kv.Key][]byte{
			root.ParamHname: hname.Bytes(),
		}),
	)
	check(err, t)
	recb, err := ret.Get(root.ParamData)
	check(err, t)
	rec, err := root.DecodeContractRecord(recb)
	check(err, t)
	require.EqualValues(t, description, rec.Description)

	expectCounter(t, hname, counter)
	return coretypes.NewContractID(chain.ChainID, hname)
}

func expectCounter(t *testing.T, hname coretypes.Hname, counter int64) {
	c := getCounter(t, hname)
	require.EqualValues(t, counter, c)
}

func getCounter(t *testing.T, hname coretypes.Hname) int64 {
	ret, err := chain.Cluster.WaspClient(0).CallView(
		chain.ContractID(hname),
		"getCounter",
		nil,
	)
	check(err, t)

	counter, _, err := codec.DecodeInt64(ret.MustGet(inccounter.VarCounter))
	check(err, t)

	return counter
}

func TestPostDeployInccounter(t *testing.T) {
	setup(t, "test_cluster")

	chain, err = clu.DeployDefaultChain()
	check(err, t)

	name := "inc"
	contractID := deployInccounter42(t, name, 42)
	t.Logf("-------------- deployed contract. Name: '%s' id: %s", name, contractID.String())
}

func TestPost1Request(t *testing.T) {
	setup(t, "test_cluster")

	chain, err = clu.DeployDefaultChain()
	check(err, t)

	name := "inc"
	contractID := deployInccounter42(t, name, 42)
	t.Logf("-------------- deployed contract. Name: '%s' id: %s", name, contractID.String())

	testOwner := wallet.WithIndex(1)
	mySigScheme := testOwner.SigScheme()
	myAddress := testOwner.Address()
	err = requestFunds(clu, myAddress, "myAddress")
	check(err, t)

	myClient := chain.SCClient(contractID.Hname(), mySigScheme)

	tx, err := myClient.PostRequest(inccounter.FuncIncCounter)
	check(err, t)

	err = chain.CommitteeMultiClient().WaitUntilAllRequestsProcessed(tx, 30*time.Second)
	check(err, t)

	expectCounter(t, contractID.Hname(), 43)
}

func TestPost3Recursive(t *testing.T) {
	setup(t, "test_cluster")

	chain, err = clu.DeployDefaultChain()
	check(err, t)

	name := "inc"
	contractID := deployInccounter42(t, name, 42)
	t.Logf("-------------- deployed contract. Name: '%s' id: %s", name, contractID.String())

	testOwner := wallet.WithIndex(1)
	mySigScheme := testOwner.SigScheme()
	myAddress := testOwner.Address()
	err = requestFunds(clu, myAddress, "myAddress")
	check(err, t)

	myClient := chain.SCClient(contractID.Hname(), mySigScheme)

	tx, err := myClient.PostRequest(inccounter.FuncIncAndRepeatMany, chainclient.PostRequestParams{
		Transfer: cbalances.NewIotasOnly(1),
		Args: requestargs.New().AddEncodeSimpleMany(codec.MakeDict(map[string]interface{}{
			inccounter.VarNumRepeats: 3,
		})),
	})
	check(err, t)

	err = chain.CommitteeMultiClient().WaitUntilAllRequestsProcessed(tx, 30*time.Second)
	check(err, t)

	// must wait for recursion to complete
	time.Sleep(10 * time.Second)

	expectCounter(t, contractID.Hname(), 43+3)
}

func TestPost5Requests(t *testing.T) {
	setup(t, "test_cluster")

	chain, err = clu.DeployDefaultChain()
	check(err, t)

	name := "inc"
	contractID := deployInccounter42(t, name, 42)
	t.Logf("-------------- deployed contract. Name: '%s' id: %s", name, contractID.String())

	testOwner := wallet.WithIndex(1)
	mySigScheme := testOwner.SigScheme()
	myAddress := testOwner.Address()
	myAgentID := coretypes.NewAgentIDFromAddress(*myAddress)
	err = requestFunds(clu, myAddress, "myAddress")
	check(err, t)

	myClient := chain.SCClient(contractID.Hname(), mySigScheme)

	for i := 0; i < 5; i++ {
		tx, err := myClient.PostRequest(inccounter.FuncIncCounter)
		check(err, t)
		err = chain.CommitteeMultiClient().WaitUntilAllRequestsProcessed(tx, 30*time.Second)
		check(err, t)
	}

	expectCounter(t, contractID.Hname(), 42+5)
	checkBalanceOnChain(t, chain, myAgentID, balance.ColorIOTA, 5)

	if !clu.VerifyAddressBalances(myAddress, testutil.RequestFundsAmount-5, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount - 5,
	}, "myAddress in the end") {
		t.Fail()
	}
	checkLedger(t, chain)
}

func TestPost5AsyncRequests(t *testing.T) {
	setup(t, "test_cluster")

	chain, err = clu.DeployDefaultChain()
	check(err, t)

	name := "inc"
	contractID := deployInccounter42(t, name, 42)
	t.Logf("-------------- deployed contract. Name: '%s' id: %s", name, contractID.String())

	testOwner := wallet.WithIndex(1)
	mySigScheme := testOwner.SigScheme()
	myAddress := testOwner.Address()
	myAgentID := coretypes.NewAgentIDFromAddress(*myAddress)
	err = requestFunds(clu, myAddress, "myAddress")
	check(err, t)

	myClient := chain.SCClient(contractID.Hname(), mySigScheme)

	tx := [5]*sctransaction.Transaction{}
	var err error

	for i := 0; i < 5; i++ {
		tx[i], err = myClient.PostRequest(inccounter.FuncIncCounter)
		check(err, t)
	}

	for i := 0; i < 5; i++ {
		err = chain.CommitteeMultiClient().WaitUntilAllRequestsProcessed(tx[i], 30*time.Second)
		check(err, t)
	}

	expectCounter(t, contractID.Hname(), 42+5)
	checkBalanceOnChain(t, chain, myAgentID, balance.ColorIOTA, 5)

	if !clu.VerifyAddressBalances(myAddress, testutil.RequestFundsAmount-5, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount - 5,
	}, "myAddress in the end") {
		t.Fail()
	}
	checkLedger(t, chain)
}
