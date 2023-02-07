package tests

import (
	"context"
	"testing"
	"time"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/clients/apiclient"
	"github.com/iotaledger/wasp/clients/apiextensions"
	"github.com/iotaledger/wasp/clients/chainclient"
	"github.com/iotaledger/wasp/contracts/native/inccounter"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/utxodb"
	"github.com/iotaledger/wasp/packages/vm/core/root"
)

const inccounterName = "inc"

func deployInccounter42(e *ChainEnv) *isc.ContractAgentID {
	hname := isc.Hn(inccounterName)
	description := "testing contract deployment with inccounter"
	programHash := inccounter.Contract.ProgramHash

	_, err := e.Chain.DeployContract(inccounterName, programHash.String(), description, map[string]interface{}{
		inccounter.VarCounter: 42,
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

		cr, ok := lo.Find(contractRegistry, func(item apiclient.ContractInfoResponse) bool {
			return item.HName == hname.String()
		})
		require.True(e.t, ok)
		require.NotNil(e.t, cr)

		require.EqualValues(e.t, programHash.Hex(), cr.ProgramHash)
		require.EqualValues(e.t, description, cr.Description)
		require.EqualValues(e.t, cr.Name, inccounterName)

		counterValue, err := e.Chain.GetCounterValue(hname, i)
		require.NoError(e.t, err)
		require.EqualValues(e.t, 42, counterValue)
	}

	result, err := apiextensions.CallView(context.Background(), e.Chain.Cluster.WaspClient(), apiclient.ContractCallViewRequest{
		ChainId:       e.Chain.ChainID.String(),
		ContractHName: root.Contract.Hname().String(),
		FunctionHName: root.ViewFindContract.Hname().String(),
		Arguments: apiextensions.DictToAPIJsonDict(dict.Dict{
			root.ParamHname: hname.Bytes(),
		}),
	})

	recb, err := result.Get(root.ParamContractRecData)
	require.NoError(e.t, err)

	rec, err := root.ContractRecordFromBytes(recb)
	require.NoError(e.t, err)
	require.EqualValues(e.t, description, rec.Description)

	e.expectCounter(hname, 42)
	return isc.NewContractAgentID(e.Chain.ChainID, hname)
}

func (e *ChainEnv) expectCounter(hname isc.Hname, counter int64) {
	c := e.getNativeContractCounter(hname)
	require.EqualValues(e.t, counter, c)
}

func (e *ChainEnv) getNativeContractCounter(hname isc.Hname) int64 {
	return e.getCounterForNode(hname, 0)
}

func (e *ChainEnv) getCounterForNode(hname isc.Hname, nodeIndex int) int64 {
	result, _, err := e.Chain.Cluster.WaspClient().RequestsApi.CallView(context.Background()).ContractCallViewRequest(apiclient.ContractCallViewRequest{
		ChainId:       e.Chain.ChainID.String(),
		ContractHName: hname.String(),
		FunctionName:  "getCounter",
	}).Execute()

	decodedDict, err := apiextensions.APIJsonDictToDict(*result)
	require.NoError(e.t, err)

	counter, err := codec.DecodeInt64(decodedDict.MustGet(inccounter.VarCounter), 0)
	require.NoError(e.t, err)

	return counter
}

func (e *ChainEnv) waitUntilCounterEquals(hname isc.Hname, expected int64, duration time.Duration) {
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

// executed in cluster_test.go
func testPostDeployInccounter(t *testing.T, e *ChainEnv) {
	contractID := deployInccounter42(e)
	t.Logf("-------------- deployed contract. Name: '%s' id: %s", inccounterName, contractID.String())
}

// executed in cluster_test.go
func testPost1Request(t *testing.T, e *ChainEnv) {
	contractID := deployInccounter42(e)
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

// executed in cluster_test.go
func testPost3Recursive(t *testing.T, e *ChainEnv) {
	contractID := deployInccounter42(e)
	t.Logf("-------------- deployed contract. Name: '%s' id: %s", inccounterName, contractID.String())

	myWallet, _, err := e.Clu.NewKeyPairWithFunds()
	require.NoError(t, err)

	myClient := e.Chain.SCClient(contractID.Hname(), myWallet)

	tx, err := myClient.PostRequest(inccounter.FuncIncAndRepeatMany.Name, chainclient.PostRequestParams{
		Transfer:  isc.NewAssetsBaseTokens(10 * isc.Million),
		Allowance: isc.NewAssetsBaseTokens(9 * isc.Million),
		Args: codec.MakeDict(map[string]interface{}{
			inccounter.VarNumRepeats: 3,
		}),
	})
	require.NoError(t, err)

	_, err = e.Chain.CommitteeMultiClient().WaitUntilAllRequestsProcessedSuccessfully(e.Chain.ChainID, tx, 30*time.Second)
	require.NoError(t, err)

	e.waitUntilCounterEquals(contractID.Hname(), 43+3, 10*time.Second)
}

// executed in cluster_test.go
func testPost5Requests(t *testing.T, e *ChainEnv) {
	contractID := deployInccounter42(e)
	t.Logf("-------------- deployed contract. Name: '%s' id: %s", inccounterName, contractID.String())

	myWallet, myAddress, err := e.Clu.NewKeyPairWithFunds()
	require.NoError(t, err)
	myAgentID := isc.NewAgentID(myAddress)
	myClient := e.Chain.SCClient(contractID.Hname(), myWallet)

	e.checkBalanceOnChain(myAgentID, isc.BaseTokenID, 0)
	onChainBalance := uint64(0)
	for i := 0; i < 5; i++ {
		baseTokesSent := 1 * isc.Million
		tx, err := myClient.PostRequest(inccounter.FuncIncCounter.Name, chainclient.PostRequestParams{
			Transfer: isc.NewAssets(baseTokesSent, nil),
		})
		require.NoError(t, err)

		receipts, err := e.Chain.CommitteeMultiClient().WaitUntilAllRequestsProcessedSuccessfully(e.Chain.ChainID, tx, 30*time.Second)
		require.NoError(t, err)

		gasFeeCharged, err := iotago.DecodeUint64(receipts[0].GasFeeCharged)
		require.NoError(t, err)

		onChainBalance += baseTokesSent - gasFeeCharged
	}

	e.expectCounter(contractID.Hname(), 42+5)
	e.checkBalanceOnChain(myAgentID, isc.BaseTokenID, onChainBalance)

	e.checkLedger()
}

// executed in cluster_test.go
func testPost5AsyncRequests(t *testing.T, e *ChainEnv) {
	contractID := deployInccounter42(e)
	t.Logf("-------------- deployed contract. Name: '%s' id: %s", inccounterName, contractID.String())

	myWallet, myAddress, err := e.Clu.NewKeyPairWithFunds()
	require.NoError(t, err)
	myAgentID := isc.NewAgentID(myAddress)

	myClient := e.Chain.SCClient(contractID.Hname(), myWallet)

	tx := [5]*iotago.Transaction{}
	onChainBalance := uint64(0)
	baseTokesSent := 1 * isc.Million
	for i := 0; i < 5; i++ {
		tx[i], err = myClient.PostRequest(inccounter.FuncIncCounter.Name, chainclient.PostRequestParams{
			Transfer: isc.NewAssets(baseTokesSent, nil),
		})
		require.NoError(t, err)
	}

	for i := 0; i < 5; i++ {
		receipts, err := e.Chain.CommitteeMultiClient().WaitUntilAllRequestsProcessedSuccessfully(e.Chain.ChainID, tx[i], 30*time.Second)
		require.NoError(t, err)

		gasFeeCharged, err := iotago.DecodeUint64(receipts[0].GasFeeCharged)
		require.NoError(t, err)

		onChainBalance += baseTokesSent - gasFeeCharged
	}

	e.expectCounter(contractID.Hname(), 42+5)
	e.checkBalanceOnChain(myAgentID, isc.BaseTokenID, onChainBalance)

	if !e.Clu.AssertAddressBalances(myAddress,
		isc.NewAssetsBaseTokens(utxodb.FundsFromFaucetAmount-5*baseTokesSent)) {
		t.Fatal()
	}
	e.checkLedger()
}
