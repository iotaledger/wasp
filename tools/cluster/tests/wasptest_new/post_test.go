package wasptest

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/client/chainclient"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/root"
	"github.com/iotaledger/wasp/packages/vm/examples/inccounter"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func deployInccounter42(t *testing.T, name string, counter int64) coretypes.ContractID {
	hname := coretypes.Hn(name)
	description := "testing contract deployment with inccounter"
	programHash = inccounter.ProgramHash

	_, err = chain.DeployContract(name, inccounter.ProgramHashStr, description, map[string]interface{}{
		inccounter.VarCounter: counter,
		root.ParamName:        name,
	})
	check(err, t)

	chain.WithSCState(root.Interface.Hname(), func(host string, blockIndex uint32, state codec.ImmutableMustCodec) bool {
		require.EqualValues(t, 2, blockIndex)
		checkRoots(t, chain)

		contractRegistry := state.GetMap(root.VarContractRegistry)
		crBytes := contractRegistry.GetAt(hname.Bytes())
		require.NotNil(t, crBytes)
		cr, err := root.DecodeContractRecord(crBytes)
		check(err, t)

		require.EqualValues(t, programHash, cr.ProgramHash)
		require.EqualValues(t, description, cr.Description)
		require.EqualValues(t, 0, cr.NodeFee)
		require.EqualValues(t, cr.Name, name)

		return true
	})

	chain.WithSCState(hname, func(host string, blockIndex uint32, state codec.ImmutableMustCodec) bool {
		counterValue, _ := state.GetInt64(inccounter.VarCounter)
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

	c := codec.NewMustCodec(ret)
	counter, _ := c.GetInt64(inccounter.VarCounter)
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

	myClient := chainclient.New(clu.NodeClient, clu.WaspClient(0), chain.ChainID, mySigScheme)

	tx, err := myClient.PostRequest(contractID.Hname(), inccounter.EntryPointIncCounter, nil, nil, nil)
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

	myClient := chainclient.New(clu.NodeClient, clu.WaspClient(0), chain.ChainID, mySigScheme)

	tx, err := myClient.PostRequest(contractID.Hname(), inccounter.EntryPointIncAndRepeatMany, nil,
		map[balance.Color]int64{
			balance.ColorIOTA: 1, // needs 1 iota for recursive calls
		},
		codec.EncodeDictFromMap(map[string]interface{}{
			inccounter.VarNumRepeats: 3,
		}),
	)
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

	myClient := chainclient.New(clu.NodeClient, clu.WaspClient(0), chain.ChainID, mySigScheme)

	tx1, err := myClient.PostRequest(contractID.Hname(), inccounter.EntryPointIncCounter, nil, nil, nil)
	check(err, t)
	err = chain.CommitteeMultiClient().WaitUntilAllRequestsProcessed(tx1, 30*time.Second)
	check(err, t)

	tx2, err := myClient.PostRequest(contractID.Hname(), inccounter.EntryPointIncCounter, nil, nil, nil)
	check(err, t)
	err = chain.CommitteeMultiClient().WaitUntilAllRequestsProcessed(tx2, 30*time.Second)
	check(err, t)

	tx3, err := myClient.PostRequest(contractID.Hname(), inccounter.EntryPointIncCounter, nil, nil, nil)
	check(err, t)
	err = chain.CommitteeMultiClient().WaitUntilAllRequestsProcessed(tx3, 30*time.Second)
	check(err, t)

	tx4, err := myClient.PostRequest(contractID.Hname(), inccounter.EntryPointIncCounter, nil, nil, nil)
	check(err, t)
	err = chain.CommitteeMultiClient().WaitUntilAllRequestsProcessed(tx4, 30*time.Second)
	check(err, t)

	tx5, err := myClient.PostRequest(contractID.Hname(), inccounter.EntryPointIncCounter, nil, nil, nil)
	check(err, t)
	err = chain.CommitteeMultiClient().WaitUntilAllRequestsProcessed(tx5, 30*time.Second)
	check(err, t)

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

	myClient := chainclient.New(clu.NodeClient, clu.WaspClient(0), chain.ChainID, mySigScheme)

	tx1, err := myClient.PostRequest(contractID.Hname(), inccounter.EntryPointIncCounter, nil, nil, nil)
	check(err, t)
	tx2, err := myClient.PostRequest(contractID.Hname(), inccounter.EntryPointIncCounter, nil, nil, nil)
	check(err, t)
	tx3, err := myClient.PostRequest(contractID.Hname(), inccounter.EntryPointIncCounter, nil, nil, nil)
	check(err, t)
	tx4, err := myClient.PostRequest(contractID.Hname(), inccounter.EntryPointIncCounter, nil, nil, nil)
	check(err, t)
	tx5, err := myClient.PostRequest(contractID.Hname(), inccounter.EntryPointIncCounter, nil, nil, nil)
	check(err, t)

	err = chain.CommitteeMultiClient().WaitUntilAllRequestsProcessed(tx1, 30*time.Second)
	//check(err, t)
	err = chain.CommitteeMultiClient().WaitUntilAllRequestsProcessed(tx2, 30*time.Second)
	//check(err, t)
	err = chain.CommitteeMultiClient().WaitUntilAllRequestsProcessed(tx3, 30*time.Second)
	//check(err, t)
	err = chain.CommitteeMultiClient().WaitUntilAllRequestsProcessed(tx4, 30*time.Second)
	//check(err, t)
	err = chain.CommitteeMultiClient().WaitUntilAllRequestsProcessed(tx5, 30*time.Second)
	//check(err, t)

	expectCounter(t, contractID.Hname(), 42+5)
	checkBalanceOnChain(t, chain, myAgentID, balance.ColorIOTA, 5)

	if !clu.VerifyAddressBalances(myAddress, testutil.RequestFundsAmount-5, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount - 5,
	}, "myAddress in the end") {
		t.Fail()
	}
	checkLedger(t, chain)
}
