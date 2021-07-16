package tests

import (
	"testing"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/client/chainclient"
	"github.com/iotaledger/wasp/contracts/native/inccounter"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/request"
	"github.com/iotaledger/wasp/packages/iscp/requestargs"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/tools/cluster/templates"
	clutest "github.com/iotaledger/wasp/tools/cluster/testutil"
	"github.com/stretchr/testify/require"
)

func TestMissingRequests(t *testing.T) {
	// disable offledger request gossip between nodes
	modifyConfig := func(nodeIndex int, configParams *templates.WaspConfigParams) *templates.WaspConfigParams {
		configParams.OffledgerBroadcastUpToNPeers = 0
		return configParams
	}
	clu1 := clutest.NewCluster(t, 4, nil, modifyConfig)
	cmt1 := []int{0, 1, 2, 3}
	addr1, err := clu1.RunDKG(cmt1, 4)
	require.NoError(t, err)

	chain1, err := clu1.DeployChain("chain", clu1.Config.AllNodes(), cmt1, 4, addr1)
	require.NoError(t, err)
	chainID := chain1.ChainID

	deployIncCounterSC(t, chain1, nil)

	waitUntil(t, contractIsDeployed(chain1, incCounterSCName), clu1.Config.AllNodes(), 30*time.Second)

	userWallet := wallet.KeyPair(0)
	userAddress := ledgerstate.NewED25519Address(userWallet.PublicKey)

	// deposit funds before sending the off-ledger request
	err = requestFunds(clu1, userAddress, "userWallet")
	check(err, t)
	chClient := chainclient.New(clu1.GoshimmerClient(), clu1.WaspClient(0), chainID, userWallet)
	reqTx, err := chClient.Post1Request(accounts.Contract.Hname(), accounts.FuncDeposit.Hname(), chainclient.PostRequestParams{
		Transfer: iscp.NewTransferIotas(100),
	})
	check(err, t)
	err = chain1.CommitteeMultiClient().WaitUntilAllRequestsProcessed(chainID, reqTx, 30*time.Second)
	check(err, t)

	// send off-ledger request to all nodes except #3
	req := request.NewRequestOffLedger(incCounterSCHname, inccounter.FuncIncCounter.Hname(), requestargs.RequestArgs{}) //.WithTransfer(par.Transfer)
	req.Sign(userWallet)

	err = clu1.WaspClient(0).PostOffLedgerRequest(&chainID, req)
	check(err, t)
	err = clu1.WaspClient(1).PostOffLedgerRequest(&chainID, req)
	check(err, t)

	// TODO try to send to only 2 nodes
	err = clu1.WaspClient(2).PostOffLedgerRequest(&chainID, req)
	check(err, t)
	// err = clu1.WaspClient(3).PostOffLedgerRequest(&chainID, req)
	// check(err, t)

	//------
	// send a dummy request to node #3, so that it proposes a batch and the consensus hang is broken
	req2 := request.NewRequestOffLedger(iscp.Hn("foo"), iscp.Hn("bar"), nil)
	req2.Sign(userWallet)
	err = clu1.WaspClient(3).PostOffLedgerRequest(&chainID, req2)
	check(err, t)
	//-------

	// expect request to be successful, as node #3 must ask for the missing request from other nodes
	waitUntil(t, counterEquals(chain1, 43), clu1.Config.AllNodes(), 30*time.Second)
}
