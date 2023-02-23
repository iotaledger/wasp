// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package mempool_test

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/core/kvstore/mapdb"
	"github.com/iotaledger/hive.go/core/logger"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/tpkg"
	"github.com/iotaledger/wasp/contracts/native/inccounter"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/chain/mempool"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/testutil/testchain"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/packages/testutil/testpeers"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/packages/utxodb"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/core/coreprocessors"
	"github.com/iotaledger/wasp/packages/vm/processors"
	"github.com/iotaledger/wasp/packages/vm/runvm"
)

type tc struct {
	n        int
	f        int
	reliable bool
}

func TestBasic(t *testing.T) {
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
			func(tt *testing.T) { testBasic(tt, tst.n, tst.f, tst.reliable) },
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
func testBasic(t *testing.T, n, f int, reliable bool) {
	t.Parallel()
	te := newEnv(t, n, f, reliable)
	defer te.close()
	chainInitReqs := te.tcl.MakeTxChainInit()
	require.Len(t, chainInitReqs, 1)
	chainInitReq := chainInitReqs[0]
	//
	offLedgerReq := isc.NewOffLedgerRequest(isc.RandomChainID(), isc.Hn("foo"), isc.Hn("bar"), dict.New(), 0).Sign(te.governor)
	t.Log("Sending off-ledger request")
	chosenMempool := rand.Intn(len(te.mempools))
	te.mempools[chosenMempool].ReceiveOffLedgerRequest(offLedgerReq)
	te.mempools[chosenMempool].ReceiveOffLedgerRequest(offLedgerReq) // Check for duplicate receives.
	t.Log("Sending on-ledger request")
	for _, node := range te.mempools {
		node.ReceiveOnLedgerRequest(chainInitReq.(isc.OnLedgerRequest))
	}
	t.Log("ServerNodesUpdated")
	tangleTime := time.Now()
	for _, node := range te.mempools {
		node.ServerNodesUpdated(te.peerPubKeys, te.peerPubKeys)
		node.TangleTimeUpdated(tangleTime)
	}
	t.Log("TrackNewChainHead")
	for i, node := range te.mempools {
		node.TrackNewChainHead(te.stateForAO(i, te.originAO), nil, te.originAO, []state.Block{}, []state.Block{})
	}
	t.Log("Ask for proposals")
	proposals := make([]<-chan []*isc.RequestRef, len(te.mempools))
	for i, node := range te.mempools {
		proposals[i] = node.ConsensusProposalsAsync(te.ctx, te.originAO)
	}
	t.Log("Wait for proposals and ask for decided requests")
	decided := make([]<-chan []isc.Request, len(te.mempools))
	for i, node := range te.mempools {
		proposal := <-proposals[i]
		require.True(t, len(proposal) == 1 || len(proposal) == 2)
		decided[i] = node.ConsensusRequestsAsync(te.ctx, isc.RequestRefsFromRequests([]isc.Request{chainInitReq, offLedgerReq}))
	}
	t.Log("Wait for decided requests")
	for i := range te.mempools {
		nodeDecidedReqs := <-decided[i]
		require.Len(t, nodeDecidedReqs, 2)
	}
	//
	// Make a block consuming those 2 requests.
	store := te.stores[0]
	vmTask := &vm.VMTask{
		Processors:             processors.MustNew(coreprocessors.NewConfigWithCoreContracts().WithNativeContracts(inccounter.Processor)),
		AnchorOutput:           te.originAO.GetAliasOutput(),
		AnchorOutputID:         te.originAO.OutputID(),
		Store:                  store,
		Requests:               []isc.Request{chainInitReq, offLedgerReq},
		TimeAssumption:         tangleTime,
		Entropy:                hashing.HashDataBlake2b([]byte{2, 1, 7}),
		ValidatorFeeTarget:     te.chainID.CommonAccount(),
		EstimateGasMode:        false,
		EnableGasBurnLogging:   false,
		MaintenanceModeEnabled: false,
		Log:                    te.log.Named("VM"),
	}
	require.NoError(t, runvm.NewVMRunner().Run(vmTask))
	block := store.Commit(vmTask.StateDraft)
	chainState, err := store.StateByTrieRoot(block.TrieRoot())
	require.NoError(t, err)
	//
	// Check if block has both requests as consumed.
	receipts, err := blocklog.RequestReceiptsFromBlock(block)
	require.NoError(t, err)
	require.Len(t, receipts, 2)
	blockReqs := []isc.Request{}
	for i := range receipts {
		blockReqs = append(blockReqs, receipts[i].Request)
	}
	require.Contains(t, blockReqs, chainInitReq)
	require.Contains(t, blockReqs, offLedgerReq)
	nextAO, _ := te.tcl.FakeTX(te.originAO, te.cmtAddress)
	//
	// Ask proposals for the next
	proposals = make([]<-chan []*isc.RequestRef, len(te.mempools))
	for i := range te.mempools {
		proposals[i] = te.mempools[i].ConsensusProposalsAsync(te.ctx, nextAO) // Intentionally invalid order (vs TrackNewChainHead).
		te.mempools[i].TrackNewChainHead(chainState, te.originAO, nextAO, []state.Block{block}, []state.Block{})
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
	offLedgerReq2 := isc.NewOffLedgerRequest(isc.RandomChainID(), isc.Hn("foo"), isc.Hn("bar"), dict.New(), 1).Sign(te.governor)
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

func testTimeLock(t *testing.T, n, f int, reliable bool) { //nolint: gocyclo
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
		reqs := <-mp.ConsensusProposalsAsync(te.ctx, te.originAO)
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
		reqs := <-mp.ConsensusProposalsAsync(te.ctx, te.originAO)
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
		reqs := <-mp.ConsensusProposalsAsync(te.ctx, te.originAO)
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
		reqs := <-mp.ConsensusProposalsAsync(te.ctx, te.originAO)
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
		reqs := <-mp.ConsensusProposalsAsync(te.ctx, te.originAO)
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
		reqs := <-mp.ConsensusProposalsAsync(te.ctx, te.originAO)
		require.Len(t, reqs, 1)
		require.Contains(t, reqs, reqRefs[0])
	}
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
	originator       *cryptolib.KeyPair
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
	te.originator = cryptolib.NewKeyPair()
	_, err := te.utxoDB.GetFundsFromFaucet(te.governor.Address())
	require.NoError(t, err)
	_, err = te.utxoDB.GetFundsFromFaucet(te.originator.Address())
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
	te.tcl = testchain.NewTestChainLedger(t, te.utxoDB, te.governor, te.originator)
	te.originAO, te.chainID = te.tcl.MakeTxChainOrigin(te.cmtAddress)
	//
	// Initialize the nodes.
	te.mempools = make([]mempool.Mempool, len(te.peerIdentities))
	te.stores = make([]state.Store, len(te.peerIdentities))
	for i := range te.peerIdentities {
		te.stores[i] = state.InitChainStore(mapdb.NewMapDB())
		te.mempools[i] = mempool.New(
			te.ctx,
			te.chainID,
			te.peerIdentities[i],
			te.networkProviders[i],
			te.log.Named(fmt.Sprintf("N#%v", i)),
			&MockMempoolMetrics{},
			chain.NewEmptyChainListener(),
		)
	}
	return te
}

func (te *testEnv) stateForAO(i int, ao *isc.AliasOutputWithID) state.State {
	l1Commitment, err := state.L1CommitmentFromAliasOutput(ao.GetAliasOutput())
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
// MockMempoolMetrics

type MockMempoolMetrics struct {
	mock.Mock
	offLedgerRequestCounter int
	onLedgerRequestCounter  int
	processedRequestCounter int
}

func (m *MockMempoolMetrics) IncRequestsReceived(req isc.Request) {
	if req.IsOffLedger() {
		m.offLedgerRequestCounter++
	} else {
		m.onLedgerRequestCounter++
	}
}

func (m *MockMempoolMetrics) IncRequestsProcessed() {
	m.processedRequestCounter++
}

func (m *MockMempoolMetrics) SetRequestProcessingTime(reqID isc.RequestID, elapse time.Duration) {}

func (m *MockMempoolMetrics) IncBlocksPerChain() {}

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
			isc.Hn("dummySenderContract"),
			requestParams,
		)
		outputID := tpkg.RandOutputID(uint16(i))
		var err error
		result[i], err = isc.OnLedgerFromUTXO(output, outputID)
		require.NoError(t, err)
	}
	return result
}
