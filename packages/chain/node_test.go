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

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/kvstore/mapdb"
	"github.com/iotaledger/hive.go/logger"

	"github.com/iotaledger/wasp/clients"
	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/clients/iota-go/iotasigner"
	iotatest2 "github.com/iotaledger/wasp/clients/iota-go/iotatest"
	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/clients/iscmove/iscmoveclient"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/chain/cons/cons_gr"
	"github.com/iotaledger/wasp/packages/chain/mempool"
	"github.com/iotaledger/wasp/packages/chain/statemanager/sm_gpa"
	"github.com/iotaledger/wasp/packages/chain/statemanager/sm_gpa/sm_gpa_utils"
	"github.com/iotaledger/wasp/packages/chain/statemanager/sm_snapshots"
	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/metrics"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/shutdown"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/state/indexedstore"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/testutil/l1starter"
	"github.com/iotaledger/wasp/packages/testutil/testchain"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/packages/testutil/testpeers"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/packages/util/bcs"
	"github.com/iotaledger/wasp/packages/util/patient_log"

	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/coreprocessors"
	"github.com/iotaledger/wasp/packages/vm/core/testcore/contracts/inccounter"
)

type tc struct {
	n        int
	f        int
	reliable bool
	timeout  time.Duration
}

func TestMain(m *testing.M) {
	l1starter.TestMain(m)
}

func TestNodeBasic(t *testing.T) {
	t.Parallel()
	tests := []tc{
		{n: 1, f: 0, reliable: true, timeout: 1000 * time.Second},   // Low N
		{n: 2, f: 0, reliable: true, timeout: 2000 * time.Second},   // Low N
		{n: 3, f: 0, reliable: true, timeout: 3000 * time.Second},   // Low N
		{n: 4, f: 0, reliable: true, timeout: 4000 * time.Second},   // Minimal robust config.
		{n: 4, f: 1, reliable: true, timeout: 5000 * time.Second},   // Minimal robust config.
		{n: 10, f: 3, reliable: true, timeout: 13000 * time.Second}, // Typical config.
	}
	if !testing.Short() {
		tests = append(tests,
			// TODO these "unreliable" tests are crazy, they either succeed in 10~20s or run forever...
			tc{n: 4, f: 1, reliable: false, timeout: 50 * time.Minute},   // Minimal robust config.
			tc{n: 10, f: 3, reliable: false, timeout: 150 * time.Minute}, // Typical config.
			tc{n: 31, f: 10, reliable: true, timeout: 250 * time.Minute}, // Large cluster, reliable - to make test faster.
		)
	}
	for _, tst := range tests {
		t.Run(
			fmt.Sprintf("N=%v,F=%v,Reliable=%v", tst.n, tst.f, tst.reliable),
			func(tt *testing.T) { testNodeBasic(tt, tst.n, tst.f, tst.reliable, tst.timeout, l1starter.Instance()) },
		)
	}
}

func testNodeBasic(t *testing.T, n, f int, reliable bool, timeout time.Duration, node l1starter.IotaNodeEndpoint) {
	t.Parallel()
	te := newEnv(t, n, f, reliable, node)
	defer te.close()

	ctxTimeout, ctxTimeoutCancel := context.WithTimeout(te.ctx, timeout)
	defer ctxTimeoutCancel()

	te.log.Debugf("All started.")
	for _, tnc := range te.nodeConns {
		tnc.waitAttached()
	}
	te.log.Debugf("All attached to node conns.")

	// Create SC L1Client account with some deposit
	scClient := cryptolib.NewKeyPair()
	err := te.l1Client.RequestFunds(context.Background(), *scClient.Address())
	require.NoError(t, err)

	//
	// The first AO should be reported by L1/NodeConn to the nodes.
	for _, tnc := range te.nodeConns {
		tnc.recvAnchor(te.anchor)
	}

	// Invoke off-ledger requests on the contract, wait for the counter to reach the expected value.
	// We only send the requests to the first node. Mempool has to disseminate them.
	incCount := 10
	incRequests := make([]iscmove.RefWithObject[iscmove.Request], incCount)

	for i := 0; i < incCount; i++ {
		assets := iscmove.NewAssets(10000000)
		allowance := iscmove.NewAssets(assets.BaseToken() / 10)
		req, err := te.l2Client.CreateAndSendRequestWithAssets(ctxTimeout, &iscmoveclient.CreateAndSendRequestWithAssetsRequest{
			Signer:        scClient,
			PackageID:     te.iscPackageID,
			AnchorAddress: te.anchor.GetObjectID(),
			Assets:        assets,
			Message: &iscmove.Message{
				Contract: uint32(isc.Hn("accounts")),
				Function: uint32(isc.Hn("deposit")),
			},
			Allowance:        allowance,
			OnchainGasBudget: 100000,
			GasPrice:         iotaclient.DefaultGasPrice,
			GasBudget:        iotaclient.DefaultGasBudget,
		})
		require.NoError(t, err)
		reqRef, err := req.GetCreatedObjectInfo(iscmove.RequestModuleName, iscmove.RequestObjectName)
		require.NoError(t, err)
		reqWithObj, err := te.l2Client.GetRequestFromObjectID(context.Background(), reqRef.ObjectID)
		require.NoError(t, err)
		//		onLedgerReq, err := isc.OnLedgerFromRequest(reqWithObj, cryptolib.NewAddressFromIota(te.anchor.GetObjectID()))
		//		require.NoError(t, err)

		incRequests[i] = *reqWithObj
	}

	collectedRequests := make([]isc.Request, 0)
	for _, tnc := range te.nodeConns {
		for _, req := range incRequests {
			onLedgerRequest, err := isc.OnLedgerFromRequest(&req, tnc.chainID.AsAddress())
			require.NoError(t, err)
			collectedRequests = append(collectedRequests, onLedgerRequest)
			tnc.recvRequest(
				onLedgerRequest,
			)
		}
	}

	awaitRequestsProcessed(ctxTimeout, te, collectedRequests, "incRequests")

	// assert state
	for i, node := range te.nodes {
		for {
			latestState, err := node.LatestState(chain.ActiveOrCommittedState)
			require.NoError(t, err)
			cnt := inccounter.NewStateAccess(latestState).GetCounter()
			patient_log.LogTimeLimited("node_test.counter", 1*time.Second, func() {
				te.log.Debugf("Counter[node=%v]=%v", i, cnt)
			})
			if cnt >= int64(incCount) {
				// TODO: Double-check with the published TX.
				/*
					latestTX := te.nodeConns[i].published[len(te.nodeConns[i].published)-1]
					_, latestAONoID, err := transaction.GetAnchorFromTransaction(latestTX)
					require.NoError(t, err)
					latestL1Commitment, err := transaction.L1CommitmentFromAliasOutput(latestAONoID)
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
					inccounter.FuncIncCounter.Message(nil),
					uint64(ii),
					20000,
				).Sign(scClient)
				te.nodes[0].ReceiveOffLedgerRequest(scRequest, scClient.GetPublicKey())
			}
			time.Sleep(100 * time.Millisecond)
		}
		// Check if LastAliasOutput() works as expected.
		awaitPredicate(te, ctxTimeout, "LatestAliasOutput", func() bool {
			confirmedAO, err := node.LatestAnchor(chain.ConfirmedState)
			require.NoError(t, err)
			activeAO, err := node.LatestAnchor(chain.ActiveState)
			require.NoError(t, err)
			lastPublishedTX := te.nodeConns[i].published[len(te.nodeConns[i].published)-1]
			lastPublishedAO := isc.NewStateAnchor(lastPublishedTX, te.iscPackageID)
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
				select {
				case rec := <-node.AwaitRequestProcessed(ctx, reqRef.ID, confirmed):
					if rec.Error != nil {
						te.t.Fatalf("request processed with an error, %s", rec.Error.Error())
					}
				case <-ctx.Done():
					if ctx.Err() != nil {
						te.t.Fatalf("awaitRequestsProcessed (%t) failed: %s, context timeout", confirmed, desc)
					}
				}
			}

			await(false)
			//await(true)
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
	t           *testing.T
	chainID     isc.ChainID
	published   []*iscmove.AnchorWithRef
	recvRequest chain.RequestHandler
	recvAnchor  chain.AnchorHandler
	attachWG    *sync.WaitGroup

	l1Client     clients.L1Client
	l2Client     clients.L2Client
	iscPackageID iotago.PackageID
}

func (tnc *testNodeConn) GetL1Params() *parameters.L1Params {
	// TODO implement me
	panic("implement me")
}

func (tnc *testNodeConn) GetGasCoinRef(ctx context.Context, chainID isc.ChainID) (*coin.CoinWithRef, error) {
	panic("implement me")
}

var _ chain.NodeConnection = &testNodeConn{}

func newTestNodeConn(t *testing.T, l1Client clients.L1Client, iscPackageID iotago.PackageID) *testNodeConn {
	tnc := &testNodeConn{
		t:            t,
		published:    []*iscmove.AnchorWithRef{},
		attachWG:     &sync.WaitGroup{},
		l1Client:     l1Client,
		l2Client:     l1Client.L2(),
		iscPackageID: iscPackageID,
	}
	tnc.attachWG.Add(1)
	return tnc
}

func (tnc *testNodeConn) PublishTX(
	ctx context.Context,
	chainID isc.ChainID,
	tx iotasigner.SignedTransaction,
	callback chain.TxPostHandler,
) error {
	if tnc.chainID.Empty() {
		tnc.t.Error("NodeConn::PublishTX before attach.")
	}
	if !tnc.chainID.Equals(chainID) {
		tnc.t.Error("unexpected chain id")
	}

	txBytes, err := bcs.Marshal(tx.Data)
	if err != nil {
		return err
	}

	res, err := tnc.l1Client.ExecuteTransactionBlock(ctx, iotaclient.ExecuteTransactionBlockRequest{
		TxDataBytes: txBytes,
		Signatures:  tx.Signatures,
		Options: &iotajsonrpc.IotaTransactionBlockResponseOptions{
			ShowInput:          true,
			ShowRawInput:       true,
			ShowEffects:        true,
			ShowEvents:         true,
			ShowObjectChanges:  true,
			ShowBalanceChanges: true,
			ShowRawEffects:     true,
		},
		RequestType: iotajsonrpc.TxnRequestTypeWaitForLocalExecution,
	})
	if err != nil {
		tnc.t.Logf("ExecuteTransactionBlock, err=%v", err)
		return err
	}

	time.Sleep(5 * time.Second)

	res, err = tnc.l1Client.GetTransactionBlock(ctx, iotaclient.GetTransactionBlockRequest{
		Digest: &res.Digest,

		Options: &iotajsonrpc.IotaTransactionBlockResponseOptions{
			ShowInput:          true,
			ShowRawInput:       true,
			ShowEffects:        true,
			ShowEvents:         true,
			ShowObjectChanges:  true,
			ShowBalanceChanges: true,
			ShowRawEffects:     true,
		},
	})
	if err != nil {
		tnc.t.Logf("GetTransactionBlock, err=%v", err)
		return err
	}

	anchorInfo, err := res.GetMutatedObjectInfo(iscmove.AnchorModuleName, iscmove.AnchorObjectName)
	if err != nil {
		return err
	}

	anchor, err := tnc.l2Client.GetAnchorFromObjectID(ctx, anchorInfo.ObjectID)
	if err != nil {
		return err
	}
	tnc.t.Logf("GOT NEW MUTATED ANCHOR: %v | %v\n", anchorInfo, anchor)

	tnc.published = append(tnc.published, anchor)
	stateAnchor := isc.NewStateAnchor(anchor, tnc.iscPackageID)

	callback(tx, &stateAnchor, nil)

	tnc.recvAnchor(&stateAnchor)
	return nil
}

func (tnc *testNodeConn) AttachChain(
	ctx context.Context,
	chainID isc.ChainID,
	recvRequest chain.RequestHandler,
	recvAnchor chain.AnchorHandler,
	onChainConnect func(),
	onChainDisconnect func(),
) {
	if !tnc.chainID.Empty() {
		tnc.t.Error("duplicate attach")
	}

	tnc.chainID = chainID
	tnc.recvAnchor = recvAnchor
	tnc.recvRequest = recvRequest
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

func (tnc *testNodeConn) ConsensusGasPriceProposal(
	ctx context.Context,
	anchor *isc.StateAnchor,
) <-chan cons_gr.NodeConnGasInfo {
	t := make(chan cons_gr.NodeConnGasInfo)

	// TODO: Refactor this separate goroutine and place it somewhere connection related instead
	go func() {
		stateMetadata, err := transaction.StateMetadataFromBytes(anchor.GetStateMetadata())
		if err != nil {
			panic(err)
		}

		gasCoin, err := tnc.l1Client.GetObject(context.Background(), iotaclient.GetObjectRequest{
			ObjectID: stateMetadata.GasCoinObjectID,
			Options:  &iotajsonrpc.IotaObjectDataOptions{ShowBcs: true},
		})
		if err != nil {
			panic(err)
		}

		var moveBalance iscmoveclient.MoveCoin
		err = iotaclient.UnmarshalBCS(gasCoin.Data.Bcs.Data.MoveObject.BcsBytes, &moveBalance)
		if err != nil {
			panic("failed to decode gas coin object: " + err.Error())
		}

		ref := gasCoin.Data.Ref()
		var coinInfo cons_gr.NodeConnGasInfo = &testNodeConnGasInfo{
			gasCoins: []*coin.CoinWithRef{{
				Type:  coin.BaseTokenType,
				Value: coin.Value(moveBalance.Balance),
				Ref:   &ref,
			}},
			gasPrice: parameters.L1Default.Protocol.ReferenceGasPrice.Uint64(),
		}

		t <- coinInfo
	}()

	return t
}

type testNodeConnGasInfo struct {
	gasCoins []*coin.CoinWithRef
	gasPrice uint64
}

func (tgi *testNodeConnGasInfo) GetGasCoins() []*coin.CoinWithRef { return tgi.gasCoins }
func (tgi *testNodeConnGasInfo) GetGasPrice() uint64              { return tgi.gasPrice }

// RefreshOnLedgerRequests implements chain.NodeConnection.
func (tnc *testNodeConn) RefreshOnLedgerRequests(ctx context.Context, chainID isc.ChainID) {
	// noop
}

////////////////////////////////////////////////////////////////////////////////
// testEnv

type testEnv struct {
	t                *testing.T
	ctx              context.Context
	ctxCancel        context.CancelFunc
	log              *logger.Logger
	peeringURLs      []string
	peerIdentities   []*cryptolib.KeyPair
	peerPubKeys      []*cryptolib.PublicKey
	peeringNetwork   *testutil.PeeringNetwork
	networkProviders []peering.NetworkProvider
	tcl              *testchain.TestChainLedger
	cmtAddress       *cryptolib.Address
	cmtSigner        cryptolib.Signer
	chainID          isc.ChainID
	anchor           *isc.StateAnchor
	nodeConns        []*testNodeConn
	nodes            []chain.Chain

	l1Client     clients.L1Client
	l2Client     clients.L2Client
	iscPackageID iotago.PackageID
}

func newEnv(t *testing.T, n, f int, reliable bool, node l1starter.IotaNodeEndpoint) *testEnv {
	te := &testEnv{t: t}
	te.ctx, te.ctxCancel = context.WithCancel(context.Background())
	te.log = testlogger.NewLogger(t).Named(fmt.Sprintf("%04d", rand.Intn(10000))) // For test instance ID.

	te.iscPackageID = node.ISCPackageID()
	te.l1Client = node.L1Client()
	te.l2Client = te.l1Client.L2()

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
	te.cmtSigner = testpeers.NewTestDSSSigner(te.cmtAddress, dkShareProviders, gpa.MakeTestNodeIDs(n), te.peerIdentities, te.log)

	require.NoError(t, node.L1Client().RequestFunds(context.Background(), *te.cmtSigner.Address()))
	iotatest2.EnsureCoinSplitWithBalance(t, cryptolib.SignerToIotaSigner(te.cmtSigner), node.L1Client(), isc.GasCoinMaxValue*10)

	iscPackageID := node.ISCPackageID()
	te.tcl = testchain.NewTestChainLedger(t, te.cmtSigner, &iscPackageID, te.l1Client)
	var originDeposit coin.Value
	te.anchor, originDeposit = te.tcl.MakeTxChainOrigin()
	te.chainID = te.anchor.ChainID()
	//
	// Initialize the nodes.
	te.nodeConns = make([]*testNodeConn, len(te.peerIdentities))
	te.nodes = make([]chain.Chain, len(te.peerIdentities))

	var err error

	for i := range te.peerIdentities {
		te.nodeConns[i] = newTestNodeConn(t, te.l1Client, te.iscPackageID)
		log := te.log.Named(fmt.Sprintf("N#%v", i))
		chainMetrics := metrics.NewChainMetricsProvider().GetChainMetrics(isc.EmptyChainID())
		te.nodes[i], err = chain.New(
			te.ctx,
			log,
			te.chainID,
			indexedstore.NewFake(state.NewStoreWithUniqueWriteMutex(mapdb.NewMapDB())),
			te.nodeConns[i],
			te.peerIdentities[i],
			coreprocessors.NewConfigWithTestContracts(),
			dkShareProviders[i],
			testutil.NewConsensusStateRegistry(),
			false,
			sm_gpa_utils.NewMockedTestBlockWAL(),
			sm_snapshots.NewEmptySnapshotManager(),
			chain.NewEmptyChainListener(),
			[]*cryptolib.PublicKey{}, // Access nodes.
			te.networkProviders[i],
			chainMetrics,
			shutdown.NewCoordinator("test", log),
			nil,
			nil,
			true,
			-1,
			1,
			10*time.Millisecond,
			10*time.Second,
			accounts.CommonAccount(),
			sm_gpa.NewStateManagerParameters(),
			mempool.Settings{
				TTL:                   24 * time.Hour,
				MaxOffledgerInPool:    1000,
				MaxOnledgerInPool:     1000,
				MaxTimedInPool:        1000,
				MaxOnledgerToPropose:  1000,
				MaxOffledgerToPropose: 1000,
			},
			1*time.Second,
			originDeposit,
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
