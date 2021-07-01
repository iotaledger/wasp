package tests

import (
	"crypto/rand"
	"testing"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/client/chainclient"
	"github.com/iotaledger/wasp/contracts/native/inccounter"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/requestargs"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
	"github.com/iotaledger/wasp/tools/cluster"
	clutest "github.com/iotaledger/wasp/tools/cluster/testutil"
	"github.com/stretchr/testify/require"
)

func newWalletWithFunds(t *testing.T, clust *cluster.Cluster, ch *cluster.Chain, waspnode int, seedN, iotas uint64) *chainclient.Client {
	userWallet := wallet.KeyPair(seedN)
	userAddress := ledgerstate.NewED25519Address(userWallet.PublicKey)
	userAgentID := coretypes.NewAgentID(userAddress, 0)

	chClient := chainclient.New(clust.GoshimmerClient(), clust.WaspClient(waspnode), ch.ChainID, userWallet)

	// deposit funds before sending the off-ledger requestargs
	err = requestFunds(clust, userAddress, "userWallet")
	check(err, t)
	reqTx, err := chClient.Post1Request(accounts.Interface.Hname(), coretypes.Hn(accounts.FuncDeposit), chainclient.PostRequestParams{
		Transfer: coretypes.NewTransferIotas(iotas),
	})
	check(err, t)
	err = ch.CommitteeMultiClient().WaitUntilAllRequestsProcessed(ch.ChainID, reqTx, 30*time.Second)
	check(err, t)
	checkBalanceOnChain(t, ch, userAgentID, ledgerstate.ColorIOTA, iotas)

	return chClient
}

func TestOffledgerRequest(t *testing.T) {
	setup(t, "test_cluster")

	counter1, err := clu.StartMessageCounter(map[string]int{
		"dismissed_committee": 0,
		"state":               2,
		"request_out":         1,
	})
	check(err, t)
	defer counter1.Close()

	chain1, err := clu.DeployDefaultChain()
	check(err, t)
	deployIncCounterSC(t, chain1, counter1)

	chClient := newWalletWithFunds(t, clu, chain1, 0, 1, 100)

	// send off-ledger request via Web API
	offledgerReq, err := chClient.PostOffLedgerRequest(incCounterSCHname, coretypes.Hn(inccounter.FuncIncCounter))
	check(err, t)
	err = chain1.CommitteeMultiClient().WaitUntilRequestProcessed(&chain1.ChainID, offledgerReq.ID(), 30*time.Second)
	check(err, t)

	// check off-ledger request was successfully processed
	ret, err := chain1.Cluster.WaspClient(0).CallView(
		chain1.ChainID, incCounterSCHname, inccounter.FuncGetCounter,
	)
	check(err, t)
	result, _ := ret.Get(inccounter.VarCounter)
	resultint64, _, _ := codec.DecodeInt64(result)
	require.EqualValues(t, 43, resultint64)
}

func TestOffledgerRequest1Mb(t *testing.T) {
	setup(t, "test_cluster")

	counter1, err := clu.StartMessageCounter(map[string]int{
		"dismissed_committee": 0,
		"state":               2,
		"request_out":         1,
	})
	check(err, t)
	defer counter1.Close()

	chain1, err := clu.DeployDefaultChain()
	check(err, t)

	chClient := newWalletWithFunds(t, clu, chain1, 0, 1, 100)

	// send big blob off-ledger request via Web API
	size := int64(1 * 1024 * 1024) // 1 MB
	randomData := make([]byte, size)
	_, err = rand.Read(randomData)
	check(err, t)

	paramsDict := dict.Dict{"data": randomData}
	expectedHash := blob.MustGetBlobHash(paramsDict)

	offledgerReq, err := chClient.PostOffLedgerRequest(
		blob.Interface.Hname(),
		coretypes.Hn(blob.FuncStoreBlob),
		chainclient.PostRequestParams{
			Args: requestargs.New().AddEncodeSimpleMany(paramsDict),
		})
	check(err, t)

	err = chain1.CommitteeMultiClient().WaitUntilRequestProcessed(&chain1.ChainID, offledgerReq.ID(), 30*time.Second)
	check(err, t)

	// ensure blob was stored by the cluster
	res, err := chain1.Cluster.WaspClient(2).CallView(
		chain1.ChainID, blob.Interface.Hname(), blob.FuncGetBlobField,
		dict.Dict{
			blob.ParamHash:  expectedHash[:],
			blob.ParamField: []byte("data"),
		})
	check(err, t)
	binaryData, err := res.Get(blob.ParamBytes)
	check(err, t)
	require.EqualValues(t, binaryData, randomData)
}

func TestOffledgerRequestAccessNode(t *testing.T) {
	clu1 := clutest.NewCluster(t, 10)

	cmt1 := []int{0, 1, 2, 3}

	addr1, err := clu1.RunDKG(cmt1, 3)
	require.NoError(t, err)

	chain1, err := clu1.DeployChain("chain", clu1.Config.AllNodes(), cmt1, 3, addr1)
	require.NoError(t, err)

	deployIncCounterSC(t, chain1, nil)

	waitUntil(t, createCheckContractDeployedFn(chain1, incCounterSCName), makeRange(0, 9), 30*time.Second)

	// use an access node to create the chainClient
	chClient := newWalletWithFunds(t, clu1, chain1, 5, 1, 100)

	// send off-ledger request via Web API (to the access node)
	_, err = chClient.PostOffLedgerRequest(incCounterSCHname, coretypes.Hn(inccounter.FuncIncCounter))
	check(err, t)

	waitUntil(t, createCheckCounterFn(chain1, 43), []int{0, 1, 2, 3, 6}, 30*time.Second)

	// check off-ledger request was successfully processed (check by asking another access node)
	ret, err := clu1.WaspClient(6).CallView(
		chain1.ChainID, incCounterSCHname, inccounter.FuncGetCounter,
	)
	check(err, t)
	result, _ := ret.Get(inccounter.VarCounter)
	resultint64, _, _ := codec.DecodeInt64(result)
	require.EqualValues(t, 43, resultint64)
}
