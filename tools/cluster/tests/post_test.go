package tests

import (
	"testing"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/client/chainclient"
	"github.com/iotaledger/wasp/contracts/native/inccounter"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/requestargs"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/stretchr/testify/require"
)

const name = "inc"

func deployInccounter42(t *testing.T, name string, counter int64) *iscp.AgentID {
	hname := iscp.Hn(name)
	description := "testing contract deployment with inccounter"
	programHash = inccounter.Interface.ProgramHash

	_, err = chain.DeployContract(name, programHash.String(), description, map[string]interface{}{
		inccounter.VarCounter: counter,
		root.ParamName:        name,
	})
	check(err, t)

	checkCoreContracts(t, chain)
	for i := range chain.CommitteeNodes {
		blockIndex, err := chain.BlockIndex(i)
		require.NoError(t, err)
		require.EqualValues(t, 2, blockIndex)

		contractRegistry, err := chain.ContractRegistry(i)
		require.NoError(t, err)
		cr := contractRegistry[hname]

		require.EqualValues(t, programHash, cr.ProgramHash)
		require.EqualValues(t, description, cr.Description)
		require.EqualValues(t, 0, cr.OwnerFee)
		require.EqualValues(t, cr.Name, name)

		counterValue, err := chain.GetCounterValue(hname, i)
		require.NoError(t, err)
		require.EqualValues(t, 42, counterValue)
	}

	// test calling root.FuncFindContractByName view function using client
	ret, err := chain.Cluster.WaspClient(0).CallView(
		chain.ChainID, root.Interface.Hname(), root.FuncFindContract.Name,
		dict.Dict{
			root.ParamHname: hname.Bytes(),
		})
	check(err, t)
	recb, err := ret.Get(root.VarData)
	check(err, t)
	rec, err := root.DecodeContractRecord(recb)
	check(err, t)
	require.EqualValues(t, description, rec.Description)

	expectCounter(t, hname, counter)
	return iscp.NewAgentID(chain.ChainID.AsAddress(), hname)
}

func expectCounter(t *testing.T, hname iscp.Hname, counter int64) {
	c := getCounter(t, hname)
	require.EqualValues(t, counter, c)
}

func getCounter(t *testing.T, hname iscp.Hname) int64 {
	ret, err := chain.Cluster.WaspClient(0).CallView(
		chain.ChainID, hname, "getCounter",
	)
	check(err, t)

	counter, _, err := codec.DecodeInt64(ret.MustGet(inccounter.VarCounter))
	check(err, t)

	return counter
}

func TestPostDeployInccounter(t *testing.T) {
	setup(t, "test_cluster")

	contractID := deployInccounter42(t, name, 42)
	t.Logf("-------------- deployed contract. Name: '%s' id: %s", name, contractID.String())
}

func TestPost1Request(t *testing.T) {
	setup(t, "test_cluster")

	contractID := deployInccounter42(t, name, 42)
	t.Logf("-------------- deployed contract. Name: '%s' id: %s", name, contractID.String())

	testOwner := wallet.KeyPair(1)
	myAddress := ledgerstate.NewED25519Address(testOwner.PublicKey)
	err = requestFunds(clu, myAddress, "myAddress")
	check(err, t)

	myClient := chain.SCClient(contractID.Hname(), testOwner)

	tx, err := myClient.PostRequest(inccounter.FuncIncCounter.Name)
	check(err, t)

	err = chain.CommitteeMultiClient().WaitUntilAllRequestsProcessed(chain.ChainID, tx, 30*time.Second)
	check(err, t)

	expectCounter(t, contractID.Hname(), 43)
}

func TestPost3Recursive(t *testing.T) {
	setup(t, "test_cluster")

	contractID := deployInccounter42(t, name, 42)
	t.Logf("-------------- deployed contract. Name: '%s' id: %s", name, contractID.String())

	testOwner := wallet.KeyPair(1)
	myAddress := ledgerstate.NewED25519Address(testOwner.PublicKey)
	err = requestFunds(clu, myAddress, "myAddress")
	check(err, t)

	myClient := chain.SCClient(contractID.Hname(), testOwner)

	tx, err := myClient.PostRequest(inccounter.FuncIncAndRepeatMany.Name, chainclient.PostRequestParams{
		Transfer: iscp.NewTransferIotas(1),
		Args: requestargs.New().AddEncodeSimpleMany(codec.MakeDict(map[string]interface{}{
			inccounter.VarNumRepeats: 3,
		})),
	})
	check(err, t)

	err = chain.CommitteeMultiClient().WaitUntilAllRequestsProcessed(chain.ChainID, tx, 30*time.Second)
	check(err, t)

	// must wait for recursion to complete
	time.Sleep(10 * time.Second)

	expectCounter(t, contractID.Hname(), 43+3)
}

func TestPost5Requests(t *testing.T) {
	setup(t, "test_cluster")

	contractID := deployInccounter42(t, name, 42)
	t.Logf("-------------- deployed contract. Name: '%s' id: %s", name, contractID.String())

	testOwner := wallet.KeyPair(1)
	myAddress := ledgerstate.NewED25519Address(testOwner.PublicKey)
	myAgentID := iscp.NewAgentID(myAddress, 0)
	err = requestFunds(clu, myAddress, "myAddress")
	check(err, t)

	myClient := chain.SCClient(contractID.Hname(), testOwner)

	for i := 0; i < 5; i++ {
		tx, err := myClient.PostRequest(inccounter.FuncIncCounter.Name)
		check(err, t)
		err = chain.CommitteeMultiClient().WaitUntilAllRequestsProcessed(chain.ChainID, tx, 30*time.Second)
		check(err, t)
	}

	expectCounter(t, contractID.Hname(), 42+5)
	checkBalanceOnChain(t, chain, myAgentID, ledgerstate.ColorIOTA, 0)

	if !clu.VerifyAddressBalances(myAddress, solo.Saldo-5, map[ledgerstate.Color]uint64{
		ledgerstate.ColorIOTA: solo.Saldo - 5,
	}, "myAddress in the end") {
		t.Fail()
	}
	checkLedger(t, chain)
}

func TestPost5AsyncRequests(t *testing.T) {
	setup(t, "test_cluster")

	contractID := deployInccounter42(t, name, 42)
	t.Logf("-------------- deployed contract. Name: '%s' id: %s", name, contractID.String())

	testOwner := wallet.KeyPair(1)
	myAddress := ledgerstate.NewED25519Address(testOwner.PublicKey)
	myAgentID := iscp.NewAgentID(myAddress, 0)
	err = requestFunds(clu, myAddress, "myAddress")
	check(err, t)

	myClient := chain.SCClient(contractID.Hname(), testOwner)

	tx := [5]*ledgerstate.Transaction{}
	var err error

	for i := 0; i < 5; i++ {
		tx[i], err = myClient.PostRequest(inccounter.FuncIncCounter.Name)
		check(err, t)
	}

	for i := 0; i < 5; i++ {
		err = chain.CommitteeMultiClient().WaitUntilAllRequestsProcessed(chain.ChainID, tx[i], 30*time.Second)
		check(err, t)
	}

	expectCounter(t, contractID.Hname(), 42+5)
	checkBalanceOnChain(t, chain, myAgentID, ledgerstate.ColorIOTA, 0)

	if !clu.VerifyAddressBalances(myAddress, solo.Saldo-5, map[ledgerstate.Color]uint64{
		ledgerstate.ColorIOTA: solo.Saldo - 5,
	}, "myAddress in the end") {
		t.Fail()
	}
	checkLedger(t, chain)
}
