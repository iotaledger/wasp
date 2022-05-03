package tests

import (
	"crypto/rand"
	"testing"
	"time"

	"github.com/iotaledger/wasp/client/chainclient"
	"github.com/iotaledger/wasp/contracts/native/inccounter"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
	"github.com/stretchr/testify/require"
)

func (e *chainEnv) newWalletWithFunds(waspnode int, seedN, iotas uint64, waitOnNodes ...int) *chainclient.Client {
	userWallet, userAddress, err := e.clu.NewKeyPairWithFunds()
	require.NoError(e.t, err)
	userAgentID := iscp.NewAgentID(userAddress, 0)

	chClient := chainclient.New(e.clu.L1Client(), e.clu.WaspClient(waspnode), e.chain.ChainID, userWallet)

	// deposit funds before sending the off-ledger requestargs
	reqTx, err := chClient.Post1Request(accounts.Contract.Hname(), accounts.FuncDeposit.Hname(), chainclient.PostRequestParams{
		Transfer: iscp.NewTokensIotas(iotas),
	})
	require.NoError(e.t, err)
	receipts, err := e.chain.CommitteeMultiClient().WaitUntilAllRequestsProcessedSuccessfully(e.chain.ChainID, reqTx, 30*time.Second)
	require.NoError(e.t, err)
	expectedIotas := iotas - receipts[0].GasFeeCharged
	e.checkBalanceOnChain(userAgentID, iscp.IotaTokenID, expectedIotas)

	// wait until access node syncs with account
	if len(waitOnNodes) > 0 {
		waitUntil(e.t, e.accountExists(userAgentID), waitOnNodes, 10*time.Second)
	}
	return chClient
}

func TestOffledgerRequest(t *testing.T) {
	e := setupWithNoChain(t)

	counter, err := e.clu.StartMessageCounter(map[string]int{
		"dismissed_committee": 0,
		"request_out":         1,
	})
	require.NoError(t, err)
	defer counter.Close()

	chain, err := e.clu.DeployDefaultChain()
	require.NoError(t, err)

	chEnv := newChainEnv(t, e.clu, chain)
	chEnv.deployIncCounterSC(counter)

	chClient := chEnv.newWalletWithFunds(0, 1, 1000, 0, 1, 2, 3)

	// send off-ledger request via Web API
	offledgerReq, err := chClient.PostOffLedgerRequest(
		incCounterSCHname,
		inccounter.FuncIncCounter.Hname(),
	)
	require.NoError(t, err)
	_, err = chain.CommitteeMultiClient().WaitUntilRequestProcessedSuccessfully(chain.ChainID, offledgerReq.ID(), 30*time.Second)
	require.NoError(t, err)

	// check off-ledger request was successfully processed
	ret, err := chain.Cluster.WaspClient(0).CallView(
		chain.ChainID, incCounterSCHname, inccounter.FuncGetCounter.Name, nil,
	)
	require.NoError(t, err)
	resultint64, err := codec.DecodeInt64(ret.MustGet(inccounter.VarCounter))
	require.NoError(t, err)
	require.EqualValues(t, 43, resultint64)
}

func TestOffledgerRequest900KB(t *testing.T) {
	e := setupWithNoChain(t)

	var err error
	counter, err := e.clu.StartMessageCounter(map[string]int{
		"dismissed_committee": 0,
		"state":               2,
		"request_out":         1,
	})
	require.NoError(t, err)
	defer counter.Close()

	chain, err := e.clu.DeployDefaultChain()
	require.NoError(t, err)

	chEnv := newChainEnv(t, e.clu, chain)

	chClient := chEnv.newWalletWithFunds(0, 1, 10000, 0, 1, 2, 3)

	// send big blob off-ledger request via Web API
	size := int64(1 * 900 * 1024) // 900 KB
	randomData := make([]byte, size)
	_, err = rand.Read(randomData)
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

	_, err = chain.CommitteeMultiClient().WaitUntilRequestProcessedSuccessfully(chain.ChainID, offledgerReq.ID(), 30*time.Second)
	require.NoError(t, err)

	// ensure blob was stored by the cluster
	res, err := chain.Cluster.WaspClient(2).CallView(
		chain.ChainID, blob.Contract.Hname(), blob.ViewGetBlobField.Name,
		dict.Dict{
			blob.ParamHash:  expectedHash[:],
			blob.ParamField: []byte("data"),
		})
	require.NoError(t, err)
	binaryData, err := res.Get(blob.ParamBytes)
	require.NoError(t, err)
	require.EqualValues(t, binaryData, randomData)
}

func TestOffledgerRequestAccessNode(t *testing.T) {
	const clusterSize = 10
	clu := newCluster(t, waspClusterOpts{nNodes: clusterSize})

	cmt := []int{0, 1, 2, 3}

	addr, err := clu.RunDKG(cmt, 3)
	require.NoError(t, err)

	chain, err := clu.DeployChain("chain", clu.Config.AllNodes(), cmt, 3, addr)
	require.NoError(t, err)

	e := newChainEnv(t, clu, chain)

	e.deployIncCounterSC(nil)

	waitUntil(t, e.contractIsDeployed(incCounterSCName), clu.Config.AllNodes(), 30*time.Second)

	// use an access node to create the chainClient
	chClient := e.newWalletWithFunds(5, 1, 1000, 0, 1, 2, 3, 4, 5)

	// send off-ledger request via Web API (to the access node)
	_, err = chClient.PostOffLedgerRequest(
		incCounterSCHname,
		inccounter.FuncIncCounter.Hname(),
	)
	require.NoError(t, err)

	waitUntil(t, e.counterEquals(43), clu.Config.AllNodes(), 30*time.Second)

	// check off-ledger request was successfully processed (check by asking another access node)
	ret, err := clu.WaspClient(6).CallView(
		chain.ChainID, incCounterSCHname, inccounter.FuncGetCounter.Name, nil,
	)
	require.NoError(t, err)
	resultint64, _ := codec.DecodeInt64(ret.MustGet(inccounter.VarCounter))
	require.EqualValues(t, 43, resultint64)
}
