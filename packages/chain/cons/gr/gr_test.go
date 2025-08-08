// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package gr_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	hivelog "github.com/iotaledger/hive.go/log"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago/iotatest"
	"github.com/iotaledger/wasp/v2/packages/chain/cmtlog"
	consGR "github.com/iotaledger/wasp/v2/packages/chain/cons/gr"
	"github.com/iotaledger/wasp/v2/packages/coin"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/hashing"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/kvstore/mapdb"
	"github.com/iotaledger/wasp/v2/packages/metrics"
	"github.com/iotaledger/wasp/v2/packages/origin"
	"github.com/iotaledger/wasp/v2/packages/parameters"
	"github.com/iotaledger/wasp/v2/packages/parameters/parameterstest"
	"github.com/iotaledger/wasp/v2/packages/state"
	"github.com/iotaledger/wasp/v2/packages/testutil"
	"github.com/iotaledger/wasp/v2/packages/testutil/l1starter"
	"github.com/iotaledger/wasp/v2/packages/testutil/testchain"
	"github.com/iotaledger/wasp/v2/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/v2/packages/testutil/testpeers"

	"github.com/iotaledger/wasp/v2/packages/transaction"
	"github.com/iotaledger/wasp/v2/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/v2/packages/vm/core/coreprocessors"
	"github.com/iotaledger/wasp/v2/packages/vm/gas"
)

func TestGrBasic(t *testing.T) {
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
			func(tt *testing.T) {
				testGrBasic(tt, tst.n, tst.f, tst.reliable)
			},
		)
	}
}

func testGrBasic(t *testing.T, n, f int, reliable bool) {
	t.Parallel()
	log := testlogger.NewLogger(t)
	defer log.Shutdown()

	ctx, ctxCancel := context.WithCancel(context.Background())
	defer ctxCancel()

	//
	// Create ledger accounts. Requesting funds twice to get two coin objects (so we don't need to split one later)
	originator := cryptolib.NewKeyPair()
	err := iotaclient.RequestFundsFromFaucet(ctx, originator.Address().AsIotaAddress(), l1starter.Instance().FaucetURL())
	require.NoError(t, err)
	err = iotaclient.RequestFundsFromFaucet(ctx, originator.Address().AsIotaAddress(), l1starter.Instance().FaucetURL())
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
		netLogger := testlogger.WithLevel(log.NewChildLogger("Network"), hivelog.LevelInfo, false)
		networkBehaviour = testutil.NewPeeringNetUnreliable(80, 20, 10*time.Millisecond, 200*time.Millisecond, netLogger)
	}
	peeringNetwork := testutil.NewPeeringNetwork(
		peeringURL, peerIdentities, 10000,
		networkBehaviour,
		testlogger.WithLevel(log, hivelog.LevelWarning, false),
	)
	defer peeringNetwork.Close()
	networkProviders := peeringNetwork.NetworkProviders()
	cmtAddress, dkShareProviders := testpeers.SetupDkgTrivial(t, n, f, peerIdentities, nil)
	//
	// Initialize the DSS subsystem in each node / chain.
	nodes := make([]*consGR.ConsGr, len(peerIdentities))
	mempools := make([]*testMempool, len(peerIdentities))
	stateMgrs := make([]*testStateMgr, len(peerIdentities))
	procConfig := coreprocessors.NewConfigWithTestContracts()

	l1client := l1starter.Instance().L1Client()

	iscPackage, err := l1client.DeployISCContracts(ctx, cryptolib.SignerToIotaSigner(originator))
	require.NoError(t, err)

	tcl := testchain.NewTestChainLedger(t, originator, &iscPackage, l1client)

	anchor, anchorDeposit := tcl.MakeTxChainOrigin()
	gasCoin := &coin.CoinWithRef{
		Type:  coin.BaseTokenType,
		Value: coin.Value(100),
		Ref:   iotatest.RandomObjectRef(),
	}

	logIndex := cmtlog.LogIndex(0)
	chainMetricsProvider := metrics.NewChainMetricsProvider()
	for i := range peerIdentities {
		dkShare, err := dkShareProviders[i].LoadDKShare(cmtAddress)
		require.NoError(t, err)
		chainStore := state.NewStoreWithUniqueWriteMutex(mapdb.NewMapDB())
		_, err = origin.InitChainByStateMetadataBytes(chainStore, anchor.GetStateMetadata(), anchorDeposit, parameterstest.L1Mock)
		require.NoError(t, err)
		mempools[i] = newTestMempool(t)
		stateMgrs[i] = newTestStateMgr(t, chainStore)
		chainMetrics := chainMetricsProvider.GetChainMetrics(isc.EmptyChainID())
		nodes[i] = consGR.New(
			ctx, anchor.ChainID(), chainStore, dkShare, &logIndex, peerIdentities[i],
			procConfig, mempools[i], stateMgrs[i], newTestNodeConn(gasCoin),
			networkProviders[i],
			nil,
			accounts.CommonAccount(),
			1*time.Minute, // RecoverTimeout
			1*time.Second, // RedeliveryPeriod
			5*time.Second, // PrintStatusPeriod
			chainMetrics.Consensus,
			chainMetrics.Pipe,
			log.NewChildLogger(fmt.Sprintf("N#%v", i)),
		)
	}
	//
	// Start the consensus in all nodes.
	outputChs := make([]chan *consGR.Output, len(nodes))
	for i, n := range nodes {
		outputCh := make(chan *consGR.Output, 1)
		outputChs[i] = outputCh
		n.Input(anchor, func(o *consGR.Output) { outputCh <- o }, func() {})
	}

	time.Sleep(1 * time.Second)

	//
	// Provide data from Mempool and StateMgr.
	for i := range nodes {
		nodes[i].Time(time.Now())
		mempools[i].addRequests(anchor.GetObjectRef(), []isc.Request{
			isc.NewOffLedgerRequest(isc.NewMessage(isc.Hn("foo"), isc.Hn("bar"), nil), 0, gas.LimitsDefault.MaxGasPerRequest).Sign(originator),
		})
		stateMgrs[i].addOriginState(anchor)
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
}

////////////////////////////////////////////////////////////////////////////////
// testMempool

type anchorKey = string

func anchorKeyFromAnchor(anchor *isc.StateAnchor) anchorKey {
	return anchor.GetObjectRef().String()
}

func anchorKeyFromAnchorRef(objectRef *iotago.ObjectRef) anchorKey {
	return objectRef.String()
}

type testMempool struct {
	t          *testing.T
	lock       *sync.Mutex
	reqsByAO   map[anchorKey][]isc.Request
	allReqs    []isc.Request
	qProposals map[anchorKey]chan []*isc.RequestRef
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
		reqsByAO:   map[anchorKey][]isc.Request{},
		allReqs:    []isc.Request{},
		qProposals: map[anchorKey]chan []*isc.RequestRef{},
		qRequests:  []*testMempoolReqQ{},
	}
}

func (tmp *testMempool) addRequests(anchorRef *iotago.ObjectRef, requests []isc.Request) {
	tmp.lock.Lock()
	defer tmp.lock.Unlock()
	tmp.reqsByAO[anchorKeyFromAnchorRef(anchorRef)] = requests
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

func (tmp *testMempool) ConsensusProposalAsync(ctx context.Context, anchor *isc.StateAnchor, consensusID consGR.ConsensusID) <-chan []*isc.RequestRef {
	tmp.lock.Lock()
	defer tmp.lock.Unlock()
	resp := make(chan []*isc.RequestRef, 1)
	tmp.qProposals[anchorKeyFromAnchor(anchor)] = resp
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
	qProposal  map[hashing.HashValue]chan any
	qDecided   map[hashing.HashValue]chan state.State
}

func newTestStateMgr(t *testing.T, chainStore state.Store) *testStateMgr {
	return &testStateMgr{
		t:          t,
		lock:       &sync.Mutex{},
		chainStore: chainStore,
		states:     map[hashing.HashValue]state.State{},
		qProposal:  map[hashing.HashValue]chan any{},
		qDecided:   map[hashing.HashValue]chan state.State{},
	}
}

func (tsm *testStateMgr) addOriginState(originAO *isc.StateAnchor) {
	originAOStateMetadata, err := transaction.StateMetadataFromBytes(originAO.GetStateMetadata())
	require.NoError(tsm.t, err)
	chainState, err := tsm.chainStore.StateByTrieRoot(
		originAOStateMetadata.L1Commitment.TrieRoot(),
	)
	require.NoError(tsm.t, err)
	tsm.addState(originAO, chainState)
}

func (tsm *testStateMgr) addState(aliasOutput *isc.StateAnchor, chainState state.State) { // TODO: Why is it not called from other places???
	tsm.lock.Lock()
	defer tsm.lock.Unlock()
	hash := commitmentHashFromAO(aliasOutput)
	tsm.states[hash] = chainState
	tsm.tryRespond(hash)
}

func (tsm *testStateMgr) ConsensusStateProposal(ctx context.Context, aliasOutput *isc.StateAnchor) <-chan any {
	tsm.lock.Lock()
	defer tsm.lock.Unlock()
	resp := make(chan any, 1)
	hash := commitmentHashFromAO(aliasOutput)
	tsm.qProposal[hash] = resp
	tsm.tryRespond(hash)
	return resp
}

// State manager has to ensure all the data needed for the specified alias
// output (presented as aliasOutputID+stateCommitment) is present in the DB.
func (tsm *testStateMgr) ConsensusDecidedState(ctx context.Context, anchor *isc.StateAnchor) <-chan state.State {
	tsm.lock.Lock()
	defer tsm.lock.Unlock()
	resp := make(chan state.State, 1)
	stateCommitment, err := transaction.L1CommitmentFromAnchor(anchor)
	if err != nil {
		tsm.t.Fatal(err)
	}
	hash := commitmentHash(stateCommitment)
	tsm.qDecided[hash] = resp
	tsm.tryRespond(hash)
	return resp
}

func (tsm *testStateMgr) ConsensusProducedBlock(ctx context.Context, stateDraft state.StateDraft) <-chan state.Block {
	tsm.lock.Lock()
	defer tsm.lock.Unlock()
	resp := make(chan state.Block, 1)
	block := tsm.chainStore.Commit(stateDraft)
	resp <- block
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

type testNodeConnL1Info struct {
	gasCoins []*coin.CoinWithRef
	l1params *parameters.L1Params
}

func (tgi *testNodeConnL1Info) GetGasCoins() []*coin.CoinWithRef  { return tgi.gasCoins }
func (tgi *testNodeConnL1Info) GetL1Params() *parameters.L1Params { return tgi.l1params }

type testNodeConn struct {
	gasCoin *coin.CoinWithRef
}

var _ consGR.NodeConn = &testNodeConn{}

func newTestNodeConn(gasCoin *coin.CoinWithRef) *testNodeConn {
	return &testNodeConn{gasCoin: gasCoin}
}

func (t *testNodeConn) ConsensusL1InfoProposal(ctx context.Context, anchor *isc.StateAnchor) <-chan consGR.NodeConnL1Info {
	ch := make(chan consGR.NodeConnL1Info, 1)
	ch <- &testNodeConnL1Info{
		gasCoins: []*coin.CoinWithRef{t.gasCoin},
		l1params: parameterstest.L1Mock,
	}
	close(ch)
	return ch
}

func commitmentHashFromAO(anchor *isc.StateAnchor) hashing.HashValue {
	commitment, err := transaction.L1CommitmentFromAnchor(anchor)
	if err != nil {
		panic(err)
	}
	return commitmentHash(commitment)
}

func commitmentHash(stateCommitment *state.L1Commitment) hashing.HashValue {
	return hashing.HashDataBlake2b(stateCommitment.Bytes())
}
