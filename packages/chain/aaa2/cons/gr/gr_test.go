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

	"github.com/iotaledger/hive.go/core/kvstore/mapdb"
	"github.com/iotaledger/hive.go/core/logger"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/contracts/native/inccounter"
	"github.com/iotaledger/wasp/packages/chain/aaa2/cmtLog"
	consGR "github.com/iotaledger/wasp/packages/chain/aaa2/cons/gr"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/testutil/testchain"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/packages/testutil/testpeers"
	"github.com/iotaledger/wasp/packages/utxodb"
	"github.com/iotaledger/wasp/packages/vm/core/coreprocessors"
	"github.com/iotaledger/wasp/packages/vm/processors"
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
	peerNetIDs, peerIdentities := testpeers.SetupKeys(uint16(n))
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
		peerNetIDs, peerIdentities, 10000,
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
		mempools[i] = newTestMempool(t)
		stateMgrs[i] = newTestStateMgr(t)
		nodes[i] = consGR.New(
			ctx, chainID, dkShare, &logIndex, peerIdentities[i],
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
	outputChs := make([]<-chan *consGR.Output, len(nodes))
	for i, n := range nodes {
		outputCh, _ := n.Input(originAO)
		outputChs[i] = outputCh
	}
	//
	// Provide data from Mempool and StateMgr.
	for i := range nodes {
		nodes[i].Time(time.Now())
		mempools[i].addRequests(originAO.OutputID(), chainInitReqs)
		stateMgrs[i].addOriginState(originAO, chainID)
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
		require.Equal(t, firstOutput.TX, output.TX)
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
	t         *testing.T
	lock      *sync.Mutex
	states    map[hashing.HashValue]*consGR.StateMgrDecidedState
	qProposal map[hashing.HashValue]chan interface{}
	qDecided  map[hashing.HashValue]chan *consGR.StateMgrDecidedState
}

func newTestStateMgr(t *testing.T) *testStateMgr {
	return &testStateMgr{
		t:         t,
		lock:      &sync.Mutex{},
		states:    map[hashing.HashValue]*consGR.StateMgrDecidedState{},
		qProposal: map[hashing.HashValue]chan interface{}{},
		qDecided:  map[hashing.HashValue]chan *consGR.StateMgrDecidedState{},
	}
}

func (tsm *testStateMgr) addOriginState(originAO *isc.AliasOutputWithID, chainID *isc.ChainID) {
	kvStore := mapdb.NewMapDB()
	stateSync := coreutil.NewChainStateSync()
	stateSync.SetSolidIndex(0)
	vsAccess, err := state.CreateOriginState(kvStore, chainID)
	require.NoError(tsm.t, err)
	tsm.addState(originAO, stateSync.GetSolidIndexBaseline(), vsAccess)
}

func (tsm *testStateMgr) addState(aliasOutput *isc.AliasOutputWithID, stateBaseline coreutil.StateBaseline, virtualStateAccess state.VirtualStateAccess) {
	tsm.lock.Lock()
	defer tsm.lock.Unlock()
	hash := commitmentHashFromAO(aliasOutput)
	tsm.states[hash] = &consGR.StateMgrDecidedState{
		AliasOutput:        aliasOutput,
		StateBaseline:      stateBaseline,
		VirtualStateAccess: virtualStateAccess,
	}
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
func (tsm *testStateMgr) ConsensusDecidedState(ctx context.Context, aliasOutputID *iotago.OutputID, stateCommitment *state.L1Commitment) <-chan *consGR.StateMgrDecidedState {
	tsm.lock.Lock()
	defer tsm.lock.Unlock()
	resp := make(chan *consGR.StateMgrDecidedState, 1)
	hash := commitmentHash(stateCommitment)
	tsm.qDecided[hash] = resp
	tsm.tryRespond(hash)
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
	commitment, err := state.L1CommitmentFromAliasOutput(aliasOutput.GetAliasOutput())
	if err != nil {
		panic(err)
	}
	return commitmentHash(commitment)
}

func commitmentHash(stateCommitment *state.L1Commitment) hashing.HashValue {
	return hashing.HashDataBlake2b(stateCommitment.Bytes())
}
