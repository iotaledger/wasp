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
	"github.com/iotaledger/hive.go/logger"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/tpkg"
	"github.com/iotaledger/wasp/contracts/native/inccounter"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/chain/mempool"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/metrics"
	"github.com/iotaledger/wasp/packages/origin"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/testutil/testchain"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/packages/testutil/testpeers"
	"github.com/iotaledger/wasp/packages/testutil/utxodb"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/core/coreprocessors"
	"github.com/iotaledger/wasp/packages/vm/gas"
	"github.com/iotaledger/wasp/packages/vm/processors"
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
//

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
		awaitTrackHeadChannels[i] = node.TrackNewChainHead(te.stateForAO(i, te.originAO), nil, te.originAO, []state.Block{}, []state.Block{})
	}
	for i := range te.mempools {
		<-awaitTrackHeadChannels[i]
	}

	output := transaction.BasicOutputFromPostData(
		te.governor.Address(),
		isc.EmptyContractIdentity(),
		isc.RequestParameters{
			TargetAddress: te.chainID.AsAddress(),
			Assets:        isc.NewAssetsBaseTokens(10 * isc.Million),
		},
	)
	onLedgerReq, err := isc.OnLedgerFromUTXO(output, tpkg.RandOutputID(uint16(0)))
	require.NoError(t, err)
	for _, node := range te.mempools {
		node.ReceiveOnLedgerRequest(onLedgerReq)
	}
	currentAO := blockFn(te, []isc.Request{onLedgerReq}, te.originAO, tangleTime)

	//
	offLedgerReq := isc.NewOffLedgerRequest(
		isc.RandomChainID(),
		isc.Hn("foo"),
		isc.Hn("bar"),
		dict.New(),
		0,
		gas.LimitsDefault.MaxGasPerRequest,
	).Sign(te.governor)
	t.Log("Sending off-ledger request")
	chosenMempool := rand.Intn(len(te.mempools))
	require.Nil(t, te.mempools[chosenMempool].ReceiveOffLedgerRequest(offLedgerReq))
	te.mempools[chosenMempool].ReceiveOffLedgerRequest(offLedgerReq) // Check for duplicate receives.

	t.Log("Ask for proposals")
	proposals := make([]<-chan []*isc.RequestRef, len(te.mempools))
	for i, node := range te.mempools {
		proposals[i] = node.ConsensusProposalAsync(te.ctx, currentAO)
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
	//
	// Make a block consuming those 2 requests.
	currentAO = blockFn(te, []isc.Request{offLedgerReq}, currentAO, tangleTime)

	//
	// Ask proposals for the next
	proposals = make([]<-chan []*isc.RequestRef, len(te.mempools))
	for i := range te.mempools {
		proposals[i] = te.mempools[i].ConsensusProposalAsync(te.ctx, currentAO) // Intentionally invalid order (vs TrackNewChainHead).
	}
	//
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
		isc.RandomChainID(),
		isc.Hn("foo"),
		isc.Hn("bar"),
		dict.New(),
		1,
		gas.LimitsDefault.MaxGasPerRequest,
	).Sign(te.governor)
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

func blockFn(te *testEnv, reqs []isc.Request, ao *isc.AliasOutputWithID, tangleTime time.Time) *isc.AliasOutputWithID {
	// sort reqs by nonce
	slices.SortFunc(reqs, func(a, b isc.Request) int {
		return int(a.(isc.OffLedgerRequest).Nonce() - b.(isc.OffLedgerRequest).Nonce())
	})

	store := te.stores[0]
	vmTask := &vm.VMTask{
		Processors:           processors.MustNew(coreprocessors.NewConfigWithCoreContracts().WithNativeContracts(inccounter.Processor)),
		AnchorOutput:         ao.GetAliasOutput(),
		AnchorOutputID:       ao.OutputID(),
		Store:                store,
		Requests:             reqs,
		TimeAssumption:       tangleTime,
		Entropy:              hashing.HashDataBlake2b([]byte{2, 1, 7}),
		ValidatorFeeTarget:   accounts.CommonAccount(),
		EstimateGasMode:      false,
		EnableGasBurnLogging: false,
		Log:                  te.log.Named("VM"),
	}
	vmResult, err := vmimpl.Run(vmTask)
	require.NoError(te.t, err)
	block := store.Commit(vmResult.StateDraft)
	chainState, err := store.StateByTrieRoot(block.TrieRoot())
	require.NoError(te.t, err)
	//
	// Check if block has both requests as consumed.
	receipts, err := blocklog.RequestReceiptsFromBlock(block)
	require.NoError(te.t, err)
	require.Len(te.t, receipts, len(reqs))
	blockReqs := []isc.Request{}
	for i := range receipts {
		blockReqs = append(blockReqs, receipts[i].Request)
	}
	for _, req := range reqs {
		require.Contains(te.t, blockReqs, req)
	}
	nextAO := te.tcl.FakeStateTransition(ao, block.L1Commitment())

	// sync mempools with new state
	awaitTrackHeadChannels := make([]<-chan bool, len(te.mempools))
	for i := range te.mempools {
		awaitTrackHeadChannels[i] = te.mempools[i].TrackNewChainHead(chainState, ao, nextAO, []state.Block{block}, []state.Block{})
	}
	for i := range te.mempools {
		<-awaitTrackHeadChannels[i]
	}
	return nextAO
}

func TestTimeLock(t *testing.T) {
	t.Parallel()
	tests := []tc{
		{n: 1, f: 0, reliable: true},  // Low N
		{n: 2, f: 0, reliable: true},  // Low N
		{n: 3, f: 0, reliable: true},  // Low N
		{n: 4, f: 1, reliable: true},  // Minimal robust config.
		{n: 10, f: 3, reliable: true}, // Typical config.
	}
	for _, tst := range tests {
		t.Run(
			fmt.Sprintf("N=%v,F=%v,Reliable=%v", tst.n, tst.f, tst.reliable),
			func(tt *testing.T) { testTimeLock(tt, tst.n, tst.f, tst.reliable) },
		)
	}
}

func testTimeLock(t *testing.T, n, f int, reliable bool) { //nolint:gocyclo
	t.Parallel()
	te := newEnv(t, n, f, reliable)
	defer te.close()
	start := time.Now()
	requests := getRequestsOnLedger(t, te.chainID.AsAddress(), 6, func(i int, p *isc.RequestParameters) {
		switch i {
		case 0: // + No time lock
		case 1: // + Time lock before start
			p.Options.Timelock = start.Add(-2 * time.Hour)
		case 2: // + Time lock slightly before start due to time.Now() in ReadyNow being called later than in this test
			p.Options.Timelock = start
		case 3: // - Time lock 5s after start
			p.Options.Timelock = start.Add(5 * time.Second)
		case 4: // - Time lock 2h after start
			p.Options.Timelock = start.Add(2 * time.Hour)
		case 5: // - Time lock after expiration
			p.Options.Timelock = start.Add(3 * time.Second)
			p.Options.Expiration = &isc.Expiration{
				Time:          start.Add(2 * time.Second),
				ReturnAddress: te.chainID.AsAddress(),
			}
		}
	})
	reqRefs := []*isc.RequestRef{
		isc.RequestRefFromRequest(requests[0]),
		isc.RequestRefFromRequest(requests[1]),
		isc.RequestRefFromRequest(requests[2]),
		isc.RequestRefFromRequest(requests[3]),
		isc.RequestRefFromRequest(requests[4]),
		isc.RequestRefFromRequest(requests[5]),
	}
	//
	// Add the requests.
	for _, mp := range te.mempools {
		for _, r := range requests {
			mp.ReceiveOnLedgerRequest(r)
		}
	}
	for i, mp := range te.mempools {
		mp.TangleTimeUpdated(start)
		mp.ServerNodesUpdated(te.peerPubKeys, te.peerPubKeys)
		mp.TrackNewChainHead(te.stateForAO(i, te.originAO), nil, te.originAO, []state.Block{}, []state.Block{})
	}
	//
	// Check, if requests are proposed.
	time.Sleep(100 * time.Millisecond) // Just to make sure all the events have been consumed.
	for _, mp := range te.mempools {
		reqs := <-mp.ConsensusProposalAsync(te.ctx, te.originAO)
		require.Len(t, reqs, 3)
		require.Contains(t, reqs, reqRefs[0])
		require.Contains(t, reqs, reqRefs[1])
		require.Contains(t, reqs, reqRefs[2])
	}
	//
	// Add the requests twice should keep things the same.
	for _, mp := range te.mempools {
		for _, r := range requests {
			mp.ReceiveOnLedgerRequest(r)
		}
	}
	time.Sleep(100 * time.Millisecond) // Just to make sure all the events have been consumed.
	for _, mp := range te.mempools {
		reqs := <-mp.ConsensusProposalAsync(te.ctx, te.originAO)
		require.Len(t, reqs, 3)
		require.Contains(t, reqs, reqRefs[0])
		require.Contains(t, reqs, reqRefs[1])
		require.Contains(t, reqs, reqRefs[2])
	}
	//
	// More requests are proposed after 5s
	for _, mp := range te.mempools {
		mp.TangleTimeUpdated(start.Add(10 * time.Second))
	}
	time.Sleep(100 * time.Millisecond) // Just to make sure all the events have been consumed.
	for _, mp := range te.mempools {
		reqs := <-mp.ConsensusProposalAsync(te.ctx, te.originAO)
		require.Len(t, reqs, 4)
		require.Contains(t, reqs, reqRefs[0])
		require.Contains(t, reqs, reqRefs[1])
		require.Contains(t, reqs, reqRefs[2])
		require.Contains(t, reqs, reqRefs[3])
	}
	//
	// Even more requests are proposed after 10h.
	for _, mp := range te.mempools {
		mp.TangleTimeUpdated(start.Add(10 * time.Hour))
	}
	time.Sleep(100 * time.Millisecond) // Just to make sure all the events have been consumed.
	for _, mp := range te.mempools {
		reqs := <-mp.ConsensusProposalAsync(te.ctx, te.originAO)
		require.Len(t, reqs, 5)
		require.Contains(t, reqs, reqRefs[0])
		require.Contains(t, reqs, reqRefs[1])
		require.Contains(t, reqs, reqRefs[2])
		require.Contains(t, reqs, reqRefs[3])
		require.Contains(t, reqs, reqRefs[4])
	}
}

func TestExpiration(t *testing.T) {
	t.Parallel()
	tests := []tc{
		{n: 1, f: 0, reliable: true},  // Low N
		{n: 2, f: 0, reliable: true},  // Low N
		{n: 3, f: 0, reliable: true},  // Low N
		{n: 4, f: 1, reliable: true},  // Minimal robust config.
		{n: 10, f: 3, reliable: true}, // Typical config.
	}
	for _, tst := range tests {
		t.Run(
			fmt.Sprintf("N=%v,F=%v,Reliable=%v", tst.n, tst.f, tst.reliable),
			func(tt *testing.T) { testExpiration(tt, tst.n, tst.f, tst.reliable) },
		)
	}
}

func testExpiration(t *testing.T, n, f int, reliable bool) {
	t.Parallel()
	te := newEnv(t, n, f, reliable)
	defer te.close()
	start := time.Now()
	requests := getRequestsOnLedger(t, te.chainID.AsAddress(), 4, func(i int, p *isc.RequestParameters) {
		switch i {
		case 1: // expired
			p.Options.Expiration = &isc.Expiration{
				Time:          start.Add(-isc.RequestConsideredExpiredWindow),
				ReturnAddress: te.chainID.AsAddress(),
			}
		case 2: // will expire soon
			p.Options.Expiration = &isc.Expiration{
				Time:          start.Add(isc.RequestConsideredExpiredWindow / 2),
				ReturnAddress: te.chainID.AsAddress(),
			}
		case 3: // not expired yet
			p.Options.Expiration = &isc.Expiration{
				Time:          start.Add(isc.RequestConsideredExpiredWindow * 2),
				ReturnAddress: te.chainID.AsAddress(),
			}
		}
	})
	reqRefs := []*isc.RequestRef{
		isc.RequestRefFromRequest(requests[0]),
		isc.RequestRefFromRequest(requests[1]),
		isc.RequestRefFromRequest(requests[2]),
		isc.RequestRefFromRequest(requests[3]),
	}
	//
	// Add the requests.
	for _, mp := range te.mempools {
		for _, r := range requests {
			mp.ReceiveOnLedgerRequest(r)
		}
	}
	for i, mp := range te.mempools {
		mp.TangleTimeUpdated(start)
		mp.ServerNodesUpdated(te.peerPubKeys, te.peerPubKeys)
		mp.TrackNewChainHead(te.stateForAO(i, te.originAO), nil, te.originAO, []state.Block{}, []state.Block{})
	}
	//
	// Check, if requests are proposed.
	time.Sleep(100 * time.Millisecond) // Just to make sure all the events have been consumed.
	for _, mp := range te.mempools {
		reqs := <-mp.ConsensusProposalAsync(te.ctx, te.originAO)
		require.Len(t, reqs, 2)
		require.Contains(t, reqs, reqRefs[0])
		require.Contains(t, reqs, reqRefs[3])
	}
	//
	// The remaining request with an expiry expires some time after.
	for _, mp := range te.mempools {
		mp.TangleTimeUpdated(start.Add(10 * isc.RequestConsideredExpiredWindow))
	}
	time.Sleep(100 * time.Millisecond) // Just to make sure all the events have been consumed.
	for _, mp := range te.mempools {
		reqs := <-mp.ConsensusProposalAsync(te.ctx, te.originAO)
		require.Len(t, reqs, 1)
		require.Contains(t, reqs, reqRefs[0])
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
		awaitTrackHeadChannels[i] = node.TrackNewChainHead(te.stateForAO(i, te.originAO), nil, te.originAO, []state.Block{}, []state.Block{})
	}
	for i := range te.mempools {
		<-awaitTrackHeadChannels[i]
	}

	output := transaction.BasicOutputFromPostData(
		te.governor.Address(),
		isc.EmptyContractIdentity(),
		isc.RequestParameters{
			TargetAddress: te.chainID.AsAddress(),
			Assets:        isc.NewAssetsBaseTokens(10 * isc.Million),
		},
	)
	onLedgerReq, err := isc.OnLedgerFromUTXO(output, tpkg.RandOutputID(uint16(0)))
	require.NoError(t, err)
	for _, node := range te.mempools {
		node.ReceiveOnLedgerRequest(onLedgerReq)
	}
	currentAO := blockFn(te, []isc.Request{onLedgerReq}, te.originAO, tangleTime)

	// send nonces 0,1,3,6,10
	createReqWithNonce := func(nonce uint64) isc.OffLedgerRequest {
		return isc.NewOffLedgerRequest(
			isc.RandomChainID(),
			isc.Hn("foo"),
			isc.Hn("bar"),
			dict.New(),
			nonce,
			gas.LimitsDefault.MaxGasPerRequest,
		).Sign(te.governor)
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

	askProposalExpectReqs := func(ao *isc.AliasOutputWithID, reqs ...isc.Request) *isc.AliasOutputWithID {
		t.Log("Ask for proposals")
		proposals := make([]<-chan []*isc.RequestRef, len(te.mempools))
		for i, node := range te.mempools {
			proposals[i] = node.ConsensusProposalAsync(te.ctx, ao)
		}
		t.Log("Wait for proposals and ask for decided requests")
		decided := make([]<-chan []isc.Request, len(te.mempools))
		for i, node := range te.mempools {
			proposal := <-proposals[i]
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
		return blockFn(te, nodeDecidedReqs, ao, tangleTime)
	}

	emptyProposalFn := func(ao *isc.AliasOutputWithID) {
		// ask again, nothing to be proposed
		//
		// Ask proposals for the next
		proposals := make([]<-chan []*isc.RequestRef, len(te.mempools))
		for i := range te.mempools {
			proposals[i] = te.mempools[i].ConsensusProposalAsync(te.ctx, ao) // Intentionally invalid order (vs TrackNewChainHead).
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
	currentAO = askProposalExpectReqs(currentAO, offLedgerReqs[0], offLedgerReqs[1])

	// next proposal is empty
	emptyProposalFn(currentAO)

	// send nonce 2
	reqNonce2 := createReqWithNonce(2)
	t.Log("Sending off-ledger request with nonce 2")
	require.Nil(t, te.mempools[chosenMempool].ReceiveOffLedgerRequest(reqNonce2))
	time.Sleep(200 * time.Millisecond) // give some time for the requests to reach the pool

	// ask for proposal, assert 2,3 are proposed
	currentAO = askProposalExpectReqs(currentAO, reqNonce2, offLedgerReqs[2])

	// next proposal is empty
	emptyProposalFn(currentAO)

	// send nonce 5, assert proposal is still empty (there is still a gap with the state)
	reqNonce5 := createReqWithNonce(5)
	t.Log("Sending off-ledger request with nonce 5")
	require.Nil(t, te.mempools[chosenMempool].ReceiveOffLedgerRequest(reqNonce5))
	time.Sleep(200 * time.Millisecond) // give some time for the requests to reach the pool

	emptyProposalFn(currentAO)

	// send nonce 4
	reqNonce4 := createReqWithNonce(4)
	t.Log("Sending off-ledger request with nonce 4")
	require.Nil(t, te.mempools[chosenMempool].ReceiveOffLedgerRequest(reqNonce4))
	time.Sleep(200 * time.Millisecond) // give some time for the requests to reach the pool

	// ask for proposal, assert 4,5,6 are proposed
	askProposalExpectReqs(currentAO, reqNonce4, reqNonce5, offLedgerReqs[3])
	// nonce 10 was never proposed
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
		awaitTrackHeadChannels[i] = node.TrackNewChainHead(te.stateForAO(i, te.originAO), nil, te.originAO, []state.Block{}, []state.Block{})
	}
	for i := range te.mempools {
		<-awaitTrackHeadChannels[i]
	}

	output := transaction.BasicOutputFromPostData(
		te.governor.Address(),
		isc.EmptyContractIdentity(),
		isc.RequestParameters{
			TargetAddress: te.chainID.AsAddress(),
			Assets:        isc.NewAssetsBaseTokens(10 * isc.Million),
		},
	)
	onLedgerReq, err := isc.OnLedgerFromUTXO(output, tpkg.RandOutputID(uint16(0)))
	require.NoError(t, err)
	for _, node := range te.mempools {
		node.ReceiveOnLedgerRequest(onLedgerReq)
	}
	currentAO := blockFn(te, []isc.Request{onLedgerReq}, te.originAO, tangleTime)

	initialReq := isc.NewOffLedgerRequest(
		isc.RandomChainID(),
		isc.Hn("foo"),
		isc.Hn("bar"),
		dict.New(),
		0,
		gas.LimitsDefault.MaxGasPerRequest,
	).Sign(te.governor)

	require.NoError(t, te.mempools[0].ReceiveOffLedgerRequest(initialReq))
	time.Sleep(200 * time.Millisecond) // give some time for the requests to reach the pool

	overwritingReq := isc.NewOffLedgerRequest(
		isc.RandomChainID(),
		isc.Hn("baz"),
		isc.Hn("bar"),
		dict.New(),
		0,
		gas.LimitsDefault.MaxGasPerRequest,
	).Sign(te.governor)

	require.NoError(t, te.mempools[0].ReceiveOffLedgerRequest(overwritingReq))
	time.Sleep(200 * time.Millisecond) // give some time for the requests to reach the pool
	reqRefs := <-te.mempools[0].ConsensusProposalAsync(te.ctx, currentAO)
	proposedReqs := <-te.mempools[0].ConsensusRequestsAsync(te.ctx, reqRefs)
	require.Len(t, proposedReqs, 1)
	require.Equal(t, overwritingReq, proposedReqs[0])
	require.NotEqual(t, initialReq, proposedReqs[0])
}

////////////////////////////////////////////////////////////////////////////////
// testEnv

// Setups testing environment and holds all the relevant info.
type testEnv struct {
	t                *testing.T
	ctx              context.Context
	ctxCancel        context.CancelFunc
	log              *logger.Logger
	utxoDB           *utxodb.UtxoDB
	governor         *cryptolib.KeyPair
	peeringURLs      []string
	peerIdentities   []*cryptolib.KeyPair
	peerPubKeys      []*cryptolib.PublicKey
	peeringNetwork   *testutil.PeeringNetwork
	networkProviders []peering.NetworkProvider
	tcl              *testchain.TestChainLedger
	cmtAddress       iotago.Address
	chainID          isc.ChainID
	originAO         *isc.AliasOutputWithID
	mempools         []mempool.Mempool
	stores           []state.Store
}

func newEnv(t *testing.T, n, f int, reliable bool) *testEnv {
	te := &testEnv{t: t}
	te.ctx, te.ctxCancel = context.WithCancel(context.Background())
	te.log = testlogger.NewLogger(t)
	//
	// Create ledger accounts.
	te.utxoDB = utxodb.New(utxodb.DefaultInitParams())
	te.governor = cryptolib.NewKeyPair()
	_, err := te.utxoDB.GetFundsFromFaucet(te.governor.Address())
	require.NoError(t, err)
	//
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
		netLogger := testlogger.WithLevel(te.log.Named("Network"), logger.LevelInfo, false)
		networkBehaviour = testutil.NewPeeringNetUnreliable(80, 20, 10*time.Millisecond, 200*time.Millisecond, netLogger)
	}
	te.peeringNetwork = testutil.NewPeeringNetwork(
		te.peeringURLs, te.peerIdentities, 10000,
		networkBehaviour,
		testlogger.WithLevel(te.log, logger.LevelWarn, false),
	)
	te.networkProviders = te.peeringNetwork.NetworkProviders()
	te.cmtAddress, _ = testpeers.SetupDkgTrivial(t, n, f, te.peerIdentities, nil)
	te.tcl = testchain.NewTestChainLedger(t, te.utxoDB, te.governor)
	_, te.originAO, te.chainID = te.tcl.MakeTxChainOrigin(te.cmtAddress)
	//
	// Initialize the nodes.
	te.mempools = make([]mempool.Mempool, len(te.peerIdentities))
	te.stores = make([]state.Store, len(te.peerIdentities))
	for i := range te.peerIdentities {
		te.stores[i] = state.NewStoreWithUniqueWriteMutex(mapdb.NewMapDB())
		_, err := origin.InitChainByAliasOutput(te.stores[i], te.originAO)
		require.NoError(t, err)
		chainMetrics := metrics.NewChainMetricsProvider().GetChainMetrics(isc.EmptyChainID())
		te.mempools[i] = mempool.New(
			te.ctx,
			te.chainID,
			te.peerIdentities[i],
			te.networkProviders[i],
			te.log.Named(fmt.Sprintf("N#%v", i)),
			chainMetrics.Mempool,
			chainMetrics.Pipe,
			chain.NewEmptyChainListener(),
		)
	}
	return te
}

func (te *testEnv) stateForAO(i int, ao *isc.AliasOutputWithID) state.State {
	l1Commitment, err := transaction.L1CommitmentFromAliasOutput(ao.GetAliasOutput())
	require.NoError(te.t, err)
	st, err := te.stores[i].StateByTrieRoot(l1Commitment.TrieRoot())
	require.NoError(te.t, err)
	return st
}

func (te *testEnv) close() {
	te.ctxCancel()
	te.peeringNetwork.Close()
	te.log.Sync()
}

////////////////////////////////////////////////////////////////////////////////

func getRequestsOnLedger(t *testing.T, chainAddress iotago.Address, amount int, f ...func(int, *isc.RequestParameters)) []isc.OnLedgerRequest {
	result := make([]isc.OnLedgerRequest, amount)
	for i := range result {
		requestParams := isc.RequestParameters{
			TargetAddress: chainAddress,
			Assets:        nil,
			Metadata: &isc.SendMetadata{
				TargetContract: isc.Hn("dummyTargetContract"),
				EntryPoint:     isc.Hn("dummyEP"),
				Params:         dict.New(),
				Allowance:      nil,
				GasBudget:      1000,
			},
			AdjustToMinimumStorageDeposit: true,
		}
		if len(f) == 1 {
			f[0](i, &requestParams)
		}
		output := transaction.BasicOutputFromPostData(
			tpkg.RandEd25519Address(),
			isc.EmptyContractIdentity(),
			requestParams,
		)
		outputID := tpkg.RandOutputID(uint16(i))
		var err error
		result[i], err = isc.OnLedgerFromUTXO(output, outputID)
		require.NoError(t, err)
	}
	return result
}
