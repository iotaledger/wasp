package tests

import (
	"testing"
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/utxodb"

	"github.com/iotaledger/wasp/client/chainclient"
	"github.com/iotaledger/wasp/contracts/native/inccounter"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/stretchr/testify/require"
)

const inccounterName = "inc"

func (e *chainEnv) deployInccounter42(counter int64) *iscp.AgentID { //nolint:unparam
	hname := iscp.Hn(inccounterName)
	description := "testing contract deployment with inccounter"
	programHash := inccounter.Contract.ProgramHash

	_, err := e.chain.DeployContract(inccounterName, programHash.String(), description, map[string]interface{}{
		inccounter.VarCounter: counter,
		root.ParamName:        inccounterName,
	})
	require.NoError(e.t, err)

	e.checkCoreContracts()
	for i := range e.chain.CommitteeNodes {
		blockIndex, err := e.chain.BlockIndex(i)
		require.NoError(e.t, err)
		require.EqualValues(e.t, 2, blockIndex)

		contractRegistry, err := e.chain.ContractRegistry(i)
		require.NoError(e.t, err)
		cr := contractRegistry[hname]

		require.EqualValues(e.t, programHash, cr.ProgramHash)
		require.EqualValues(e.t, description, cr.Description)
		require.EqualValues(e.t, cr.Name, inccounterName)

		counterValue, err := e.chain.GetCounterValue(hname, i)
		require.NoError(e.t, err)
		require.EqualValues(e.t, 42, counterValue)
	}

	// test calling root.FuncFindContractByName view function using client
	ret, err := e.chain.Cluster.WaspClient(0).CallView(
		e.chain.ChainID, root.Contract.Hname(), root.FuncFindContract.Name,
		dict.Dict{
			root.ParamHname: hname.Bytes(),
		})
	require.NoError(e.t, err)
	recb, err := ret.Get(root.ParamContractRecData)
	require.NoError(e.t, err)
	rec, err := root.ContractRecordFromBytes(recb)
	require.NoError(e.t, err)
	require.EqualValues(e.t, description, rec.Description)

	e.expectCounter(hname, counter)
	return iscp.NewAgentID(e.chain.ChainID.AsAddress(), hname)
}

func (e *chainEnv) expectCounter(hname iscp.Hname, counter int64) {
	c := e.getCounter(hname)
	require.EqualValues(e.t, counter, c)
}

func (e *chainEnv) getCounter(hname iscp.Hname) int64 {
	ret, err := e.chain.Cluster.WaspClient(0).CallView(
		e.chain.ChainID, hname, "getCounter", nil,
	)
	require.NoError(e.t, err)

	counter, err := codec.DecodeInt64(ret.MustGet(inccounter.VarCounter), 0)
	require.NoError(e.t, err)

	return counter
}

func TestPostDeployInccounter(t *testing.T) {
	e := setupWithChain(t)
	contractID := e.deployInccounter42(42)
	t.Logf("-------------- deployed contract. Name: '%s' id: %s", inccounterName, contractID.String())
}

func TestPost1Request(t *testing.T) {
	e := setupWithChain(t)

	contractID := e.deployInccounter42(42)
	t.Logf("-------------- deployed contract. Name: '%s' id: %s", inccounterName, contractID.String())

	testOwner := cryptolib.NewKeyPairFromSeed(wallet.SubSeed(1))
	myAddress := testOwner.Address()
	e.requestFunds(myAddress, "myAddress")

	myClient := e.chain.SCClient(contractID.Hname(), testOwner)

	tx, err := myClient.PostRequest(inccounter.FuncIncCounter.Name)
	require.NoError(t, err)

	err = e.chain.CommitteeMultiClient().WaitUntilAllRequestsProcessed(e.chain.ChainID, tx, 30*time.Second)
	require.NoError(t, err)

	e.expectCounter(contractID.Hname(), 43)
}

func TestPost3Recursive(t *testing.T) {
	e := setupWithChain(t)

	contractID := e.deployInccounter42(42)
	t.Logf("-------------- deployed contract. Name: '%s' id: %s", inccounterName, contractID.String())

	testOwner := cryptolib.NewKeyPairFromSeed(wallet.SubSeed(1))
	myAddress := testOwner.Address()
	e.requestFunds(myAddress, "myAddress")

	myClient := e.chain.SCClient(contractID.Hname(), testOwner)

	tx, err := myClient.PostRequest(inccounter.FuncIncAndRepeatMany.Name, chainclient.PostRequestParams{
		Transfer: iscp.NewTokensIotas(1),
		Args: codec.MakeDict(map[string]interface{}{
			inccounter.VarNumRepeats: 3,
		}),
	})
	require.NoError(t, err)

	err = e.chain.CommitteeMultiClient().WaitUntilAllRequestsProcessed(e.chain.ChainID, tx, 30*time.Second)
	require.NoError(t, err)

	// must wait for recursion to complete
	time.Sleep(10 * time.Second)

	e.expectCounter(contractID.Hname(), 43+3)
}

func TestPost5Requests(t *testing.T) {
	e := setupWithChain(t)

	contractID := e.deployInccounter42(42)
	t.Logf("-------------- deployed contract. Name: '%s' id: %s", inccounterName, contractID.String())

	testOwner := cryptolib.NewKeyPairFromSeed(wallet.SubSeed(1))
	myAddress := testOwner.Address()
	myAgentID := iscp.NewAgentID(myAddress, 0)
	e.requestFunds(myAddress, "myAddress")

	myClient := e.chain.SCClient(contractID.Hname(), testOwner)

	for i := 0; i < 5; i++ {
		tx, err := myClient.PostRequest(inccounter.FuncIncCounter.Name)
		require.NoError(t, err)
		err = e.chain.CommitteeMultiClient().WaitUntilAllRequestsProcessed(e.chain.ChainID, tx, 30*time.Second)
		require.NoError(t, err)
	}

	e.expectCounter(contractID.Hname(), 42+5)
	e.checkBalanceOnChain(myAgentID, iscp.IotaTokenID, 0)

	if !e.clu.AssertAddressBalances(myAddress,
		iscp.NewTokensIotas(utxodb.FundsFromFaucetAmount-5)) {
		t.Fail()
	}
	e.checkLedger()
}

func TestPost5AsyncRequests(t *testing.T) {
	e := setupWithChain(t)

	contractID := e.deployInccounter42(42)
	t.Logf("-------------- deployed contract. Name: '%s' id: %s", inccounterName, contractID.String())

	testOwner := cryptolib.NewKeyPairFromSeed(wallet.SubSeed(1))
	myAddress := testOwner.Address()
	myAgentID := iscp.NewAgentID(myAddress, 0)
	e.requestFunds(myAddress, "myAddress")

	myClient := e.chain.SCClient(contractID.Hname(), testOwner)

	tx := [5]*iotago.Transaction{}
	var err error

	for i := 0; i < 5; i++ {
		tx[i], err = myClient.PostRequest(inccounter.FuncIncCounter.Name)
		require.NoError(t, err)
	}

	for i := 0; i < 5; i++ {
		err = e.chain.CommitteeMultiClient().WaitUntilAllRequestsProcessed(e.chain.ChainID, tx[i], 30*time.Second)
		require.NoError(t, err)
	}

	e.expectCounter(contractID.Hname(), 42+5)
	e.checkBalanceOnChain(myAgentID, iscp.IotaTokenID, 0)

	if !e.clu.AssertAddressBalances(myAddress,
		iscp.NewTokensIotas(utxodb.FundsFromFaucetAmount-5)) {
		t.Fail()
	}
	e.checkLedger()
}
