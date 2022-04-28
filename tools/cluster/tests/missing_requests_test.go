package tests

import (
	"testing"
	"time"

	"github.com/iotaledger/wasp/client/chainclient"
	"github.com/iotaledger/wasp/contracts/native/inccounter"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/tools/cluster/templates"
	"github.com/stretchr/testify/require"
)

func TestMissingRequests(t *testing.T) {
	// disable offledger request gossip between nodes
	modifyConfig := func(nodeIndex int, configParams *templates.WaspConfigParams) *templates.WaspConfigParams {
		configParams.OffledgerBroadcastUpToNPeers = 0
		return configParams
	}
	clu := newCluster(t, waspClusterOpts{nNodes: 4, modifyConfig: modifyConfig})
	cmt := []int{0, 1, 2, 3}
	addr, err := clu.RunDKG(cmt, 4)
	require.NoError(t, err)

	chain, err := clu.DeployChain("chain", clu.Config.AllNodes(), cmt, 4, addr)
	require.NoError(t, err)
	chainID := chain.ChainID

	e := newChainEnv(t, clu, chain)

	e.deployIncCounterSC(nil)

	waitUntil(t, e.contractIsDeployed(incCounterSCName), clu.Config.AllNodes(), 30*time.Second)

	userWallet, userAddress, err := e.clu.NewKeyPairWithFunds()
	require.NoError(t, err)

	// deposit funds before sending the off-ledger request
	e.requestFunds(userAddress, "userWallet")
	chClient := chainclient.New(clu.L1Client(), clu.WaspClient(0), chainID, userWallet)
	reqTx, err := chClient.DepositFunds(100)
	require.NoError(t, err)
	_, err = chain.CommitteeMultiClient().WaitUntilAllRequestsProcessedSuccessfully(chainID, reqTx, 30*time.Second)
	require.NoError(t, err)

	// send off-ledger request to all nodes except #3
	req := iscp.NewOffLedgerRequest(chainID, incCounterSCHname, inccounter.FuncIncCounter.Hname(), dict.Dict{}, 0)
	req.Sign(userWallet)

	err = clu.WaspClient(0).PostOffLedgerRequest(chainID, req)
	require.NoError(t, err)
	err = clu.WaspClient(1).PostOffLedgerRequest(chainID, req)
	require.NoError(t, err)

	// TODO try to send to only 2 nodes
	err = clu.WaspClient(2).PostOffLedgerRequest(chainID, req)
	require.NoError(t, err)
	// err = clu1.WaspClient(3).PostOffLedgerRequest(&chainID, req)
	// require.NoError(t, err)

	//------
	// send a dummy request to node #3, so that it proposes a batch and the consensus hang is broken
	req2 := iscp.NewOffLedgerRequest(chainID, iscp.Hn("foo"), iscp.Hn("bar"), nil, 1)
	req2.Sign(userWallet)
	err = clu.WaspClient(3).PostOffLedgerRequest(chainID, req2)
	require.NoError(t, err)
	//-------

	// expect request to be successful, as node #3 must ask for the missing request from other nodes
	waitUntil(t, e.counterEquals(43), clu.Config.AllNodes(), 30*time.Second)
}
