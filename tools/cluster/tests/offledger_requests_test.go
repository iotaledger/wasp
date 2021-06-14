package tests

import (
	"testing"
	"time"

	"github.com/iotaledger/wasp/client/chainclient"
	"github.com/iotaledger/wasp/contracts/native/inccounter"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/requestargs"
	"github.com/iotaledger/wasp/packages/kv/codec"
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

	chain, err := clu.DeployDefaultChain()
	check(err, t)

	scHname := coretypes.Hn("inncounter1")
	deployIncCounterSC(t, chain, counter)

	// send off-ledger request via Web API

	chClient := chainclient.New(clu.GoshimmerClient(), clu.WaspClient(0), chain.ChainID, chain.OriginatorKeyPair())

	offledgerReq, err := chClient.PostOffLedgerRequest(scHname, coretypes.Hn(inccounter.FuncIncCounter), requestargs.RequestArgs{})
	check(err, t)
	err = chain.CommitteeMultiClient().WaitUntilRequestProcessed(&chain.ChainID, offledgerReq.ID(), 30*time.Second)
	check(err, t)

	// err = chain.CommitteeMultiClient().WaitUntilAllRequestsProcessed(chain.ChainID, reqTx, 30*time.Second)
	// or
	// waspclient .WaitUntilRequestProcessed(&chainID, coretypes.RequestID(out.ID())

	// check off-ledger request was successful
	ret, err := chain.Cluster.WaspClient(0).CallView(
		chain.ChainID, scHname, inccounter.FuncGetCounter,
	)
	check(err, t)
	result, _ := ret.Get(inccounter.VarCounter)
	resultint64, _, _ := codec.DecodeInt64(result)
	require.EqualValues(t, 43, resultint64)

	// chain.WithSCState(hname, func(host string, blockIndex uint32, state dict.Dict) bool {
	// 	counterValue, _, _ := codec.DecodeInt64(state.MustGet(inccounter.VarCounter))
	// 	require.EqualValues(t, 42, counterValue)
	// 	return true
	// })
}

func TestOffledgerRequestsWithAccessNode(t *testing.T) {

	setup(t, "test_cluster")

	counter, err := clu.StartMessageCounter(map[string]int{
		"dismissed_committee": 0,
		"state":               2,
		"request_out":         1,
	})
	check(err, t)
	defer counter.Close()

	nodes := clu.Config.AllNodes()
	committeeNodes := nodes[0:3]
	accessNodes := nodes[3:]
	minQuorum := len(committeeNodes)/2 + 1
	quorum := len(committeeNodes) * 3 / 4
	if quorum < minQuorum {
		quorum = minQuorum
	}
	// deploy custom chain with 3 committee nodes and 1 access node
	chain, err := clu.DeployChain("3 committee nodes", committeeNodes, uint16(quorum))
	check(err, t)

	deployIncCounterSC(t, chain, counter)

	require.GreaterOrEqual(t, len(accessNodes), 1)

	// TODO add node #3 as access node somehow (?)
	// chain.Cluster.WaspClient(0).GetChainRecord()
	a, err := clu.WaspClient(0).GetChainRecord(chain.ChainID)
	b, err := clu.WaspClient(3).GetChainRecord(chain.ChainID)
	println(a, b)
	// accessNodes[0].

	//TODO send offledger request
}
