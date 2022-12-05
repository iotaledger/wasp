// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chain_test

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
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/chain/statemanager/smGPA/smGPAUtils"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/metrics/nodeconnmetrics"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/testutil/testchain"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/packages/testutil/testpeers"
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

func testBasic(t *testing.T, n, f int, reliable bool) {
	t.Parallel()
	te := newEnv(t, n, f, reliable)
	defer te.close()
	t.Logf("All started.")
	for _, tnc := range te.nodeConns {
		tnc.waitAttached()
	}
	t.Logf("All attached to node conns.")
	//
	// Create the chain, post Chain Init, wait for the first block.
	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
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
	now := time.Now()
	initRequests := te.tcl.MakeTxChainInit()
	for _, tnc := range te.nodeConns {
		tnc.recvAliasOutput(
			[]iotago.OutputID{te.originAO.OutputID()},
			[]*iotago.AliasOutput{te.originAO.GetAliasOutput()},
		)
		tnc.recvMilestone(now)
		for _, req := range initRequests {
			tnc.recvRequestCB(req.ID().OutputID(), req.(isc.OnLedgerRequest).Output())
		}
	}
	for _, tnc := range te.nodeConns {
		for {
			if len(tnc.published) > 0 {
				break
			}
			time.Sleep(50 * time.Millisecond)
		}
	}
}

////////////////////////////////////////////////////////////////////////////////
// testNodeConn

type testNodeConn struct {
	t               *testing.T
	chainID         *isc.ChainID
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
	chainID *isc.ChainID,
	tx *iotago.Transaction,
	callback chain.TxPostHandler,
) {
	if tnc.chainID == nil {
		tnc.t.Errorf("NodeConn::PublishTX before attach.")
	}
	if !tnc.chainID.Equals(chainID) {
		tnc.t.Errorf("unexpected chain id")
	}
	tnc.published = append(tnc.published, tx)
	callback(tx, true)
}

func (tnc *testNodeConn) AttachChain(
	ctx context.Context,
	chainID *isc.ChainID,
	recvRequestCB chain.RequestOutputHandler,
	recvAliasOutput chain.AliasOutputHandler,
	recvMilestone chain.MilestoneHandler,
) {
	if tnc.chainID != nil {
		tnc.t.Errorf("duplicate attach")
	}
	tnc.chainID = chainID
	tnc.recvRequestCB = recvRequestCB
	tnc.recvAliasOutput = recvAliasOutput
	tnc.recvMilestone = recvMilestone
	tnc.attachWG.Done()
}

func (tnc *testNodeConn) GetMetrics() nodeconnmetrics.NodeConnectionMetrics {
	panic("should be unused in test")
}

func (tnc *testNodeConn) waitAttached() {
	tnc.attachWG.Wait()
}

////////////////////////////////////////////////////////////////////////////////
// testEnv

type testEnv struct {
	ctx              context.Context
	ctxCancel        context.CancelFunc
	log              *logger.Logger
	utxoDB           *utxodb.UtxoDB
	governor         *cryptolib.KeyPair
	originator       *cryptolib.KeyPair
	peerNetIDs       []string
	peerIdentities   []*cryptolib.KeyPair
	peerPubKeys      []*cryptolib.PublicKey
	peeringNetwork   *testutil.PeeringNetwork
	networkProviders []peering.NetworkProvider
	tcl              *testchain.TestChainLedger
	cmtAddress       iotago.Address
	chainID          *isc.ChainID
	originAO         *isc.AliasOutputWithID
	nodeConns        []*testNodeConn
	nodes            []chain.Chain
}

func newEnv(t *testing.T, n, f int, reliable bool) *testEnv {
	te := &testEnv{}
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
	te.peerNetIDs, te.peerIdentities = testpeers.SetupKeys(uint16(n))
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
		te.peerNetIDs, te.peerIdentities, 10000,
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
		te.nodes[i], err = chain.New(
			te.ctx,
			te.chainID,
			state.InitChainStore(mapdb.NewMapDB()),
			te.nodeConns[i],
			te.peerIdentities[i],
			coreprocessors.NewConfigWithCoreContracts().WithNativeContracts(inccounter.Processor),
			dkShareProviders[i],
			testutil.NewMockedCmtLogStore(),
			smGPAUtils.NewMockedBlockWAL(),
			te.networkProviders[i],
			te.log.Named(fmt.Sprintf("N#%v", i)),
		)
		require.NoError(t, err)
	}
	return te
}

func (te *testEnv) close() {
	te.ctxCancel()
	te.peeringNetwork.Close()
	te.log.Sync()
}
