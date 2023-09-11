package tests

import (
	"context"
	"crypto/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/clients/apiclient"
	"github.com/iotaledger/wasp/clients/apiextensions"
	"github.com/iotaledger/wasp/clients/chainclient"
	"github.com/iotaledger/wasp/contracts/native/inccounter"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

func TestOffledgerRequestAccessNode(t *testing.T) {
	const clusterSize = 10
	clu := newCluster(t, waspClusterOpts{nNodes: clusterSize})

	cmt := []int{0, 1, 2, 3}

	addr, err := clu.RunDKG(cmt, 3)
	require.NoError(t, err)

	chain, err := clu.DeployChain(clu.Config.AllNodes(), cmt, 3, addr)
	require.NoError(t, err)

	e := newChainEnv(t, clu, chain)
	e.deployNativeIncCounterSC()

	waitUntil(t, e.contractIsDeployed(), clu.Config.AllNodes(), 30*time.Second)

	// use an access node to create the chainClient
	chClient := newWalletWithFunds(e, 5, 0, 2, 4, 5, 7)

	// send off-ledger request via Web API (to the access node)
	_, err = chClient.PostOffLedgerRequest(context.Background(),
		nativeIncCounterSCHname,
		inccounter.FuncIncCounter.Hname(),
	)
	require.NoError(t, err)

	waitUntil(t, e.counterEquals(43), clu.Config.AllNodes(), 30*time.Second)

	// check off-ledger request was successfully processed (check by asking another access node)
	ret, err := apiextensions.CallView(
		context.Background(),
		clu.WaspClient(6),
		e.Chain.ChainID.String(),
		apiclient.ContractCallViewRequest{
			ContractHName: nativeIncCounterSCHname.String(),
			FunctionHName: inccounter.ViewGetCounter.Hname().String(),
		})

	require.NoError(t, err)
	resultint64, _ := codec.DecodeInt64(ret.Get(inccounter.VarCounter))
	require.EqualValues(t, 43, resultint64)
}

// executed in cluster_test.go
func testOffledgerRequest(t *testing.T, e *ChainEnv) {
	e.deployNativeIncCounterSC()

	chClient := newWalletWithFunds(e, 0, 0, 1, 2, 3)

	// send off-ledger request via Web API
	offledgerReq, err := chClient.PostOffLedgerRequest(context.Background(),
		nativeIncCounterSCHname,
		inccounter.FuncIncCounter.Hname(),
	)
	require.NoError(t, err)
	_, err = e.Chain.CommitteeMultiClient().WaitUntilRequestProcessedSuccessfully(e.Chain.ChainID, offledgerReq.ID(), false, 30*time.Second)
	require.NoError(t, err)

	ret, err := apiextensions.CallView(
		context.Background(),
		e.Chain.Cluster.WaspClient(0),
		e.Chain.ChainID.String(),
		apiclient.ContractCallViewRequest{
			ContractHName: nativeIncCounterSCHname.String(),
			FunctionHName: inccounter.ViewGetCounter.Hname().String(),
		})

	require.NoError(t, err)
	resultint64, err := codec.DecodeInt64(ret.Get(inccounter.VarCounter))
	require.NoError(t, err)
	require.EqualValues(t, 43, resultint64)
}

// executed in cluster_test.go
func testOffledgerRequest900KB(t *testing.T, e *ChainEnv) {
	chClient := newWalletWithFunds(e, 0, 0, 0, 1, 2, 3)

	// send big blob off-ledger request via Web API
	size := int64(1 * 900 * 1024) // 900 KB
	randomData := make([]byte, size)
	_, err := rand.Read(randomData)
	require.NoError(t, err)

	paramsDict := dict.Dict{"data": randomData}
	expectedHash := blob.MustGetBlobHash(paramsDict)

	// raise gas limits, gas cost for 900KB has exceeded the limits
	{
		limits1 := *gas.LimitsDefault
		limits1.MaxGasPerRequest = 10 * limits1.MaxGasPerRequest
		limits1.MaxGasExternalViewCall = 10 * limits1.MaxGasExternalViewCall
		govClient := e.Chain.SCClient(governance.Contract.Hname(), e.Chain.OriginatorKeyPair)
		gasLimitsReq, err1 := govClient.PostOffLedgerRequest(
			governance.FuncSetGasLimits.Name,
			chainclient.PostRequestParams{
				Args: dict.Dict{
					governance.ParamGasLimitsBytes: limits1.Bytes(),
				},
			},
		)
		require.NoError(t, err1)
		_, _, err = e.Clu.WaspClient(0).ChainsApi.
			WaitForRequest(context.Background(), e.Chain.ChainID.String(), gasLimitsReq.ID().String()).
			TimeoutSeconds(10).
			Execute()
		require.NoError(t, err)

		retDict, err1 := govClient.CallView(context.Background(),
			governance.ViewGetGasLimits.Name,
			dict.Dict{},
		)
		require.NoError(t, err1)
		limits2, err1 := gas.LimitsFromBytes(retDict.Get(governance.ParamGasLimitsBytes))
		require.Equal(t, limits1, *limits2)
		require.NoError(t, err1)
	}

	offledgerReq, err := chClient.PostOffLedgerRequest(context.Background(),
		blob.Contract.Hname(),
		blob.FuncStoreBlob.Hname(),
		chainclient.PostRequestParams{
			Args:      paramsDict,
			Allowance: isc.NewAssetsBaseTokens(10 * isc.Million),
		},
	)
	require.NoError(t, err)

	_, err = e.Chain.CommitteeMultiClient().
		WaitUntilRequestProcessedSuccessfully(e.Chain.ChainID, offledgerReq.ID(), false, 30*time.Second)
	require.NoError(t, err)

	// ensure blob was stored by the cluster
	res, _, err := e.Chain.Cluster.WaspClient(2).CorecontractsApi.
		BlobsGetBlobValue(context.Background(), e.Chain.ChainID.String(), expectedHash.Hex(), "data").
		Execute()
	require.NoError(t, err)

	binaryData, err := iotago.DecodeHex(res.ValueData)
	require.NoError(t, err)
	require.EqualValues(t, binaryData, randomData)
}

// executed in cluster_test.go
func testOffledgerNonce(t *testing.T, e *ChainEnv) {
	e.deployNativeIncCounterSC()

	chClient := newWalletWithFunds(e, 0, 0, 1, 2, 3)

	// send off-ledger request with a high nonce
	offledgerReq, err := chClient.PostOffLedgerRequest(context.Background(),
		nativeIncCounterSCHname,
		inccounter.FuncIncCounter.Hname(),
		chainclient.PostRequestParams{
			Nonce: 1_000_000,
		},
	)
	require.NoError(t, err)
	_, err = e.Chain.CommitteeMultiClient().WaitUntilRequestProcessedSuccessfully(e.Chain.ChainID, offledgerReq.ID(), false, 5*time.Second)
	require.Error(t, err) // wont' be processed

	// send off-ledger requests with the correct nonce
	for i := uint64(0); i < 5; i++ {
		req, err2 := chClient.PostOffLedgerRequest(context.Background(),
			nativeIncCounterSCHname,
			inccounter.FuncIncCounter.Hname(),
			chainclient.PostRequestParams{
				Nonce: i,
			},
		)
		require.NoError(t, err2)
		_, err2 = e.Chain.CommitteeMultiClient().WaitUntilRequestProcessedSuccessfully(e.Chain.ChainID, req.ID(), false, 10*time.Second)
		require.NoError(t, err2)
	}

	// try replaying an older nonce
	_, err = chClient.PostOffLedgerRequest(context.Background(),
		accounts.Contract.Hname(),
		accounts.FuncTransferAccountToChain.Hname(),
		chainclient.PostRequestParams{
			Nonce: 1,
		},
	)
	require.Error(t, err)
	require.Contains(t, string(err.(*apiclient.GenericOpenAPIError).Body()), "not added to the mempool")
}

func newWalletWithFunds(e *ChainEnv, waspnode int, waitOnNodes ...int) *chainclient.Client {
	baseTokes := 1000 * isc.Million
	userWallet, userAddress, err := e.Clu.NewKeyPairWithFunds()
	require.NoError(e.t, err)
	userAgentID := isc.NewAgentID(userAddress)

	chClient := chainclient.New(e.Clu.L1Client(), e.Clu.WaspClient(waspnode), e.Chain.ChainID, userWallet)

	// deposit funds before sending the off-ledger requestargs
	reqTx, err := chClient.Post1Request(accounts.Contract.Hname(), accounts.FuncDeposit.Hname(), chainclient.PostRequestParams{
		Transfer: isc.NewAssetsBaseTokens(baseTokes),
	})
	require.NoError(e.t, err)

	receipts, err := e.Chain.CommitteeMultiClient().WaitUntilAllRequestsProcessedSuccessfully(e.Chain.ChainID, reqTx, false, 30*time.Second)
	require.NoError(e.t, err)

	gasFeeCharged, err := iotago.DecodeUint64(receipts[0].GasFeeCharged)
	require.NoError(e.t, err)

	expectedBaseTokens := baseTokes - gasFeeCharged
	e.checkBalanceOnChain(userAgentID, isc.BaseTokenID, expectedBaseTokens)

	// wait until access node syncs with account
	if len(waitOnNodes) > 0 {
		waitUntil(e.t, e.accountExists(userAgentID), waitOnNodes, 10*time.Second)
	}
	return chClient
}
