package tests

import (
	"crypto/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/client/chainclient"
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
	_, err = chClient.PostOffLedgerRequest(
		nativeIncCounterSCHname,
		inccounter.FuncIncCounter.Hname(),
	)
	require.NoError(t, err)

	waitUntil(t, e.counterEquals(43), clu.Config.AllNodes(), 30*time.Second)

	// check off-ledger request was successfully processed (check by asking another access node)
	ret, err := clu.WaspClient(6).CallView(
		e.Chain.ChainID, nativeIncCounterSCHname, inccounter.ViewGetCounter.Name, nil,
	)
	require.NoError(t, err)
	resultint64, _ := codec.DecodeInt64(ret.MustGet(inccounter.VarCounter))
	require.EqualValues(t, 43, resultint64)
}

// executed in cluster_test.go
func testOffledgerRequest(t *testing.T, e *ChainEnv) {
	e.deployNativeIncCounterSC()

	chClient := newWalletWithFunds(e, 0, 0, 1, 2, 3)

	// send off-ledger request via Web API
	offledgerReq, err := chClient.PostOffLedgerRequest(
		nativeIncCounterSCHname,
		inccounter.FuncIncCounter.Hname(),
	)
	require.NoError(t, err)
	_, err = e.Chain.CommitteeMultiClient().WaitUntilRequestProcessedSuccessfully(e.Chain.ChainID, offledgerReq.ID(), 30*time.Second)
	require.NoError(t, err)

	// check off-ledger request was successfully processed
	ret, err := e.Chain.Cluster.WaspClient(0).CallView(
		e.Chain.ChainID, nativeIncCounterSCHname, inccounter.ViewGetCounter.Name, nil,
	)
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

	offledgerReq, err := chClient.PostOffLedgerRequest(
		blob.Contract.Hname(),
		blob.FuncStoreBlob.Hname(),
		chainclient.PostRequestParams{
			Args: paramsDict,
		})
	require.NoError(t, err)

	_, err = e.Chain.CommitteeMultiClient().WaitUntilRequestProcessedSuccessfully(e.Chain.ChainID, offledgerReq.ID(), 30*time.Second)
	require.NoError(t, err)

	// ensure blob was stored by the cluster
	res, err := e.Chain.Cluster.WaspClient(2).CallView(
		e.Chain.ChainID, blob.Contract.Hname(), blob.ViewGetBlobField.Name,
		dict.Dict{
			blob.ParamHash:  expectedHash[:],
			blob.ParamField: []byte("data"),
		})
	require.NoError(t, err)
	binaryData, err := res.Get(blob.ParamBytes)
	require.NoError(t, err)
	require.EqualValues(t, binaryData, randomData)
}

// executed in cluster_test.go
func testOffledgerNonce(t *testing.T, e *ChainEnv) {
	e.deployNativeIncCounterSC()

	chClient := newWalletWithFunds(e, 0, 0, 1, 2, 3)

	// send off-ledger request with a high nonce
	offledgerReq, err := chClient.PostOffLedgerRequest(
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
	offledgerReq, err = chClient.PostOffLedgerRequest(
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
	_, err = chClient.PostOffLedgerRequest(
		nativeIncCounterSCHname,
		inccounter.FuncIncCounter.Hname(),
		chainclient.PostRequestParams{
			Nonce: 1,
		},
	)
	require.Regexp(t, "invalid nonce", err.Error())

	// try replaying the initial request
	_, err = chClient.PostOffLedgerRequest(
		nativeIncCounterSCHname,
		inccounter.FuncIncCounter.Hname(),
		chainclient.PostRequestParams{
			Nonce: 1_000_000,
		},
	)
	require.Regexp(t, "request already processed", err.Error())

	// send a request with a higher nonce
	offledgerReq, err = chClient.PostOffLedgerRequest(
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
	expectedBaseTokens := baseTokes - receipts[0].GasFeeCharged
	e.checkBalanceOnChain(userAgentID, isc.BaseTokenID, expectedBaseTokens)

	// wait until access node syncs with account
	if len(waitOnNodes) > 0 {
		waitUntil(e.t, e.accountExists(userAgentID), waitOnNodes, 10*time.Second)
	}
	return chClient
}
