// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package mempool_test

import (
	"context"
	"fmt"
	"math/rand"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/kvstore/mapdb"
	"github.com/iotaledger/hive.go/log"

	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotago/iotatest"
	"github.com/iotaledger/wasp/packages/chain"
	consGR "github.com/iotaledger/wasp/packages/chain/cons/gr"
	"github.com/iotaledger/wasp/packages/chain/mempool"
	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/metrics"
	"github.com/iotaledger/wasp/packages/origin"
	"github.com/iotaledger/wasp/packages/parameters/parameterstest"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/testutil/l1starter"
	"github.com/iotaledger/wasp/packages/testutil/testchain"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/packages/testutil/testpeers"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/core/coreprocessors"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/core/migrations/allmigrations"
	"github.com/iotaledger/wasp/packages/vm/gas"
	"github.com/iotaledger/wasp/packages/vm/vmimpl"
)

type tc struct {
	n        int
	f        int
	reliable bool
}

func TestMempoolBasic(t *testing.T) {
	t.Parallel()
	tests := []tc{
		{n: 1, f: 0, reliable: true},  // Low N
		{n: 2, f: 0, reliable: true},  // Low N
		{n: 3, f: 0, reliable: true},  // Low N
		{n: 4, f: 1, reliable: true},  // Minimal robust config.
		{n: 10, f: 3, reliable: true}, // Typical config.
	}
	if !testing.Short() {
		tests = append(tests,
			tc{n: 4, f: 1, reliable: false},  // Minimal robust config.
			tc{n: 10, f: 3, reliable: false}, // Typical config.
			tc{n: 31, f: 10, reliable: true}, // Large cluster, reliable - to make test faster.
		)
	}
	for _, tst := range tests {
		t.Run(
			fmt.Sprintf("N=%v,F=%v,Reliable=%v", tst.n, tst.f, tst.reliable),
			func(tt *testing.T) { testMempoolBasic(tt, tst.n, tst.f, tst.reliable) },
		)
	}
}

// Scenario:
//   - Send an on-ledger/off-ledger requests to different nodes.
//   - Send BaseAO to all nodes.
//   - Get proposals in all nodes -> all have at least 1 of those reqs.
//   - Get both requests for all nodes.
//   - Send next BaseAO on all nodes.
//   - Get proposals -- all waiting.
//   - Send a request.
//   - Get proposals -- all received 1 request.
func testMempoolBasic(t *testing.T, n, f int, reliable bool) {
	t.Parallel()
	te := newEnv(t, n, f, reliable)
	defer te.close()

	t.Log("ServerNodesUpdated")
	tangleTime := time.Now()
	for _, node := range te.mempools {
		node.ServerNodesUpdated(te.peerPubKeys, te.peerPubKeys)
		node.TangleTimeUpdated(tangleTime)
	}
	awaitTrackHeadChannels := make([]<-chan bool, len(te.mempools))
	// deposit some funds so off-ledger requests can go through
	t.Log("TrackNewChainHead")
	for i, node := range te.mempools {
		awaitTrackHeadChannels[i] = node.TrackNewChainHead(te.stateForAnchor(i, te.anchor), nil, te.anchor, []state.Block{}, []state.Block{})
	}
	for i := range te.mempools {
		<-awaitTrackHeadChannels[i]
	}

	onLedgerReq, err := te.tcl.MakeTxAccountsDeposit(te.chainOwner)
	require.NoError(t, err)
	for _, node := range te.mempools {
		node.ReceiveOnLedgerRequest(onLedgerReq.(isc.OnLedgerRequest))
	}
	te.anchor = blockFn(te, []isc.Request{onLedgerReq}, te.anchor, tangleTime)

	offLedgerReq := isc.NewOffLedgerRequest(
		te.chainID,
		isc.NewMessage(isc.Hn("foo"), isc.Hn("bar"), isc.NewCallArguments()),
		0,
		gas.LimitsDefault.MaxGasPerRequest,
	).Sign(te.chainOwner)
	t.Log("Sending off-ledger request")
	chosenMempool := rand.Intn(len(te.mempools))
	require.Nil(t, te.mempools[chosenMempool].ReceiveOffLedgerRequest(offLedgerReq))
	te.mempools[chosenMempool].ReceiveOffLedgerRequest(offLedgerReq) // Check for duplicate receives.

	t.Log("Ask for proposals")
	proposals := make([]<-chan []*isc.RequestRef, len(te.mempools))
	for i, node := range te.mempools {
		proposals[i] = node.ConsensusProposalAsync(te.ctx, te.anchor, consGR.ConsensusID{})
	}
	t.Log("Wait for proposals and ask for decided requests")
	decided := make([]<-chan []isc.Request, len(te.mempools))
	for i, node := range te.mempools {
		proposal := <-proposals[i]
		require.True(t, len(proposal) == 1 || len(proposal) == 2)
		decided[i] = node.ConsensusRequestsAsync(te.ctx, isc.RequestRefsFromRequests([]isc.Request{offLedgerReq}))
	}
	t.Log("Wait for decided requests")
	for i := range te.mempools {
		nodeDecidedReqs := <-decided[i]
		require.Len(t, nodeDecidedReqs, 1)
	}

	// Make a block consuming those 2 requests.
	te.anchor = blockFn(te, []isc.Request{offLedgerReq}, te.anchor, tangleTime)

	// Ask proposals for the next
	proposals = make([]<-chan []*isc.RequestRef, len(te.mempools))
	for i := range te.mempools {
		proposals[i] = te.mempools[i].ConsensusProposalAsync(te.ctx, te.anchor, consGR.ConsensusID{}) // Intentionally invalid order (vs TrackNewChainHead).
	}

	// We should not get any requests, because old requests are consumed
	// and the new ones are not arrived yet.
	for i := range te.mempools {
		select {
		case refs := <-proposals[i]:
			t.Fatalf("should not get a value here, Got %+v", refs)
		default:
			// OK
		}
	}
	//
	// Add a message, we should get it now.
	offLedgerReq2 := isc.NewOffLedgerRequest(
		te.chainID,
		isc.NewMessage(isc.Hn("foo"), isc.Hn("bar"), isc.NewCallArguments()),
		1,
		gas.LimitsDefault.MaxGasPerRequest,
	).Sign(te.chainOwner)
	offLedgerRef2 := isc.RequestRefFromRequest(offLedgerReq2)
	for i := range te.mempools {
		te.mempools[i].ReceiveOffLedgerRequest(offLedgerReq2)
	}
	for i := range te.mempools {
		prop := <-proposals[i]
		require.Len(t, prop, 1)
		require.Contains(t, prop, offLedgerRef2)
	}
}

func TestMempoolsNonceGaps(t *testing.T) {
	// TODO how to remove the sleeps?
	// 1 node setup
	// send nonces 0,1,3,6,10
	// ask for proposal, assert 0,1 are proposed
	// ask again, nothing to be proposed
	// send nonce 2
	// ask for proposal, assert 2,3 are proposed
	// send nonce 4, assert 4,5 are proposed

	te := newEnv(t, 1, 0, true)
	defer te.close()

	t.Log("ServerNodesUpdated")
	tangleTime := time.Now()
	for _, node := range te.mempools {
		node.ServerNodesUpdated(te.peerPubKeys, te.peerPubKeys)
		node.TangleTimeUpdated(tangleTime)
	}
	awaitTrackHeadChannels := make([]<-chan bool, len(te.mempools))
	// deposit some funds so off-ledger requests can go through
	t.Log("TrackNewChainHead")
	for i, node := range te.mempools {
		awaitTrackHeadChannels[i] = node.TrackNewChainHead(te.stateForAnchor(i, te.anchor), nil, te.anchor, []state.Block{}, []state.Block{})
	}
	for i := range te.mempools {
		<-awaitTrackHeadChannels[i]
	}

	onLedgerReq, err := te.tcl.MakeTxAccountsDeposit(te.chainOwner)
	require.NoError(t, err)
	for _, node := range te.mempools {
		node.ReceiveOnLedgerRequest(onLedgerReq.(isc.OnLedgerRequest))
	}
	te.anchor = blockFn(te, []isc.Request{onLedgerReq}, te.anchor, tangleTime)

	// send nonces 0,1,3,6,10
	createReqWithNonce := func(nonce uint64) isc.OffLedgerRequest {
		return isc.NewOffLedgerRequest(
			te.chainID,
			isc.NewMessage(isc.Hn("foo"), isc.Hn("bar"), isc.NewCallArguments()),
			nonce,
			gas.LimitsDefault.MaxGasPerRequest,
		).Sign(te.chainOwner)
	}
	offLedgerReqs := []isc.Request{
		createReqWithNonce(0),
		createReqWithNonce(1),
		createReqWithNonce(3),
		createReqWithNonce(6),
		createReqWithNonce(10),
	}

	chosenMempool := rand.Intn(len(te.mempools))
	for _, req := range offLedgerReqs {
		t.Log("Sending off-ledger request with nonces 0,1,3,6,10")
		require.Nil(t, te.mempools[chosenMempool].ReceiveOffLedgerRequest(req.(isc.OffLedgerRequest)))
	}
	time.Sleep(200 * time.Millisecond) // give some time for the requests to reach the pool

	askProposalExpectReqs := func(anchor *isc.StateAnchor, reqs ...isc.Request) *isc.StateAnchor {
		t.Log("Ask for proposals")
		proposalCh := make([]<-chan []*isc.RequestRef, len(te.mempools))
		for i, node := range te.mempools {
			proposalCh[i] = node.ConsensusProposalAsync(te.ctx, anchor, consGR.ConsensusID{})
		}
		t.Log("Wait for proposals and ask for decided requests")
		decided := make([]<-chan []isc.Request, len(te.mempools))
		for i, node := range te.mempools {
			proposal := <-proposalCh[i]
			require.Len(t, proposal, len(reqs))
			decided[i] = node.ConsensusRequestsAsync(te.ctx, proposal)
		}
		t.Log("Wait for decided requests")
		var nodeDecidedReqs []isc.Request
		for i := range te.mempools {
			nodeDecidedReqs = <-decided[i]
			require.Len(t, nodeDecidedReqs, len(reqs))
			// they aren't ordered here, can be 0,1 or 1,0
			for _, r := range reqs {
				require.Contains(t, nodeDecidedReqs, r)
			}
		}
		//
		// Make a block consuming those 2 requests.
		return blockFn(te, nodeDecidedReqs, anchor, tangleTime)
	}

	emptyProposalFn := func(anchor *isc.StateAnchor) {
		// ask again, nothing to be proposed
		//
		// Ask proposals for the next
		proposals := make([]<-chan []*isc.RequestRef, len(te.mempools))
		for i := range te.mempools {
			proposals[i] = te.mempools[i].ConsensusProposalAsync(te.ctx, anchor, consGR.ConsensusID{}) // Intentionally invalid order (vs TrackNewChainHead).
		}
		//
		// We should not get any requests, there is a gap in the nonces
		for i := range te.mempools {
			select {
			case refs := <-proposals[i]:
				t.Fatalf("should not get a value here, Got %+v", refs)
			default:
				// OK
			}
		}
	}
	// ask for proposal, assert 0,1 are proposed
	te.anchor = askProposalExpectReqs(te.anchor, offLedgerReqs[0], offLedgerReqs[1])

	// next proposal is empty
	emptyProposalFn(te.anchor)

	// send nonce 2
	reqNonce2 := createReqWithNonce(2)
	t.Log("Sending off-ledger request with nonce 2")
	require.Nil(t, te.mempools[chosenMempool].ReceiveOffLedgerRequest(reqNonce2))
	time.Sleep(200 * time.Millisecond) // give some time for the requests to reach the pool

	// ask for proposal, assert 2,3 are proposed
	te.anchor = askProposalExpectReqs(te.anchor, reqNonce2, offLedgerReqs[2])

	// next proposal is empty
	emptyProposalFn(te.anchor)

	// send nonce 5, assert proposal is still empty (there is still a gap with the state)
	reqNonce5 := createReqWithNonce(5)
	t.Log("Sending off-ledger request with nonce 5")
	require.Nil(t, te.mempools[chosenMempool].ReceiveOffLedgerRequest(reqNonce5))
	time.Sleep(200 * time.Millisecond) // give some time for the requests to reach the pool

	emptyProposalFn(te.anchor)

	// send nonce 4
	reqNonce4 := createReqWithNonce(4)
	t.Log("Sending off-ledger request with nonce 4")
	require.Nil(t, te.mempools[chosenMempool].ReceiveOffLedgerRequest(reqNonce4))
	time.Sleep(200 * time.Millisecond) // give some time for the requests to reach the pool

	// ask for proposal, assert 4,5,6 are proposed
	askProposalExpectReqs(te.anchor, reqNonce4, reqNonce5, offLedgerReqs[3])
	// nonce 10 was never proposed

	t.Run("request with old nonce is rejected", func(t *testing.T) {
		reqNonce0 := createReqWithNonce(0)
		err = te.mempools[chosenMempool].ReceiveOffLedgerRequest(reqNonce0)
		require.ErrorContains(t, err, "bad nonce")
	})

	t.Run("request with not enough funds is rejected", func(t *testing.T) {
		kp := cryptolib.NewKeyPair()
		req := isc.NewOffLedgerRequest(
			te.chainID,
			isc.NewMessage(isc.Hn("foo"), isc.Hn("bar"), isc.NewCallArguments()),
			0,
			gas.LimitsDefault.MaxGasPerRequest,
		).Sign(kp)
		err = te.mempools[chosenMempool].ReceiveOffLedgerRequest(req)
		require.ErrorContains(t, err, "not enough funds")
	})
}

func TestMempoolChainOwner(t *testing.T) {
	te := newEnv(t, 1, 0, true)
	defer te.close()

	t.Log("ServerNodesUpdated")
	tangleTime := time.Now()
	for _, node := range te.mempools {
		node.ServerNodesUpdated(te.peerPubKeys, te.peerPubKeys)
		node.TangleTimeUpdated(tangleTime)
	}
	awaitTrackHeadChannels := make([]<-chan bool, len(te.mempools))
	// deposit some funds so off-ledger requests can go through
	t.Log("TrackNewChainHead")
	for i, node := range te.mempools {
		awaitTrackHeadChannels[i] = node.TrackNewChainHead(te.stateForAnchor(i, te.anchor), nil, te.anchor, []state.Block{}, []state.Block{})
	}
	for i := range te.mempools {
		<-awaitTrackHeadChannels[i]
	}
	require.Equal(t, te.chainOwner.Address().String(), te.anchor.Owner().String(), "chainOwner and anchor owner are not the same")

	governanceState := governance.NewStateReaderFromChainState(te.stateForAnchor(0, te.anchor))
	chainOwner := governanceState.GetChainAdmin()
	chainOwnerAddress, success := isc.AddressFromAgentID(chainOwner)
	require.True(t, success, "unable to get address from chain owner agentID")
	require.Equal(t, te.chainOwner.Address().String(), chainOwnerAddress.String(), "chain owner incorrect")
}

func TestMempoolOverrideNonce(t *testing.T) {
	// 1 node setup
	// send nonce 0
	// send another request with the same nonce 0
	// assert the last request is proposed
	te := newEnv(t, 1, 0, true)
	defer te.close()

	tangleTime := time.Now()
	for _, node := range te.mempools {
		node.ServerNodesUpdated(te.peerPubKeys, te.peerPubKeys)
		node.TangleTimeUpdated(tangleTime)
	}
	awaitTrackHeadChannels := make([]<-chan bool, len(te.mempools))
	// deposit some funds so off-ledger requests can go through
	t.Log("TrackNewChainHead")
	for i, node := range te.mempools {
		awaitTrackHeadChannels[i] = node.TrackNewChainHead(te.stateForAnchor(i, te.anchor), nil, nil, []state.Block{}, []state.Block{})
	}
	for i := range te.mempools {
		<-awaitTrackHeadChannels[i]
	}

	onLedgerReq, err := te.tcl.MakeTxAccountsDeposit(te.chainOwner)
	require.NoError(t, err)
	for _, node := range te.mempools {
		node.ReceiveOnLedgerRequest(onLedgerReq.(isc.OnLedgerRequest))
	}
	te.anchor = blockFn(te, []isc.Request{onLedgerReq}, te.anchor, tangleTime)

	initialReq := isc.NewOffLedgerRequest(
		te.chainID,
		isc.NewMessage(isc.Hn("foo"), isc.Hn("bar"), isc.NewCallArguments()),
		0,
		gas.LimitsDefault.MaxGasPerRequest,
	).Sign(te.chainOwner)
	time.Sleep(400 * time.Millisecond) // give some time for the requests to reach the pool
	require.NoError(t, te.mempools[0].ReceiveOffLedgerRequest(initialReq))
	time.Sleep(200 * time.Millisecond) // give some time for the requests to reach the pool

	overwritingReq := isc.NewOffLedgerRequest(
		te.chainID,
		isc.NewMessage(isc.Hn("baz"), isc.Hn("bar"), isc.NewCallArguments()),
		0,
		gas.LimitsDefault.MaxGasPerRequest,
	).Sign(te.chainOwner)

	require.NoError(t, te.mempools[0].ReceiveOffLedgerRequest(overwritingReq))
	time.Sleep(200 * time.Millisecond) // give some time for the requests to reach the pool
	reqRefs := <-te.mempools[0].ConsensusProposalAsync(te.ctx, te.anchor, consGR.ConsensusID{})
	proposedReqs := <-te.mempools[0].ConsensusRequestsAsync(te.ctx, reqRefs)
	require.Len(t, proposedReqs, 1)
	require.Equal(t, overwritingReq, proposedReqs[0])
	require.NotEqual(t, initialReq, proposedReqs[0])
}

func TestTTL(t *testing.T) {
	te := newEnv(t, 1, 0, true)
	// override the TTL
	chainMetrics := metrics.NewChainMetricsProvider().GetChainMetrics(isc.EmptyChainID())
	te.mempools[0] = mempool.New(
		te.ctx,
		te.chainID,
		te.peerIdentities[0],
		te.networkProviders[0],
		te.log.NewChildLogger(fmt.Sprintf("N#%v", 0)),
		chainMetrics.Mempool,
		chainMetrics.Pipe,
		chain.NewEmptyChainListener(),
		mempool.Settings{
			TTL:                    200 * time.Millisecond,
			MaxOffledgerInPool:     1000,
			MaxOnledgerInPool:      1000,
			MaxTimedInPool:         1000,
			MaxOnledgerToPropose:   1000,
			MaxOffledgerToPropose:  1000,
			MaxOffledgerPerAccount: 1000,
		},
		1*time.Second,
		func() {},
	)
	defer te.close()
	start := time.Now()
	mp := te.mempools[0]
	mp.TangleTimeUpdated(start)

	// deposit some funds so off-ledger requests can go through
	<-mp.TrackNewChainHead(te.stateForAnchor(0, te.anchor), nil, nil, []state.Block{}, []state.Block{})

	onLedgerReq1, err := te.tcl.MakeTxAccountsDeposit(te.chainOwner)
	require.NoError(t, err)
	for _, node := range te.mempools {
		node.ReceiveOnLedgerRequest(onLedgerReq1.(isc.OnLedgerRequest))
	}
	te.anchor = blockFn(te, []isc.Request{onLedgerReq1}, te.anchor, start)

	// send offledger request, assert it is returned, make 201ms pass, assert it is not returned anymore
	offLedgerReq := isc.NewOffLedgerRequest(
		te.chainID,
		isc.NewMessage(isc.Hn("foo"), isc.Hn("bar"), isc.NewCallArguments()),
		0,
		gas.LimitsDefault.MaxGasPerRequest,
	).Sign(te.chainOwner)
	t.Log("Sending off-ledger request")
	require.Nil(t, mp.ReceiveOffLedgerRequest(offLedgerReq))

	reqs := <-mp.ConsensusProposalAsync(te.ctx, te.anchor, consGR.ConsensusID{})
	require.Len(t, reqs, 1)
	time.Sleep(201 * time.Millisecond)

	// we need to add some request because ConsensusProposalAsync will not return an empty list.
	onLedgerReq2, err := te.tcl.MakeTxAccountsDeposit(te.chainOwner)
	require.NoError(t, err)
	mp.ReceiveOnLedgerRequest(onLedgerReq2.(isc.OnLedgerRequest))

	reqs2 := <-mp.ConsensusProposalAsync(te.ctx, te.anchor, consGR.ConsensusID{})
	require.Len(t, reqs2, 1) // only the last request is returned
}

func blockFn(te *testEnv, reqs []isc.Request, anchor *isc.StateAnchor, tangleTime time.Time) *isc.StateAnchor {
	// sort reqs by nonce
	slices.SortFunc(reqs, func(a, b isc.Request) int {
		return int(a.(isc.OffLedgerRequest).Nonce() - b.(isc.OffLedgerRequest).Nonce())
	})

	store := te.stores[0]
	vmTask := &vm.VMTask{
		Processors: coreprocessors.NewConfigWithTestContracts(),
		Anchor:     anchor,
		GasCoin: &coin.CoinWithRef{
			Value: isc.GasCoinTargetValue,
			Type:  coin.BaseTokenType,
			Ref:   iotatest.RandomObjectRef(),
		},
		Store:                store,
		Requests:             reqs,
		Timestamp:            tangleTime,
		Entropy:              hashing.HashDataBlake2b([]byte{2, 1, 7}),
		ValidatorFeeTarget:   accounts.CommonAccount(),
		L1Params:             parameterstest.L1Mock,
		EstimateGasMode:      false,
		EnableGasBurnLogging: false,
		Log:                  te.log.NewChildLogger("VM"),
		Migrations:           allmigrations.DefaultScheme,
	}
	vmResult, err := vmimpl.Run(vmTask)
	require.NoError(te.t, err)
	block := store.Commit(vmResult.StateDraft)
	chainState, err := store.StateByTrieRoot(block.TrieRoot())
	require.NoError(te.t, err)
	anchor, err = te.tcl.RunOnChainStateTransition(anchor, vmResult.UnsignedTransaction)
	require.NoError(te.t, err)

	// Check if block has both requests as consumed.
	receipts, err := blocklog.RequestReceiptsFromBlock(block)
	require.NoError(te.t, err)
	require.Len(te.t, receipts, len(reqs))

	// FIXME directly compare two object with their pointer may cause a false negative result, so using byte slice would be easier
	// blockReqs := []isc.Request{}
	blockReqBytes := [][]byte{}
	for i := range receipts {
		blockReqBytes = append(blockReqBytes, receipts[i].Request.Bytes())
	}
	for _, req := range reqs {
		require.Contains(te.t, blockReqBytes, req.Bytes())
	}

	// sync mempools with new state
	awaitTrackHeadChannels := make([]<-chan bool, len(te.mempools))
	for i := range te.mempools {
		awaitTrackHeadChannels[i] = te.mempools[i].TrackNewChainHead(chainState, nil, anchor, []state.Block{block}, []state.Block{})
	}
	for i := range te.mempools {
		<-awaitTrackHeadChannels[i]
	}
	return anchor
}

/////////////////////////////////////testEnv/////////////////////////////////////

// Setups testing environment and holds all the relevant info.
type testEnv struct {
	t                *testing.T
	ctx              context.Context
	ctxCancel        context.CancelFunc
	log              log.Logger
	chainOwner       *cryptolib.KeyPair
	peeringURLs      []string
	peerIdentities   []*cryptolib.KeyPair
	peerPubKeys      []*cryptolib.PublicKey
	peeringNetwork   *testutil.PeeringNetwork
	networkProviders []peering.NetworkProvider
	tcl              *testchain.TestChainLedger
	cmtAddress       *cryptolib.Address
	chainID          isc.ChainID
	anchor           *isc.StateAnchor
	mempools         []mempool.Mempool
	stores           []state.Store
}

func newEnv(t *testing.T, n, f int, reliable bool) *testEnv {
	te := &testEnv{t: t}
	te.ctx, te.ctxCancel = context.WithCancel(context.Background())
	te.log = testlogger.NewLogger(t)

	// Create ledger accounts. Requesting funds twice to get two coin objects (so we don't need to split one later)
	te.chainOwner = cryptolib.NewKeyPair()
	require.NoError(t, iotaclient.RequestFundsFromFaucet(context.Background(), te.chainOwner.Address().AsIotaAddress(), l1starter.Instance().FaucetURL()))
	require.NoError(t, iotaclient.RequestFundsFromFaucet(context.Background(), te.chainOwner.Address().AsIotaAddress(), l1starter.Instance().FaucetURL()))

	// Create a fake network and keys for the tests.
	te.peeringURLs, te.peerIdentities = testpeers.SetupKeys(uint16(n))
	te.peerPubKeys = make([]*cryptolib.PublicKey, len(te.peerIdentities))
	for i := range te.peerPubKeys {
		te.peerPubKeys[i] = te.peerIdentities[i].GetPublicKey()
	}
	var networkBehaviour testutil.PeeringNetBehavior
	if reliable {
		networkBehaviour = testutil.NewPeeringNetReliable(te.log)
	} else {
		netLogger := testlogger.WithLevel(te.log.NewChildLogger("Network"), log.LevelInfo, false)
		networkBehaviour = testutil.NewPeeringNetUnreliable(80, 20, 10*time.Millisecond, 200*time.Millisecond, netLogger)
	}
	te.peeringNetwork = testutil.NewPeeringNetwork(
		te.peeringURLs, te.peerIdentities, 10000,
		networkBehaviour,
		testlogger.WithLevel(te.log, log.LevelWarning, false),
	)
	te.networkProviders = te.peeringNetwork.NetworkProviders()
	te.cmtAddress, _ = testpeers.SetupDkgTrivial(t, n, f, te.peerIdentities, nil)

	l1client := l1starter.Instance().L1Client()

	objs, err := l1client.GetAllCoins(context.Background(), iotaclient.GetAllCoinsRequest{
		Owner: te.chainOwner.Address().AsIotaAddress(),
	})
	require.NoError(t, err)

	fmt.Println(objs)

	iscPackage, err := l1client.DeployISCContracts(context.Background(), cryptolib.SignerToIotaSigner(te.chainOwner))
	require.NoError(t, err)

	te.tcl = testchain.NewTestChainLedger(t, te.chainOwner, &iscPackage, l1client)
	var originDepositVal coin.Value
	te.anchor, originDepositVal = te.tcl.MakeTxChainOrigin()

	// Initialize the nodes.
	te.mempools = make([]mempool.Mempool, len(te.peerIdentities))
	te.stores = make([]state.Store, len(te.peerIdentities))
	for i := range te.peerIdentities {
		te.stores[i] = state.NewStoreWithUniqueWriteMutex(mapdb.NewMapDB())
		origin.InitChainByStateMetadataBytes(te.stores[i], te.anchor.GetStateMetadata(), originDepositVal, parameterstest.L1Mock)
		require.NoError(t, err)
		chainMetrics := metrics.NewChainMetricsProvider().GetChainMetrics(isc.EmptyChainID())
		te.mempools[i] = mempool.New(
			te.ctx,
			te.chainID,
			te.peerIdentities[i],
			te.networkProviders[i],
			testlogger.WithLevel(te.log.NewChildLogger(fmt.Sprintf("N#%v", i)), log.LevelDebug, false),
			chainMetrics.Mempool,
			chainMetrics.Pipe,
			chain.NewEmptyChainListener(),
			mempool.Settings{
				TTL:                    24 * time.Hour,
				MaxOffledgerInPool:     1000,
				MaxOnledgerInPool:      1000,
				MaxTimedInPool:         1000,
				MaxOnledgerToPropose:   1000,
				MaxOffledgerToPropose:  1000,
				MaxOffledgerPerAccount: 1000,
			},
			1*time.Second,
			func() {},
		)
	}
	return te
}

func (te *testEnv) stateForAnchor(i int, anchor *isc.StateAnchor) state.State {
	l1Commitment, err := transaction.L1CommitmentFromAnchor(anchor)
	require.NoError(te.t, err)
	st, err := te.stores[i].StateByTrieRoot(l1Commitment.TrieRoot())
	require.NoError(te.t, err)
	return st
}

func (te *testEnv) close() {
	te.ctxCancel()
	te.peeringNetwork.Close()
	te.log.Shutdown()
}
