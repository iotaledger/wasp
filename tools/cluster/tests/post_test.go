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
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/testutil/utxodb"
	"github.com/iotaledger/wasp/packages/vm/core/root"
)

func deployInccounter42(e *ChainEnv) *isc.ContractAgentID {
	_, err := e.Chain.DeployContract(inccounter.Contract.Name, inccounter.Contract.ProgramHash.String(), inccounter.InitParams(42))
	require.NoError(e.t, err)

	e.checkCoreContracts()
	for i := range e.Chain.CommitteeNodes {
		blockIndex, err2 := e.Chain.BlockIndex(i)
		require.NoError(e.t, err2)
		require.Greater(e.t, blockIndex, uint32(2))

		contractRegistry, err2 := e.Chain.ContractRegistry(i)
		require.NoError(e.t, err2)

		cr, ok := lo.Find(contractRegistry, func(item apiclient.ContractInfoResponse) bool {
			return item.HName == inccounter.Contract.Hname().String()
		})
		require.True(e.t, ok)
		require.NotNil(e.t, cr)

		require.EqualValues(e.t, inccounter.Contract.ProgramHash.Hex(), cr.ProgramHash)
		require.EqualValues(e.t, cr.Name, inccounter.Contract.Name)

		counterValue, err2 := e.Chain.GetCounterValue(i)
		require.NoError(e.t, err2)
		require.EqualValues(e.t, 42, counterValue)
	}

	result, err := apiextensions.CallView(
		context.Background(),
		e.Chain.Cluster.WaspClient(),
		e.Chain.ChainID.String(),
		apiclient.ContractCallViewRequest{
			ContractHName: root.Contract.Hname().String(),
			FunctionHName: root.ViewFindContract.Hname().String(),
			Arguments: apiextensions.DictToAPIJsonDict(dict.Dict{
				root.ParamHname: inccounter.Contract.Hname().Bytes(),
			}),
		})
	require.NoError(e.t, err)

	recb := result.Get(root.ParamContractRecData)

	_, err = root.ContractRecordFromBytes(recb)
	require.NoError(e.t, err)

	e.expectCounter(42)
	return isc.NewContractAgentID(e.Chain.ChainID, inccounter.Contract.Hname())
}

// executed in cluster_test.go
func testPostDeployInccounter(t *testing.T, e *ChainEnv) {
	contractID := deployInccounter42(e)
	t.Logf("-------------- deployed contract. Name: '%s' id: %s", inccounter.Contract.Name, contractID.String())
}

// executed in cluster_test.go
func testPost1Request(t *testing.T, e *ChainEnv) {
	contractID := deployInccounter42(e)
	t.Logf("-------------- deployed contract. Name: '%s' id: %s", inccounter.Contract.Name, contractID.String())

	myWallet, _, err := e.Clu.NewKeyPairWithFunds()
	require.NoError(t, err)

	myClient := e.Chain.Client(myWallet)

	tx, err := myClient.PostRequest(inccounter.FuncIncCounter.Message(nil))
	require.NoError(t, err)

	_, err = e.Chain.CommitteeMultiClient().WaitUntilAllRequestsProcessedSuccessfully(e.Chain.ChainID, tx, false, 30*time.Second)
	require.NoError(t, err)

	e.expectCounter(43)
}

// executed in cluster_test.go
func testPost3Recursive(t *testing.T, e *ChainEnv) {
	contractID := deployInccounter42(e)
	t.Logf("-------------- deployed contract. Name: '%s' id: %s", inccounter.Contract.Name, contractID.String())

	myWallet, _, err := e.Clu.NewKeyPairWithFunds()
	require.NoError(t, err)

	myClient := e.Chain.Client(myWallet)

	var numRepeats int64 = 3
	tx, err := myClient.PostRequest(inccounter.FuncIncAndRepeatMany.Message(nil, &numRepeats), chainclient.PostRequestParams{
		Transfer:  isc.NewAssetsBaseTokens(10 * isc.Million),
		Allowance: isc.NewAssetsBaseTokens(9 * isc.Million),
	})
	require.NoError(t, err)

	_, err = e.Chain.CommitteeMultiClient().WaitUntilAllRequestsProcessedSuccessfully(e.Chain.ChainID, tx, false, 30*time.Second)
	require.NoError(t, err)

	e.waitUntilCounterEquals(43+3, 10*time.Second)
}

// executed in cluster_test.go
func testPost5Requests(t *testing.T, e *ChainEnv) {
	contractID := deployInccounter42(e)
	t.Logf("-------------- deployed contract. Name: '%s' id: %s", inccounter.Contract.Name, contractID.String())

	myWallet, myAddress, err := e.Clu.NewKeyPairWithFunds()
	require.NoError(t, err)
	myAgentID := isc.NewAgentID(myAddress)
	myClient := e.Chain.Client(myWallet)

	e.checkBalanceOnChain(myAgentID, isc.BaseTokenID, 0)
	onChainBalance := uint64(0)
	for i := 0; i < 5; i++ {
		baseTokesSent := 1 * isc.Million
		tx, err := myClient.PostRequest(inccounter.FuncIncCounter.Message(nil), chainclient.PostRequestParams{
			Transfer: isc.NewAssets(baseTokesSent, nil),
		})
		require.NoError(t, err)

		receipts, err := e.Chain.CommitteeMultiClient().WaitUntilAllRequestsProcessedSuccessfully(e.Chain.ChainID, tx, false, 30*time.Second)
		require.NoError(t, err)

		gasFeeCharged, err := iotago.DecodeUint64(receipts[0].GasFeeCharged)
		require.NoError(t, err)

		onChainBalance += baseTokesSent - gasFeeCharged
	}

	e.expectCounter(42 + 5)
	e.checkBalanceOnChain(myAgentID, isc.BaseTokenID, onChainBalance)
}

// executed in cluster_test.go
func testPost5AsyncRequests(t *testing.T, e *ChainEnv) {
	contractID := deployInccounter42(e)
	t.Logf("-------------- deployed contract. Name: '%s' id: %s", inccounter.Contract.Name, contractID.String())

	myWallet, myAddress, err := e.Clu.NewKeyPairWithFunds()
	require.NoError(t, err)
	myAgentID := isc.NewAgentID(myAddress)

	myClient := e.Chain.Client(myWallet)

	tx := [5]*iotago.Transaction{}
	onChainBalance := uint64(0)
	baseTokesSent := 1 * isc.Million
	for i := 0; i < 5; i++ {
		tx[i], err = myClient.PostRequest(inccounter.FuncIncCounter.Message(nil), chainclient.PostRequestParams{
			Transfer: isc.NewAssets(baseTokesSent, nil),
		})
		require.NoError(t, err)
	}

	for i := 0; i < 5; i++ {
		receipts, err := e.Chain.CommitteeMultiClient().WaitUntilAllRequestsProcessedSuccessfully(e.Chain.ChainID, tx[i], false, 30*time.Second)
		require.NoError(t, err)

		gasFeeCharged, err := iotago.DecodeUint64(receipts[0].GasFeeCharged)
		require.NoError(t, err)

		onChainBalance += baseTokesSent - gasFeeCharged
	}

	e.expectCounter(42 + 5)
	e.checkBalanceOnChain(myAgentID, isc.BaseTokenID, onChainBalance)

	if !e.Clu.AssertAddressBalances(myAddress,
		isc.NewAssetsBaseTokens(utxodb.FundsFromFaucetAmount-5*baseTokesSent)) {
		t.Fatal()
	}
}
