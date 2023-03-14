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

	"github.com/iotaledger/hive.go/core/kvstore/mapdb"
	"github.com/iotaledger/hive.go/core/logger"
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
	if !testing.Short() { // TODO: Uncomment them when the CmtLog is fixed.
		tests = append(tests,
			tc{n: 4, f: 1, reliable: false},  // Minimal robust config.
			tc{n: 10, f: 3, reliable: false}, // Typical config.
			tc{n: 31, f: 10, reliable: true}, // Large cluster, reliable - to make test faster.
		)
	}
	for _, tst := range tests {
		t.Run(
			fmt.Sprintf("N=%v,F=%v,Reliable=%v", tst.n, tst.f, tst.reliable),
			func(tt *testing.T) { testBasic(tt, tst.n, tst.f, tst.reliable, 5*time.Minute) },
		)
	}
}

//nolint:gocyclo
func testBasic(t *testing.T, n, f int, reliable bool, timeout time.Duration) {
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
	initTime := time.Now()
	go func() {
		ticker := time.NewTicker(100 * time.Millisecond) // TODO: Maybe we have to pass initial time to all sub-components...
		closed := te.ctx.Done()
		for {
			select {
			case now := <-ticker.C:
				for _, tnc := range te.nodeConns {
					tnc.recvMilestone(now)
				}
			case <-closed:
				return
			}
		}
	}()
	//
	// Create the chain, post Chain Init, wait for the first block.
	initRequests := te.tcl.MakeTxChainInit()
	for _, tnc := range te.nodeConns {
		tnc.recvAliasOutput(
			isc.NewOutputInfo(te.originAO.OutputID(), te.originAO.GetAliasOutput(), iotago.TransactionID{}),
		)
		tnc.recvMilestone(initTime)
		for _, req := range initRequests {
			onLedgerRequest := req.(isc.OnLedgerRequest)
			tnc.recvRequestCB(
				isc.NewOutputInfo(onLedgerRequest.ID().OutputID(), onLedgerRequest.Output(), iotago.TransactionID{}),
			)
		}
	}
	awaitRequestsProcessed(te, ctxTimeout, initRequests, "initRequests")
	awaitPredicate(te, ctxTimeout, "len(published) > 0", func() bool {
		for _, tnc := range te.nodeConns {
			if len(tnc.published) == 0 {
				return false
			}
		}
		return true
	})
	//
	// Create SC Client account with some deposit, deploy a contract, wait for a confirming TX.
	scClient := cryptolib.NewKeyPair()
	_, err := te.utxoDB.GetFundsFromFaucet(scClient.Address(), 150_000_000)
	require.NoError(t, err)
	deployReqs := append(te.tcl.MakeTxDeployIncCounterContract(), te.tcl.MakeTxAccountsDeposit(scClient)...)
	deployBaseAnchor, deployBaseAONoID, err := transaction.GetAnchorFromTransaction(te.nodeConns[0].published[0])
	require.NoError(t, err)
	deployBaseAO := isc.NewAliasOutputWithID(deployBaseAONoID, deployBaseAnchor.OutputID)
	for _, tnc := range te.nodeConns {
		tnc.recvAliasOutput(
			isc.NewOutputInfo(deployBaseAO.OutputID(), deployBaseAO.GetAliasOutput(), iotago.TransactionID{}),
		)
		for _, req := range deployReqs {
			onLedgerRequest := req.(isc.OnLedgerRequest)
			tnc.recvRequestCB(
				isc.NewOutputInfo(onLedgerRequest.ID().OutputID(), onLedgerRequest.Output(), iotago.TransactionID{}),
			)
		}
	}
	awaitRequestsProcessed(te, ctxTimeout, deployReqs, "deployReqs")
	awaitPredicate(te, ctxTimeout, "len(tnc.published) > 1", func() bool {
		for _, tnc := range te.nodeConns {
			if len(tnc.published) <= 1 {
				return false
			}
		}
		return true
	})
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
		).WithGasBudget(20000).Sign(scClient)
		te.nodes[0].ReceiveOffLedgerRequest(scRequest, scClient.GetPublicKey())
		incRequests[i] = scRequest
	}
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
					latestL1Commitment, err := state.L1CommitmentFromAliasOutput(latestAONoID)
					require.NoError(t, err)
					st, err := node.GetStateReader().StateByTrieRoot(latestL1Commitment.GetTrieRoot())
					require.NoError(t, err)
					require.GreaterOrEqual(t, incCount, inccounter.NewStateAccess(st).GetCounter())
				*/
				break
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
			lastPublishedAO, err := transaction.GetAliasOutput(lastPublishedTX, te.chainID.AsAddress())
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
	//
	// Check if all requests were processed.
	awaitRequestsProcessed(te, ctxTimeout, incRequests, "incRequests")
}

//nolint:revive
func awaitRequestsProcessed(te *testEnv, ctx context.Context, requests []isc.Request, desc string) {
	reqRefs := isc.RequestRefsFromRequests(requests)
	for i, node := range te.nodes {
		for reqNum, reqRef := range reqRefs {
			te.log.Debugf("Going to AwaitRequestProcessed %v at node=%v, req[%v]=%v...", desc, i, reqNum, reqRef.ID.String())

			select {
			case <-ctx.Done():
				require.FailNowf(te.t, "awaitRequestsProcessed failed: %s", desc)
			case <-node.AwaitRequestProcessed(ctx, reqRef.ID, false):
				require.NoError(te.t, ctx.Err(), "awaitRequestsProcessed failed, context timeout")
			}

			select {
			case <-ctx.Done():
				require.FailNowf(te.t, "awaitRequestsProcessed failed: %s", desc)
			case <-node.AwaitRequestProcessed(ctx, reqRef.ID, true):
				require.NoError(te.t, ctx.Err(), "awaitRequestsProcessed failed, context timeout")
			}

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
	te.tcl = testchain.NewTestChainLedger(t, te.utxoDB, te.governor, te.originator)
	te.originAO, te.chainID = te.tcl.MakeTxChainOrigin(te.cmtAddress)
	//
	// Initialize the nodes.
	te.nodeConns = make([]*testNodeConn, len(te.peerIdentities))
	te.nodes = make([]chain.Chain, len(te.peerIdentities))
	for i := range te.peerIdentities {
		te.nodeConns[i] = newTestNodeConn(t)
		log := te.log.Named(fmt.Sprintf("N#%v", i))
		te.nodes[i], err = chain.New(
			te.ctx,
			te.chainID,
			state.InitChainStore(mapdb.NewMapDB()),
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
