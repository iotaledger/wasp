// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chain_test

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/kvstore/mapdb"
	"github.com/iotaledger/hive.go/logger"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/contracts/native/inccounter"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/chain/statemanager/smGPA/smGPAUtils"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/metrics/nodeconnmetrics"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/shutdown"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/testutil"
	testparameters "github.com/iotaledger/wasp/packages/testutil/parameters"
	"github.com/iotaledger/wasp/packages/testutil/testchain"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/packages/testutil/testpeers"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/packages/utxodb"
	"github.com/iotaledger/wasp/packages/vm/core/coreprocessors"
)

type tc struct {
	n        int
	f        int
	reliable bool
	timeout  time.Duration
}

func TestNodeBasic(t *testing.T) {
	t.Parallel()
	tests := []tc{
		{n: 1, f: 0, reliable: true, timeout: 10 * time.Second},   // Low N
		{n: 2, f: 0, reliable: true, timeout: 20 * time.Second},   // Low N
		{n: 3, f: 0, reliable: true, timeout: 30 * time.Second},   // Low N
		{n: 4, f: 0, reliable: true, timeout: 40 * time.Second},   // Minimal robust config.
		{n: 4, f: 1, reliable: true, timeout: 50 * time.Second},   // Minimal robust config.
		{n: 10, f: 3, reliable: true, timeout: 130 * time.Second}, // Typical config.
	}
	if !testing.Short() {
		tests = append(tests,
			// TODO these "unreliable" tests are crazy, they either succeed in 10~20s or run forever...
			// tc{n: 4, f: 1, reliable: false,  timeout: 5*time.Minute},  // Minimal robust config.
			// tc{n: 10, f: 3, reliable: false,  timeout: 5*time.Minute}, // Typical config.
			tc{n: 31, f: 10, reliable: true, timeout: 5 * time.Minute}, // Large cluster, reliable - to make test faster.
		)
	}
	for _, tst := range tests {
		t.Run(
			fmt.Sprintf("N=%v,F=%v,Reliable=%v", tst.n, tst.f, tst.reliable),
			func(tt *testing.T) { testNodeBasic(tt, tst.n, tst.f, tst.reliable, tst.timeout) },
		)
	}
}

//nolint:gocyclo
func testNodeBasic(t *testing.T, n, f int, reliable bool, timeout time.Duration) {
	t.Parallel()
	te := newEnv(t, n, f, reliable)
	defer te.close()

	ctxTimeout, ctxTimeoutCancel := context.WithTimeout(te.ctx, timeout)
	defer ctxTimeoutCancel()

	te.log.Debugf("All started.")
	for _, tnc := range te.nodeConns {
		tnc.waitAttached()
	}
	te.log.Debugf("All attached to node conns.")
	go func() {
		for {
			if te.ctx.Err() != nil {
				return
			}
			for _, tnc := range te.nodeConns {
				tnc.recvMilestone(time.Now())
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()

	deployBaseAnchor, deployBaseAONoID, err := transaction.GetAnchorFromTransaction(te.originTx)
	require.NoError(t, err)
	deployBaseAO := isc.NewAliasOutputWithID(deployBaseAONoID, deployBaseAnchor.OutputID)
	for _, tnc := range te.nodeConns {
		tnc.recvAliasOutput(
			isc.NewOutputInfo(deployBaseAO.OutputID(), deployBaseAO.GetAliasOutput(), iotago.TransactionID{}),
		)
	}

	sendAndAwait := func(reqs []isc.Request, expectedBlockIndex int, desc string) {
		for _, tnc := range te.nodeConns {
			for _, req := range reqs {
				onLedgerRequest := req.(isc.OnLedgerRequest)
				tnc.recvRequestCB(
					isc.NewOutputInfo(onLedgerRequest.ID().OutputID(), onLedgerRequest.Output(), iotago.TransactionID{}),
				)
			}
		}
		awaitRequestsProcessed(ctxTimeout, te, reqs, desc)
		awaitPredicate(te, ctxTimeout, fmt.Sprintf("len(tnc.published) >= %d", expectedBlockIndex), func() bool {
			for _, tnc := range te.nodeConns {
				if len(tnc.published) < expectedBlockIndex {
					return false
				}
			}
			return true
		})
	}

	//
	// Create SC Client account with some deposit
	scClient := cryptolib.NewKeyPair()
	_, err = te.utxoDB.GetFundsFromFaucet(scClient.Address(), 150_000_000)
	require.NoError(t, err)
	depositReqs := te.tcl.MakeTxAccountsDeposit(scClient)
	sendAndAwait(depositReqs, 1, "depositReqs")

	//
	// Deploy a contract, wait for a confirming TX.
	deployReqs := te.tcl.MakeTxDeployIncCounterContract()
	sendAndAwait(deployReqs, 2, "deployReqs")

	//
	// Invoke off-ledger requests on the contract, wait for the counter to reach the expected value.
	// We only send the requests to the first node. Mempool has to disseminate them.
	incCount := 10
	incRequests := make([]isc.Request, incCount)
	for i := 0; i < incCount; i++ {
		scRequest := isc.NewOffLedgerRequest(
			te.chainID,
			inccounter.Contract.Hname(),
			inccounter.FuncIncCounter.Hname(),
			dict.New(), uint64(i),
		).WithGasBudget(2000000).Sign(scClient)
		te.nodes[0].ReceiveOffLedgerRequest(scRequest, scClient.GetPublicKey())
		incRequests[i] = scRequest
	}

	// Check if all requests were processed.
	awaitRequestsProcessed(ctxTimeout, te, incRequests, "incRequests")

	// assert state
	for i, node := range te.nodes {
		for {
			latestState, err := node.LatestState(chain.ActiveOrCommittedState)
			require.NoError(t, err)
			cnt := inccounter.NewStateAccess(latestState).GetCounter()
			te.log.Debugf("Counter[node=%v]=%v", i, cnt)
			if cnt >= int64(incCount) {
				// TODO: Double-check with the published TX.
				/*
					latestTX := te.nodeConns[i].published[len(te.nodeConns[i].published)-1]
					_, latestAONoID, err := transaction.GetAnchorFromTransaction(latestTX)
					require.NoError(t, err)
					latestL1Commitment, err := vmcontext.L1CommitmentFromAliasOutput(latestAONoID)
					require.NoError(t, err)
					st, err := node.GetStateReader().StateByTrieRoot(latestL1Commitment.GetTrieRoot())
					require.NoError(t, err)
					require.GreaterOrEqual(t, incCount, inccounter.NewStateAccess(st).GetCounter())
				*/
				break
			}
			if reliable {
				continue
			}
			//
			// For the unreliable-network tests we have to retry the requests.
			// That's because the gossip in the mempool is primitive for now.
			for ii := 0; ii < incCount; ii++ {
				scRequest := isc.NewOffLedgerRequest(
					te.chainID,
					inccounter.Contract.Hname(),
					inccounter.FuncIncCounter.Hname(),
					dict.New(), uint64(ii),
				).WithGasBudget(20000).Sign(scClient)
				te.nodes[0].ReceiveOffLedgerRequest(scRequest, scClient.GetPublicKey())
			}
			time.Sleep(100 * time.Millisecond)
		}
		// Check if LastAliasOutput() works as expected.
		awaitPredicate(te, ctxTimeout, "LatestAliasOutput", func() bool {
			confirmedAO, activeAO := node.LatestAliasOutput()
			lastPublishedTX := te.nodeConns[i].published[len(te.nodeConns[i].published)-1]
			lastPublishedAO, err := isc.AliasOutputWithIDFromTx(lastPublishedTX, te.chainID.AsAddress())
			require.NoError(t, err)
			if !lastPublishedAO.Equals(confirmedAO) { // In this test we confirm outputs immediately.
				te.log.Debugf("lastPublishedAO(%v) != confirmedAO(%v)", lastPublishedAO, confirmedAO)
				return false
			}
			if !lastPublishedAO.Equals(activeAO) {
				te.log.Debugf("lastPublishedAO(%v) != activeAO(%v)", lastPublishedAO, activeAO)
				return false
			}
			return true
		})
	}
}

func awaitRequestsProcessed(ctx context.Context, te *testEnv, requests []isc.Request, desc string) {
	reqRefs := isc.RequestRefsFromRequests(requests)
	for i, node := range te.nodes {
		for reqNum, reqRef := range reqRefs {
			te.log.Debugf("Going to AwaitRequestProcessed %v at node=%v, req[%v]=%v...", desc, i, reqNum, reqRef.ID.String())

			await := func(confirmed bool) {
				rec := <-node.AwaitRequestProcessed(ctx, reqRef.ID, confirmed)
				if ctx.Err() != nil {
					te.t.Fatalf("awaitRequestsProcessed (%t) failed: %s, context timeout", confirmed, desc)
				}
				if rec.Error != nil {
					te.t.Fatalf("request processed with an error, %s", rec.Error.Error())
				}
			}

			await(false)
			await(true)
			te.log.Debugf("Going to AwaitRequestProcessed %v at node=%v, req[%v]=%v...Done", desc, i, reqNum, reqRef.ID.String())
		}
	}
}

//nolint:revive
func awaitPredicate(te *testEnv, ctx context.Context, desc string, predicate func() bool) {
	for {
		select {
		case <-ctx.Done():
			require.FailNowf(te.t, "awaitPredicate failed: %s", desc)
		default:
			if predicate() {
				te.log.Debugf("Predicate %v become true.", desc)
				return
			}
			te.log.Debugf("Predicate %v still false, will retry.", desc)
			time.Sleep(100 * time.Millisecond)
		}
	}
}

////////////////////////////////////////////////////////////////////////////////
// testNodeConn

type testNodeConn struct {
	t               *testing.T
	chainID         isc.ChainID
	published       []*iotago.Transaction
	recvRequestCB   chain.RequestOutputHandler
	recvAliasOutput chain.AliasOutputHandler
	recvMilestone   chain.MilestoneHandler
	attachWG        *sync.WaitGroup
}

func newTestNodeConn(t *testing.T) *testNodeConn {
	tnc := &testNodeConn{
		t:         t,
		published: []*iotago.Transaction{},
		attachWG:  &sync.WaitGroup{},
	}
	tnc.attachWG.Add(1)
	return tnc
}

func (tnc *testNodeConn) PublishTX(
	ctx context.Context,
	chainID isc.ChainID,
	tx *iotago.Transaction,
	callback chain.TxPostHandler,
) error {
	if tnc.chainID.Empty() {
		tnc.t.Error("NodeConn::PublishTX before attach.")
	}
	if !tnc.chainID.Equals(chainID) {
		tnc.t.Error("unexpected chain id")
	}
	txID, err := tx.ID()
	require.NoError(tnc.t, err)
	existing := lo.ContainsBy(tnc.published, func(publishedTX *iotago.Transaction) bool {
		publishedID, err2 := publishedTX.ID()
		require.NoError(tnc.t, err2)
		return txID == publishedID
	})
	if existing {
		tnc.t.Logf("Already seen a TX with ID=%v", txID)
		return nil
	}
	tnc.published = append(tnc.published, tx)
	callback(tx, true)

	stateAnchor, aoNoID, err := transaction.GetAnchorFromTransaction(tx)
	if err != nil {
		return err
	}
	tnc.recvAliasOutput(
		isc.NewOutputInfo(stateAnchor.OutputID, aoNoID, iotago.TransactionID{}),
	)

	return nil
}

func (tnc *testNodeConn) AttachChain(
	ctx context.Context,
	chainID isc.ChainID,
	recvRequestCB chain.RequestOutputHandler,
	recvAliasOutput chain.AliasOutputHandler,
	recvMilestone chain.MilestoneHandler,
) {
	if !tnc.chainID.Empty() {
		tnc.t.Error("duplicate attach")
	}
	tnc.chainID = chainID
	tnc.recvRequestCB = recvRequestCB
	tnc.recvAliasOutput = recvAliasOutput
	tnc.recvMilestone = recvMilestone
	tnc.attachWG.Done()
}

func (tnc *testNodeConn) Run(ctx context.Context) error {
	panic("should be unused in test")
}

func (tnc *testNodeConn) waitAttached() {
	tnc.attachWG.Wait()
}

func (tnc *testNodeConn) WaitUntilInitiallySynced(ctx context.Context) error {
	panic("should be unused in test")
}

func (tnc *testNodeConn) GetBech32HRP() iotago.NetworkPrefix {
	return testparameters.GetBech32HRP()
}

func (tnc *testNodeConn) GetL1Params() *parameters.L1Params {
	return testparameters.GetL1ParamsForTesting()
}

func (tnc *testNodeConn) GetL1ProtocolParams() *iotago.ProtocolParameters {
	return testparameters.GetL1ProtocolParamsForTesting()
}

func (tnc *testNodeConn) GetMetrics() nodeconnmetrics.NodeConnectionMetrics {
	panic("should be unused in test")
}

////////////////////////////////////////////////////////////////////////////////
// testEnv

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
	originTx         *iotago.Transaction
	nodeConns        []*testNodeConn
	nodes            []chain.Chain
}

func newEnv(t *testing.T, n, f int, reliable bool) *testEnv {
	te := &testEnv{t: t}
	te.ctx, te.ctxCancel = context.WithCancel(context.Background())
	te.log = testlogger.NewLogger(t).Named(fmt.Sprintf("%04d", rand.Intn(10000))) // For test instance ID.
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
	var dkShareProviders []registry.DKShareRegistryProvider
	te.cmtAddress, dkShareProviders = testpeers.SetupDkgTrivial(t, n, f, te.peerIdentities, nil)
	te.tcl = testchain.NewTestChainLedger(t, te.utxoDB, te.originator)
	te.originTx, te.originAO, te.chainID = te.tcl.MakeTxChainOrigin(te.cmtAddress)
	//
	// Initialize the nodes.
	te.nodeConns = make([]*testNodeConn, len(te.peerIdentities))
	te.nodes = make([]chain.Chain, len(te.peerIdentities))
	require.NoError(t, err)
	for i := range te.peerIdentities {
		te.nodeConns[i] = newTestNodeConn(t)
		log := te.log.Named(fmt.Sprintf("N#%v", i))
		te.nodes[i], err = chain.New(
			te.ctx,
			te.chainID,
			state.NewStore(mapdb.NewMapDB()),
			te.nodeConns[i],
			te.peerIdentities[i],
			coreprocessors.NewConfigWithCoreContracts().WithNativeContracts(inccounter.Processor),
			dkShareProviders[i],
			testutil.NewConsensusStateRegistry(),
			smGPAUtils.NewMockedTestBlockWAL(),
			chain.NewEmptyChainListener(),
			[]*cryptolib.PublicKey{}, // Access nodes.
			te.networkProviders[i],
			shutdown.NewCoordinator("test", log),
			log,
		)
		require.NoError(t, err)
		te.nodes[i].ServersUpdated(te.peerPubKeys)
	}
	te.log = te.log.Named("TC")
	return te
}

func (te *testEnv) close() {
	te.ctxCancel()
	te.peeringNetwork.Close()
	te.log.Sync()
}
