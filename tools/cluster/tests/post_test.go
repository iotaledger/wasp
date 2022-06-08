package tests

import (
	"testing"
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/client/chainclient"
	"github.com/iotaledger/wasp/contracts/native/inccounter"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/utxodb"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/stretchr/testify/require"
)

const inccounterName = "inc"

func (e *ChainEnv) deployInccounter42(counter int64) *iscp.ContractAgentID {
	hname := iscp.Hn(inccounterName)
	description := "testing contract deployment with inccounter"
	programHash := inccounter.Contract.ProgramHash

	_, err := e.Chain.DeployContract(inccounterName, programHash.String(), description, map[string]interface{}{
		inccounter.VarCounter: counter,
		root.ParamName:        inccounterName,
	})
	require.NoError(e.t, err)

	e.checkCoreContracts()
	for i := range e.Chain.CommitteeNodes {
		blockIndex, err := e.Chain.BlockIndex(i)
		require.NoError(e.t, err)
		require.Greater(e.t, blockIndex, uint32(2))

		contractRegistry, err := e.Chain.ContractRegistry(i)
		require.NoError(e.t, err)
		cr := contractRegistry[hname]

		require.EqualValues(e.t, programHash, cr.ProgramHash)
		require.EqualValues(e.t, description, cr.Description)
		require.EqualValues(e.t, cr.Name, inccounterName)

		counterValue, err := e.Chain.GetCounterValue(hname, i)
		require.NoError(e.t, err)
		require.EqualValues(e.t, 42, counterValue)
	}

	// test calling root.FuncFindContractByName view function using client
	ret, err := e.Chain.Cluster.WaspClient(0).CallView(
		e.Chain.ChainID, root.Contract.Hname(), root.ViewFindContract.Name,
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
	return iscp.NewContractAgentID(e.Chain.ChainID, hname)
}

func (e *ChainEnv) expectCounter(hname iscp.Hname, counter int64) {
	c := e.getCounter(hname)
	require.EqualValues(e.t, counter, c)
}

func (e *ChainEnv) getCounter(hname iscp.Hname) int64 {
	return e.getCounterForNode(hname, 0)
}

func (e *ChainEnv) getCounterForNode(hname iscp.Hname, nodeIndex int) int64 {
	ret, err := e.Chain.Cluster.WaspClient(nodeIndex).CallView(
		e.Chain.ChainID, hname, "getCounter", nil,
	)
	require.NoError(e.t, err)

	counter, err := codec.DecodeInt64(ret.MustGet(inccounter.VarCounter), 0)
	require.NoError(e.t, err)

	return counter
}

func (e *ChainEnv) waitUntilCounterEquals(hname iscp.Hname, expected int64, duration time.Duration) {
	timeout := time.After(duration)
	var c int64
	allNodesEqualFun := func() bool {
		for _, node := range e.Chain.AllPeers {
			c = e.getCounterForNode(hname, node)
			if c != expected {
				return false
			}
		}
		return true
	}
	for {
		select {
		case <-timeout:
			e.t.Errorf("timeout waiting for inccounter, current: %d, expected: %d", c, expected)
			e.t.Fatal()
		default:
			if allNodesEqualFun() {
				return // success
			}
		}
		time.Sleep(1 * time.Second)
	}
}

func TestPostDeployInccounter(t *testing.T) {
	e := SetupWithChain(t)
	contractID := e.deployInccounter42(42)
	t.Logf("-------------- deployed contract. Name: '%s' id: %s", inccounterName, contractID.String())
}

func TestPost1Request(t *testing.T) {
	e := SetupWithChain(t)

	contractID := e.deployInccounter42(42)
	t.Logf("-------------- deployed contract. Name: '%s' id: %s", inccounterName, contractID.String())

	myWallet, _, err := e.Clu.NewKeyPairWithFunds()
	require.NoError(t, err)

	myClient := e.Chain.SCClient(contractID.Hname(), myWallet)

	tx, err := myClient.PostRequest(inccounter.FuncIncCounter.Name)
	require.NoError(t, err)

	_, err = e.Chain.CommitteeMultiClient().WaitUntilAllRequestsProcessedSuccessfully(e.Chain.ChainID, tx, 30*time.Second)
	require.NoError(t, err)

	e.expectCounter(contractID.Hname(), 43)
}

func TestPost3Recursive(t *testing.T) {
	e := SetupWithChain(t)

	contractID := e.deployInccounter42(42)
	t.Logf("-------------- deployed contract. Name: '%s' id: %s", inccounterName, contractID.String())

	myWallet, _, err := e.Clu.NewKeyPairWithFunds()
	require.NoError(t, err)

	myClient := e.Chain.SCClient(contractID.Hname(), myWallet)

	tx, err := myClient.PostRequest(inccounter.FuncIncAndRepeatMany.Name, chainclient.PostRequestParams{
		Transfer:  iscp.NewTokensIotas(10 * iscp.Mi),
		Allowance: iscp.NewAllowanceIotas(9 * iscp.Mi),
		Args: codec.MakeDict(map[string]interface{}{
			inccounter.VarNumRepeats: 3,
		}),
	})
	require.NoError(t, err)

	_, err = e.Chain.CommitteeMultiClient().WaitUntilAllRequestsProcessedSuccessfully(e.Chain.ChainID, tx, 30*time.Second)
	require.NoError(t, err)

	e.waitUntilCounterEquals(contractID.Hname(), 43+3, 10*time.Second)
}

func TestPost5Requests(t *testing.T) {
	e := SetupWithChain(t)

	contractID := e.deployInccounter42(42)
	t.Logf("-------------- deployed contract. Name: '%s' id: %s", inccounterName, contractID.String())

	myWallet, myAddress, err := e.Clu.NewKeyPairWithFunds()
	require.NoError(t, err)
	myAgentID := iscp.NewAgentID(myAddress)
	myClient := e.Chain.SCClient(contractID.Hname(), myWallet)

	e.checkBalanceOnChain(myAgentID, iscp.IotaTokenID, 0)
	onChainBalance := uint64(0)
	for i := 0; i < 5; i++ {
		iotasSent := 1 * iscp.Mi
		tx, err := myClient.PostRequest(inccounter.FuncIncCounter.Name, chainclient.PostRequestParams{
			Transfer: iscp.NewFungibleTokens(iotasSent, nil),
		})
		require.NoError(t, err)
		receipts, err := e.Chain.CommitteeMultiClient().WaitUntilAllRequestsProcessedSuccessfully(e.Chain.ChainID, tx, 30*time.Second)
		require.NoError(t, err)
		onChainBalance += iotasSent - receipts[0].GasFeeCharged
	}

	e.expectCounter(contractID.Hname(), 42+5)
	e.checkBalanceOnChain(myAgentID, iscp.IotaTokenID, onChainBalance)

	e.checkLedger()
}

func TestPost5AsyncRequests(t *testing.T) {
	e := SetupWithChain(t)

	contractID := e.deployInccounter42(42)
	t.Logf("-------------- deployed contract. Name: '%s' id: %s", inccounterName, contractID.String())

	myWallet, myAddress, err := e.Clu.NewKeyPairWithFunds()
	require.NoError(t, err)
	myAgentID := iscp.NewAgentID(myAddress)

	myClient := e.Chain.SCClient(contractID.Hname(), myWallet)

	tx := [5]*iotago.Transaction{}
	onChainBalance := uint64(0)
	iotasSent := 1 * iscp.Mi
	for i := 0; i < 5; i++ {
		tx[i], err = myClient.PostRequest(inccounter.FuncIncCounter.Name, chainclient.PostRequestParams{
			Transfer: iscp.NewFungibleTokens(iotasSent, nil),
		})
		require.NoError(t, err)
	}

	for i := 0; i < 5; i++ {
		receipts, err := e.Chain.CommitteeMultiClient().WaitUntilAllRequestsProcessedSuccessfully(e.Chain.ChainID, tx[i], 30*time.Second)
		require.NoError(t, err)
		onChainBalance += iotasSent - receipts[0].GasFeeCharged
	}

	e.expectCounter(contractID.Hname(), 42+5)
	e.checkBalanceOnChain(myAgentID, iscp.IotaTokenID, onChainBalance)

	if !e.Clu.AssertAddressBalances(myAddress,
		iscp.NewTokensIotas(utxodb.FundsFromFaucetAmount-5*iotasSent)) {
		t.Fatal()
	}
	e.checkLedger()
}
