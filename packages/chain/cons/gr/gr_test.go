// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package consGR_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/kvstore/mapdb"
	"github.com/iotaledger/hive.go/logger"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/contracts/native/inccounter"
	"github.com/iotaledger/wasp/packages/chain/cmtLog"
	consGR "github.com/iotaledger/wasp/packages/chain/cons/gr"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/testutil/testchain"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/packages/testutil/testpeers"
	"github.com/iotaledger/wasp/packages/utxodb"
	"github.com/iotaledger/wasp/packages/vm/core/coreprocessors"
	"github.com/iotaledger/wasp/packages/vm/processors"
	"github.com/iotaledger/wasp/packages/vm/vmcontext"
)

func TestBasic(t *testing.T) {
	t.Parallel()
	type test struct {
		n        int
		f        int
		reliable bool
	}
	tests := []test{
		{n: 1, f: 0, reliable: true},  // Low N
		{n: 2, f: 0, reliable: true},  // Low N
		{n: 3, f: 0, reliable: true},  // Low N
		{n: 4, f: 1, reliable: true},  // Minimal robust config.
		{n: 10, f: 3, reliable: true}, // Typical config.
	}
	if !testing.Short() {
		tests = append(tests,
			test{n: 4, f: 1, reliable: false},  // Minimal robust config.
			test{n: 10, f: 3, reliable: false}, // Typical config.
			test{n: 31, f: 10, reliable: true}, // Large cluster, reliable - to make test faster.
		)
	}
	for _, tst := range tests {
		t.Run(
			fmt.Sprintf("N=%v,F=%v,Reliable=%v", tst.n, tst.f, tst.reliable),
			func(tt *testing.T) { testGeneric(tt, tst.n, tst.f, tst.reliable) },
		)
	}
}

func testGeneric(t *testing.T, n, f int, reliable bool) {
	t.Parallel()
	log := testlogger.NewLogger(t)
	defer log.Sync()
	//
	// Create ledger accounts.
	utxoDB := utxodb.New(utxodb.DefaultInitParams())
	governor := cryptolib.NewKeyPair()
	originator := cryptolib.NewKeyPair()
	_, err := utxoDB.GetFundsFromFaucet(governor.Address())
	require.NoError(t, err)
	_, err = utxoDB.GetFundsFromFaucet(originator.Address())
	require.NoError(t, err)
	//
	// Create a fake network and keys for the tests.
	peeringURL, peerIdentities := testpeers.SetupKeys(uint16(n))
	peerPubKeys := make([]*cryptolib.PublicKey, len(peerIdentities))
	for i := range peerPubKeys {
		peerPubKeys[i] = peerIdentities[i].GetPublicKey()
	}
	var networkBehaviour testutil.PeeringNetBehavior
	if reliable {
		networkBehaviour = testutil.NewPeeringNetReliable(log)
	} else {
		netLogger := testlogger.WithLevel(log.Named("Network"), logger.LevelInfo, false)
		networkBehaviour = testutil.NewPeeringNetUnreliable(80, 20, 10*time.Millisecond, 200*time.Millisecond, netLogger)
	}
	var peeringNetwork *testutil.PeeringNetwork = testutil.NewPeeringNetwork(
		peeringURL, peerIdentities, 10000,
		networkBehaviour,
		testlogger.WithLevel(log, logger.LevelWarn, false),
	)
	defer peeringNetwork.Close()
	var networkProviders []peering.NetworkProvider = peeringNetwork.NetworkProviders()
	cmtAddress, dkShareProviders := testpeers.SetupDkgTrivial(t, n, f, peerIdentities, nil)
	//
	// Initialize the DSS subsystem in each node / chain.
	nodes := make([]*consGR.ConsGr, len(peerIdentities))
	mempools := make([]*testMempool, len(peerIdentities))
	stateMgrs := make([]*testStateMgr, len(peerIdentities))
	procConfig := coreprocessors.NewConfigWithCoreContracts().WithNativeContracts(inccounter.Processor)
	tcl := testchain.NewTestChainLedger(t, utxoDB, governor, originator)
	originAO, chainID := tcl.MakeTxChainOrigin(cmtAddress)
	chainInitReqs := tcl.MakeTxChainInit()
	ctx, ctxCancel := context.WithCancel(context.Background())
	logIndex := cmtLog.LogIndex(0)
	for i := range peerIdentities {
		procCache := processors.MustNew(procConfig)
		dkShare, err := dkShareProviders[i].LoadDKShare(cmtAddress)
		require.NoError(t, err)
		chainStore := state.InitChainStore(mapdb.NewMapDB())
		mempools[i] = newTestMempool(t)
		stateMgrs[i] = newTestStateMgr(t, chainStore)
		nodes[i] = consGR.New(
			ctx, chainID, chainStore, dkShare, &logIndex, peerIdentities[i],
			procCache, mempools[i], stateMgrs[i],
			networkProviders[i],
			1*time.Minute, // RecoverTimeout
			1*time.Second, // RedeliveryPeriod
			5*time.Second, // PrintStatusPeriod
			log.Named(fmt.Sprintf("N#%v", i)),
		)
	}
	//
	// Start the consensus in all nodes.
	outputChs := make([]chan *consGR.Output, len(nodes))
	for i, n := range nodes {
		outputCh := make(chan *consGR.Output, 1)
		outputChs[i] = outputCh
		n.Input(originAO, func(o *consGR.Output) { outputCh <- o }, func() {})
	}
	//
	// Provide data from Mempool and StateMgr.
	for i := range nodes {
		nodes[i].Time(time.Now())
		mempools[i].addRequests(originAO.OutputID(), chainInitReqs)
		stateMgrs[i].addOriginState(originAO)
	}
	//
	// Wait for outputs.
	var firstOutput *consGR.Output
	for _, outputCh := range outputChs {
		output := <-outputCh
		require.NotNil(t, output)
		if firstOutput == nil {
			firstOutput = output
		}
		require.Equal(t, firstOutput.Result.Transaction, output.Result.Transaction)
	}
	ctxCancel()
}

////////////////////////////////////////////////////////////////////////////////
// testMempool

type testMempool struct {
	t          *testing.T
	lock       *sync.Mutex
	reqsByAO   map[iotago.OutputID][]isc.Request
	allReqs    []isc.Request
	qProposals map[iotago.OutputID]chan []*isc.RequestRef
	qRequests  []*testMempoolReqQ
}

type testMempoolReqQ struct {
	refs []*isc.RequestRef
	resp chan []isc.Request
}

func newTestMempool(t *testing.T) *testMempool {
	return &testMempool{
		t:          t,
		lock:       &sync.Mutex{},
		reqsByAO:   map[iotago.OutputID][]isc.Request{},
		allReqs:    []isc.Request{},
		qProposals: map[iotago.OutputID]chan []*isc.RequestRef{},
		qRequests:  []*testMempoolReqQ{},
	}
}

func (tmp *testMempool) addRequests(aliasOutputID iotago.OutputID, requests []isc.Request) {
	tmp.lock.Lock()
	defer tmp.lock.Unlock()
	tmp.reqsByAO[aliasOutputID] = requests
	tmp.allReqs = append(tmp.allReqs, requests...)
	tmp.tryRespondProposalQueries()
	tmp.tryRespondRequestQueries()
}

func (tmp *testMempool) tryRespondProposalQueries() {
	for ao, resp := range tmp.qProposals {
		if reqs, ok := tmp.reqsByAO[ao]; ok {
			resp <- isc.RequestRefsFromRequests(reqs)
			close(resp)
			delete(tmp.qProposals, ao)
		}
	}
}

func (tmp *testMempool) tryRespondRequestQueries() {
	remaining := []*testMempoolReqQ{}
	for _, query := range tmp.qRequests {
		found := []isc.Request{}
		for _, ref := range query.refs {
			for rIndex := range tmp.allReqs {
				if ref.IsFor(tmp.allReqs[rIndex]) {
					found = append(found, tmp.allReqs[rIndex])
					break
				}
			}
		}
		if len(found) == len(query.refs) {
			query.resp <- found
			close(query.resp)
			continue
		}
		remaining = append(remaining, query)
	}
	tmp.qRequests = remaining
}

func (tmp *testMempool) ConsensusProposalsAsync(ctx context.Context, aliasOutput *isc.AliasOutputWithID) <-chan []*isc.RequestRef {
	tmp.lock.Lock()
	defer tmp.lock.Unlock()
	outputID := aliasOutput.OutputID()
	resp := make(chan []*isc.RequestRef, 1)
	tmp.qProposals[outputID] = resp
	tmp.tryRespondProposalQueries()
	return resp
}

func (tmp *testMempool) ConsensusRequestsAsync(ctx context.Context, requestRefs []*isc.RequestRef) <-chan []isc.Request {
	tmp.lock.Lock()
	defer tmp.lock.Unlock()
	resp := make(chan []isc.Request, 1)
	tmp.qRequests = append(tmp.qRequests, &testMempoolReqQ{resp: resp, refs: requestRefs})
	tmp.tryRespondRequestQueries()
	return resp
}

////////////////////////////////////////////////////////////////////////////////
// testStateMgr

type testStateMgr struct {
	t          *testing.T
	lock       *sync.Mutex
	chainStore state.Store
	states     map[hashing.HashValue]state.State
	qProposal  map[hashing.HashValue]chan interface{}
	qDecided   map[hashing.HashValue]chan state.State
}

func newTestStateMgr(t *testing.T, chainStore state.Store) *testStateMgr {
	return &testStateMgr{
		t:          t,
		lock:       &sync.Mutex{},
		chainStore: chainStore,
		states:     map[hashing.HashValue]state.State{},
		qProposal:  map[hashing.HashValue]chan interface{}{},
		qDecided:   map[hashing.HashValue]chan state.State{},
	}
}

func (tsm *testStateMgr) addOriginState(originAO *isc.AliasOutputWithID) {
	chainState, err := tsm.chainStore.StateByTrieRoot(state.OriginL1Commitment().TrieRoot())
	require.NoError(tsm.t, err)
	tsm.addState(originAO, chainState)
}

func (tsm *testStateMgr) addState(aliasOutput *isc.AliasOutputWithID, chainState state.State) { // TODO: Why is it not called from other places???
	tsm.lock.Lock()
	defer tsm.lock.Unlock()
	hash := commitmentHashFromAO(aliasOutput)
	tsm.states[hash] = chainState
	tsm.tryRespond(hash)
}

func (tsm *testStateMgr) ConsensusStateProposal(ctx context.Context, aliasOutput *isc.AliasOutputWithID) <-chan interface{} {
	tsm.lock.Lock()
	defer tsm.lock.Unlock()
	resp := make(chan interface{}, 1)
	hash := commitmentHashFromAO(aliasOutput)
	tsm.qProposal[hash] = resp
	tsm.tryRespond(hash)
	return resp
}

// State manager has to ensure all the data needed for the specified alias
// output (presented as aliasOutputID+stateCommitment) is present in the DB.
func (tsm *testStateMgr) ConsensusDecidedState(ctx context.Context, aliasOutput *isc.AliasOutputWithID) <-chan state.State {
	tsm.lock.Lock()
	defer tsm.lock.Unlock()
	resp := make(chan state.State, 1)
	stateCommitment, err := vmcontext.L1CommitmentFromAliasOutput(aliasOutput.GetAliasOutput())
	if err != nil {
		panic(err)
	}
	hash := commitmentHash(stateCommitment)
	tsm.qDecided[hash] = resp
	tsm.tryRespond(hash)
	return resp
}

func (tsm *testStateMgr) ConsensusProducedBlock(ctx context.Context, block state.StateDraft) <-chan error {
	tsm.lock.Lock()
	defer tsm.lock.Unlock()
	resp := make(chan error, 1)
	resp <- nil // We don't save it in the test for now, just respond it is already saved.
	close(resp)
	return resp
}

func (tsm *testStateMgr) tryRespond(hash hashing.HashValue) {
	s, ok := tsm.states[hash]
	if !ok {
		return
	}
	if qProposal, ok := tsm.qProposal[hash]; ok {
		qProposal <- nil
		close(qProposal)
		delete(tsm.qProposal, hash)
	}
	if qDecided, ok := tsm.qDecided[hash]; ok {
		qDecided <- s
		close(qDecided)
		delete(tsm.qDecided, hash)
	}
}

func commitmentHashFromAO(aliasOutput *isc.AliasOutputWithID) hashing.HashValue {
	commitment, err := vmcontext.L1CommitmentFromAliasOutput(aliasOutput.GetAliasOutput())
	if err != nil {
		panic(err)
	}
	return commitmentHash(commitment)
}

func commitmentHash(stateCommitment *state.L1Commitment) hashing.HashValue {
	return hashing.HashDataBlake2b(stateCommitment.Bytes())
}
