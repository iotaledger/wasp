package tests

import (
	"testing"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/client/chainclient"
	"github.com/iotaledger/wasp/contracts/native/inccounter"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/stretchr/testify/require"
)

func TestOffledgerRequests(t *testing.T) {
	setup(t, "test_cluster")

	counter, err := clu.StartMessageCounter(map[string]int{
		"dismissed_committee": 0,
		"state":               2,
		"request_out":         1,
	})
	check(err, t)
	defer counter.Close()

	chain1, err := clu.DeployDefaultChain()
	check(err, t)

	scHname := coretypes.Hn("inncounter1")
	deployIncCounterSC(t, chain1, counter)

	userWallet := wallet.KeyPair(1)
	userAddress := ledgerstate.NewED25519Address(userWallet.PublicKey)
	userAgentID := coretypes.NewAgentID(myAddress, 0)

	chClient := chainclient.New(clu.GoshimmerClient(), clu.WaspClient(0), chain1.ChainID, userWallet)

	// deposit funds before sending the off-ledger requestargs
	err = requestFunds(clu, userAddress, "userWallet")
	check(err, t)
	reqTx, err := chClient.Post1Request(accounts.Interface.Hname(), coretypes.Hn(accounts.FuncDeposit), chainclient.PostRequestParams{
		Transfer: coretypes.NewTransferIotas(100),
	})
	check(err, t)
	err = chain1.CommitteeMultiClient().WaitUntilAllRequestsProcessed(chain1.ChainID, reqTx, 30*time.Second)
	check(err, t)
	checkBalanceOnChain(t, chain1, userAgentID, ledgerstate.ColorIOTA, 100)

	// send off-ledger request via Web API
	offledgerReq, err := chClient.PostOffLedgerRequest(scHname, coretypes.Hn(inccounter.FuncIncCounter))
	check(err, t)
	err = chain1.CommitteeMultiClient().WaitUntilRequestProcessed(&chain1.ChainID, offledgerReq.ID(), 30*time.Second)
	check(err, t)

	// check off-ledger request was successfully processed
	ret, err := chain1.Cluster.WaspClient(0).CallView(
		chain1.ChainID, scHname, inccounter.FuncGetCounter,
	)
	check(err, t)
	result, _ := ret.Get(inccounter.VarCounter)
	resultint64, _, _ := codec.DecodeInt64(result)
	require.EqualValues(t, 43, resultint64)
}

// TODO add a test with an access node that is not party of the committee
