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
)

func TestOffledgerRequestAccessNode(t *testing.T) {
	const clusterSize = 10
	clu := newCluster(t, waspClusterOpts{nNodes: clusterSize})

	cmt := []int{0, 1, 2, 3}

	addr, err := clu.RunDKG(cmt, 3)
	require.NoError(t, err)

	chain, err := clu.DeployChain("chain", clu.Config.AllNodes(), cmt, 3, addr)
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
	ret, err := apiextensions.CallView(context.Background(), clu.WaspClient(6), apiclient.ContractCallViewRequest{
		ChainId:       e.Chain.ChainID.String(),
		ContractHName: nativeIncCounterSCHname.String(),
		FunctionHName: inccounter.ViewGetCounter.Hname().String(),
	})

	require.NoError(t, err)
	resultint64, _ := codec.DecodeInt64(ret.MustGet(inccounter.VarCounter))
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
	_, err = e.Chain.CommitteeMultiClient().WaitUntilRequestProcessedSuccessfully(e.Chain.ChainID, offledgerReq.ID(), 30*time.Second)
	require.NoError(t, err)

	ret, err := apiextensions.CallView(context.Background(), e.Chain.Cluster.WaspClient(0), apiclient.ContractCallViewRequest{
		ChainId:       e.Chain.ChainID.String(),
		ContractHName: nativeIncCounterSCHname.String(),
		FunctionHName: inccounter.ViewGetCounter.Hname().String(),
	})

	require.NoError(t, err)
	resultint64, err := codec.DecodeInt64(ret.MustGet(inccounter.VarCounter))
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

	offledgerReq, err := chClient.PostOffLedgerRequest(context.Background(),
		blob.Contract.Hname(),
		blob.FuncStoreBlob.Hname(),
		chainclient.PostRequestParams{
			Args: paramsDict,
		})
	require.NoError(t, err)

	_, err = e.Chain.CommitteeMultiClient().WaitUntilRequestProcessedSuccessfully(e.Chain.ChainID, offledgerReq.ID(), 30*time.Second)
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
	_, err = e.Chain.CommitteeMultiClient().WaitUntilRequestProcessedSuccessfully(e.Chain.ChainID, offledgerReq.ID(), 30*time.Second)
	require.NoError(t, err)

	// send off-ledger request with a high nonce -1
	offledgerReq, err = chClient.PostOffLedgerRequest(context.Background(),
		nativeIncCounterSCHname,
		inccounter.FuncIncCounter.Hname(),
		chainclient.PostRequestParams{
			Nonce: 999_999,
		},
	)
	require.NoError(t, err)
	_, err = e.Chain.CommitteeMultiClient().WaitUntilRequestProcessedSuccessfully(e.Chain.ChainID, offledgerReq.ID(), 30*time.Second)
	require.NoError(t, err)

	// send off-ledger request with a much lower nonce
	_, err = chClient.PostOffLedgerRequest(context.Background(),
		nativeIncCounterSCHname,
		inccounter.FuncIncCounter.Hname(),
		chainclient.PostRequestParams{
			Nonce: 1,
		},
	)

	apiError, ok := apiextensions.AsAPIError(err)
	require.True(t, ok)
	require.NotNil(t, apiError.DetailError)
	require.Regexp(t, "invalid nonce", apiError.DetailError.Error)

	// try replaying the initial request
	_, err = chClient.PostOffLedgerRequest(context.Background(),
		nativeIncCounterSCHname,
		inccounter.FuncIncCounter.Hname(),
		chainclient.PostRequestParams{
			Nonce: 1_000_000,
		},
	)

	apiError, ok = apiextensions.AsAPIError(err)
	require.True(t, ok)
	require.NotNil(t, apiError.DetailError)
	require.Regexp(t, "request already processed", apiError.DetailError.Error)

	// send a request with a higher nonce
	offledgerReq, err = chClient.PostOffLedgerRequest(context.Background(),
		nativeIncCounterSCHname,
		inccounter.FuncIncCounter.Hname(),
		chainclient.PostRequestParams{
			Nonce: 1_000_001,
		},
	)
	require.NoError(t, err)
	_, err = e.Chain.CommitteeMultiClient().WaitUntilRequestProcessedSuccessfully(e.Chain.ChainID, offledgerReq.ID(), 30*time.Second)
	require.NoError(t, err)
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

	receipts, err := e.Chain.CommitteeMultiClient().WaitUntilAllRequestsProcessedSuccessfully(e.Chain.ChainID, reqTx, 30*time.Second)
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
