// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cons_test

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/iotaledger/hive.go/core/kvstore/mapdb"
	"github.com/iotaledger/hive.go/core/logger"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/contracts/native/inccounter"
	"github.com/iotaledger/wasp/packages/chain/cons"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/testutil/testchain"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/packages/testutil/testpeers"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/packages/utxodb"
	"github.com/iotaledger/wasp/packages/vm/core/coreprocessors"
	"github.com/iotaledger/wasp/packages/vm/processors"
	"github.com/iotaledger/wasp/packages/vm/runvm"
)

// Here we run a single consensus instance, step by step with
// regards to the requests to external components (mempool, stateMgr, VM).
func TestBasic(t *testing.T) {
	t.Parallel()
	type test struct {
		n int
		f int
	}
	tests := []test{
		{n: 1, f: 0},  // Low N.
		{n: 2, f: 0},  // Low N.
		{n: 3, f: 0},  // Low N.
		{n: 4, f: 1},  // Smallest reasonable config.
		{n: 10, f: 3}, // Typical config?
		{n: 12, f: 3}, // Non-optimal N/F.
	}
	if !testing.Short() {
		tests = append(tests, test{n: 31, f: 10}) // Large cluster.
	}
	for _, test := range tests {
		t.Run(
			fmt.Sprintf("N=%v,F=%v", test.n, test.f),
			func(tt *testing.T) { testBasic(tt, test.n, test.f) },
		)
	}
}

func testBasic(t *testing.T, n, f int) {
	t.Parallel()
	log := testlogger.NewLogger(t)
	defer log.Sync()
	//
	// Node Identities and shared key.
	_, peerIdentities := testpeers.SetupKeys(uint16(n))
	committeeAddress, dkShareProviders := testpeers.SetupDkgTrivial(t, n, f, peerIdentities, nil)
	//
	// Construct the chain on L1.
	utxoDB := utxodb.New(utxodb.DefaultInitParams())
	//
	// Construct the chain on L1: Create the accounts.
	governor := cryptolib.NewKeyPair()
	originator := cryptolib.NewKeyPair()
	_, err := utxoDB.GetFundsFromFaucet(originator.Address())
	require.NoError(t, err)
	//
	// Construct the chain on L1: Create the origin TX.
	outs, outIDs := utxoDB.GetUnspentOutputs(originator.Address())
	originTX, chainID, err := transaction.NewChainOriginTransaction(
		originator,
		committeeAddress,
		governor.Address(),
		1_000_000,
		outs,
		outIDs,
	)
	require.NoError(t, err)
	stateAnchor, aliasOutput, err := transaction.GetAnchorFromTransaction(originTX)
	require.NoError(t, err)
	require.NotNil(t, stateAnchor)
	require.NotNil(t, aliasOutput)
	ao0 := isc.NewAliasOutputWithID(aliasOutput, stateAnchor.OutputID.UTXOInput())
	err = utxoDB.AddToLedger(originTX)
	require.NoError(t, err)
	//
	// Construct the chain on L1: Create the Init Request TX.
	outs, outIDs = utxoDB.GetUnspentOutputs(originator.Address())
	initTX, err := transaction.NewRootInitRequestTransaction(
		originator,
		chainID,
		"my test chain",
		outs,
		outIDs,
	)
	require.NoError(t, err)
	require.NotNil(t, initTX)
	err = utxoDB.AddToLedger(initTX)
	require.NoError(t, err)
	//
	// Construct the chain on L1: Find the requests (the init request).
	initReqs := []isc.Request{}
	initReqRefs := []*isc.RequestRef{}
	outs, _ = utxoDB.GetUnspentOutputs(chainID.AsAddress())
	for outID, out := range outs {
		if out.Type() == iotago.OutputAlias {
			zeroAliasID := iotago.AliasID{}
			outAsAlias := out.(*iotago.AliasOutput)
			if outAsAlias.AliasID == *chainID.AsAliasID() {
				continue // That's our alias output, not the request, skip it here.
			}
			if outAsAlias.AliasID == zeroAliasID {
				implicitAliasID := iotago.AliasIDFromOutputID(outID)
				if implicitAliasID == *chainID.AsAliasID() {
					continue // That's our origin alias output, not the request, skip it here.
				}
			}
		}
		req, err := isc.OnLedgerFromUTXO(out, outID.UTXOInput())
		if err != nil {
			continue
		}
		initReqs = append(initReqs, req)
		initReqRefs = append(initReqRefs, isc.RequestRefFromRequest(req))
	}
	//
	// Construct the nodes.
	consInstID := []byte{1, 2, 3} // ID of the consensus.
	chainStates := map[gpa.NodeID]state.Store{}
	procConfig := coreprocessors.NewConfigWithCoreContracts().WithNativeContracts(inccounter.Processor)
	procCache := processors.MustNew(procConfig)
	nodeIDs := nodeIDsFromPubKeys(testpeers.PublicKeys(peerIdentities))
	nodes := map[gpa.NodeID]gpa.GPA{}
	for i, nid := range nodeIDs {
		nodeLog := log.Named(string(nid))
		nodeSK := peerIdentities[i].GetPrivateKey()
		nodeDKShare, err := dkShareProviders[i].LoadDKShare(committeeAddress)
		chainStates[nid] = state.InitChainStore(mapdb.NewMapDB())
		require.NoError(t, err)
		nodes[nid] = cons.New(*chainID, chainStates[nid], nid, nodeSK, nodeDKShare, procCache, consInstID, nodeIDFromPubKey, nodeLog).AsGPA()
	}
	tc := gpa.NewTestContext(nodes)
	//
	// Provide inputs.
	t.Logf("############ Provide Inputs.")
	now := time.Now()
	inputs := map[gpa.NodeID]gpa.Input{}
	for _, nid := range nodeIDs {
		inputs[nid] = cons.NewInputProposal(ao0)
	}
	tc.WithInputs(inputs).RunAll()
	tc.PrintAllStatusStrings("After Inputs", t.Logf)
	//
	// Provide SM and MP responses on proposals (and time data).
	t.Logf("############ Provide TimeData and Proposals from SM/MP.")
	for nid, node := range nodes {
		out := node.Output().(*cons.Output)
		require.Equal(t, cons.Running, out.Status)
		require.NotNil(t, out.NeedMempoolProposal)
		require.NotNil(t, out.NeedStateMgrStateProposal)
		tc.WithInput(nid, cons.NewInputMempoolProposal(initReqRefs))
		tc.WithInput(nid, cons.NewInputStateMgrProposalConfirmed())
		tc.WithInput(nid, cons.NewInputTimeData(now))
	}
	tc.RunAll()
	tc.PrintAllStatusStrings("After MP/SM proposals", t.Logf)
	//
	// Provide Decided data from SM and MP.
	t.Logf("############ Provide Decided Data from SM/MP.")
	for nid, node := range nodes {
		out := node.Output().(*cons.Output)
		require.Equal(t, cons.Running, out.Status)
		require.Nil(t, out.NeedMempoolProposal)
		require.Nil(t, out.NeedStateMgrStateProposal)
		require.NotNil(t, out.NeedMempoolRequests)
		require.NotNil(t, out.NeedStateMgrDecidedState)
		l1Commitment, err := state.L1CommitmentFromAliasOutput(out.NeedStateMgrDecidedState.GetAliasOutput())
		require.NoError(t, err)
		chainState, err := chainStates[nid].StateByTrieRoot(l1Commitment.GetTrieRoot())
		require.NoError(t, err)
		tc.WithInput(nid, cons.NewInputMempoolRequests(initReqs))
		tc.WithInput(nid, cons.NewInputStateMgrDecidedVirtualState(chainState))
	}
	tc.RunAll()
	tc.PrintAllStatusStrings("After MP/SM data", t.Logf)
	//
	// Provide Decided data from SM and MP.
	t.Logf("############ Run VM, validate the result.")
	for nid, node := range nodes {
		out := node.Output().(*cons.Output)
		require.Equal(t, cons.Running, out.Status)
		require.Nil(t, out.NeedMempoolProposal)
		require.Nil(t, out.NeedStateMgrStateProposal)
		require.Nil(t, out.NeedMempoolRequests)
		require.Nil(t, out.NeedStateMgrDecidedState)
		require.NotNil(t, out.NeedVMResult)
		out.NeedVMResult.Log = out.NeedVMResult.Log.Desugar().WithOptions(zap.IncreaseLevel(logger.LevelError)).Sugar() // Decrease VM logging.
		require.NoError(t, runvm.NewVMRunner().Run(out.NeedVMResult))
		tc.WithInput(nid, cons.NewInputVMResult(out.NeedVMResult))
	}
	tc.RunAll()
	//
	// Provide Decided data from SM and MP.
	t.Logf("############ After VM the VM Run.")
	tc.PrintAllStatusStrings("After VM the VM Run", t.Logf)
	for nid, node := range nodes {
		out := node.Output().(*cons.Output)
		require.Equal(t, cons.Running, out.Status)
		require.Nil(t, out.NeedMempoolProposal)
		require.Nil(t, out.NeedStateMgrStateProposal)
		require.Nil(t, out.NeedMempoolRequests)
		require.Nil(t, out.NeedStateMgrDecidedState)
		require.Nil(t, out.NeedVMResult)
		require.NotNil(t, out.NeedStateMgrSaveBlock)
		tc.WithInput(nid, cons.NewInputStateMgrBlockSaved())
	}
	tc.RunAll()
	t.Logf("############ All should be done now.")
	tc.PrintAllStatusStrings("All done.", t.Logf)
	for nid, node := range nodes {
		out := node.Output().(*cons.Output)
		require.Equal(t, cons.Completed, out.Status)
		require.True(t, out.Terminated)
		require.Nil(t, out.NeedMempoolProposal)
		require.Nil(t, out.NeedStateMgrStateProposal)
		require.Nil(t, out.NeedMempoolRequests)
		require.Nil(t, out.NeedStateMgrDecidedState)
		require.Nil(t, out.NeedVMResult)
		require.NotNil(t, out.ResultTransaction)
		require.NotNil(t, out.ResultNextAliasOutput)
		require.NotNil(t, out.ResultState)
		block := chainStates[nid].Commit(out.ResultState)
		require.NotNil(t, block)
		if nid == nodeIDs[0] { // Just do this once.
			require.NoError(t, utxoDB.AddToLedger(out.ResultTransaction))
		}
	}
}

// Run several consensus instances in a chain, receiving inputs from each other.
// This test case has much less of synchronization, because we don't wait for
// all messages to be delivered before responding to the instance requests to
// mempool, stateMgr and VM.
func TestChained(t *testing.T) {
	t.Parallel()
	type test struct {
		n int
		f int
		b int
	}
	var tests []test
	if testing.Short() {
		tests = []test{
			{n: 1, f: 0, b: 10}, // Low N
			{n: 2, f: 0, b: 10}, // Low N
			{n: 3, f: 0, b: 10}, // Low N
			{n: 4, f: 1, b: 10}, // Smallest possible resilient config.
			{n: 10, f: 3, b: 5}, // Maybe a typical config.
			{n: 12, f: 3, b: 3}, // Check a non-optimal N/F combinations.
		}
	} else {
		tests = []test{ // Block counts chosen to keep test time similar in all cases.
			{n: 1, f: 0, b: 700}, // Low N
			{n: 2, f: 0, b: 500}, // Low N
			{n: 3, f: 0, b: 300}, // Low N
			{n: 4, f: 1, b: 250}, // Smallest possible resilient config.
			{n: 10, f: 3, b: 50}, // Maybe a typical config.
			{n: 12, f: 3, b: 35}, // Check a non-optimal N/F combinations.
			{n: 31, f: 10, b: 2}, // A large cluster.
		}
	}
	for _, test := range tests {
		t.Run(
			fmt.Sprintf("N=%v,F=%v,Blocks=%v", test.n, test.f, test.b),
			func(tt *testing.T) { testChained(tt, test.n, test.f, test.b) },
		)
	}
}

func testChained(t *testing.T, n, f, b int) {
	t.Parallel()
	log := testlogger.NewLogger(t)
	defer log.Sync()
	//
	// Node Identities, shared key and ledger.
	_, peerIdentities := testpeers.SetupKeys(uint16(n))
	committeeAddress, dkShareProviders := testpeers.SetupDkgTrivial(t, n, f, peerIdentities, nil)
	nodeIDs := nodeIDsFromPubKeys(testpeers.PublicKeys(peerIdentities))
	utxoDB := utxodb.New(utxodb.DefaultInitParams())
	//
	// Create the accounts.
	scClient := cryptolib.NewKeyPair()
	governor := cryptolib.NewKeyPair()
	originator := cryptolib.NewKeyPair()
	_, err := utxoDB.GetFundsFromFaucet(governor.Address())
	require.NoError(t, err)
	_, err = utxoDB.GetFundsFromFaucet(originator.Address())
	require.NoError(t, err)
	//
	// Construct the chain on L1 and prepare requests.
	tcl := testchain.NewTestChainLedger(t, utxoDB, governor, originator)
	originAO, chainID := tcl.MakeTxChainOrigin(committeeAddress)
	allRequests := map[int][]isc.Request{}
	allRequests[0] = tcl.MakeTxChainInit()
	if b > 1 {
		_, err = utxoDB.GetFundsFromFaucet(scClient.Address(), 150_000_000)
		require.NoError(t, err)
		allRequests[1] = append(tcl.MakeTxAccountsDeposit(scClient), tcl.MakeTxDeployIncCounterContract()...)
	}
	incTotal := 0
	for i := 2; i < b; i++ {
		reqs := []isc.Request{}
		reqPerBlock := 3
		for ii := 0; ii < reqPerBlock; ii++ {
			scRequest := isc.NewOffLedgerRequest(
				chainID,
				inccounter.Contract.Hname(),
				inccounter.FuncIncCounter.Hname(),
				dict.New(), uint64(i*reqPerBlock+ii),
			).WithGasBudget(20000).Sign(scClient)
			reqs = append(reqs, scRequest)
			incTotal++
		}
		allRequests[i] = reqs
	}
	//
	// Construct the nodes for each instance.
	procConfig := coreprocessors.NewConfigWithCoreContracts().WithNativeContracts(inccounter.Processor)
	procCache := processors.MustNew(procConfig)
	doneCHs := map[gpa.NodeID]chan *testInstInput{}
	for _, nid := range nodeIDs {
		doneCHs[nid] = make(chan *testInstInput, 1)
	}
	testNodeStates := map[gpa.NodeID]state.Store{}
	for _, nid := range nodeIDs {
		testNodeStates[nid] = state.InitChainStore(mapdb.NewMapDB())
	}
	testChainInsts := make([]testConsInst, b)
	for i := range testChainInsts {
		ii := i // Copy.
		doneCB := func(nextInput *testInstInput) {
			if ii == b-1 {
				doneCHs[nextInput.nodeID] <- nextInput
				return
			}
			testChainInsts[ii+1].input(nextInput)
		}
		testChainInsts[i] = *newTestConsInst(
			t, chainID, committeeAddress, i, procCache, nodeIDs,
			testNodeStates, peerIdentities, dkShareProviders,
			allRequests[i], doneCB, log,
		)
	}
	// Start the threads for each instance.
	for i := range testChainInsts {
		go testChainInsts[i].run()
	}
	// Start the process by providing input to the first instance.
	for _, nid := range nodeIDs {
		t.Logf("Going to provide inputs.")
		originL1Commitment, err := state.L1CommitmentFromAliasOutput(originAO.GetAliasOutput())
		require.NoError(t, err)
		originState, err := testNodeStates[nid].StateByTrieRoot(originL1Commitment.GetTrieRoot())
		require.NoError(t, err)
		testChainInsts[0].input(&testInstInput{
			nodeID:          nid,
			baseAliasOutput: originAO,
			baseState:       originState,
		})
	}
	// Wait for all the instances to output.
	t.Logf("Waiting for DONE for the last in the chain.")
	doneVals := map[gpa.NodeID]*testInstInput{}
	for nid, doneCH := range doneCHs {
		doneVals[nid] = <-doneCH
	}
	t.Logf("Waiting for all instances to terminate.")
	for _, tci := range testChainInsts {
		<-tci.tcTerminated
	}
	t.Logf("Done, last block was output and all instances terminated.")
	for _, doneVal := range doneVals {
		require.Equal(t, int64(incTotal), inccounter.NewStateAccess(doneVal.baseState).GetCounter())
	}
}

////////////////////////////////////////////////////////////////////////////////
// testConsInst

type testInstInput struct {
	nodeID          gpa.NodeID
	baseAliasOutput *isc.AliasOutputWithID
	baseState       state.State // State committed with the baseAliasOutput
}

type testConsInst struct {
	t            *testing.T
	nodes        map[gpa.NodeID]gpa.GPA
	nodeStates   map[gpa.NodeID]state.Store
	stateIndex   int
	requests     []isc.Request
	tc           *gpa.TestContext
	tcInputCh    chan map[gpa.NodeID]gpa.Input // These channels are for sending data to TC.
	tcTerminated chan interface{}
	//
	// Inputs received from the previous instance.
	lock            *sync.RWMutex                 // inputs value is checked in the TC thread and written in TCI.
	compInputPipe   chan map[gpa.NodeID]gpa.Input // Local component inputs. This queue is used to send message from TCI/TC to TCI.
	compInputClosed *atomic.Bool                  // Can be closed from TC or TCI.
	inputCh         chan *testInstInput           // These channels are to send data to TCI.
	inputs          map[gpa.NodeID]*testInstInput // Inputs received to this TCI.
	//
	// The latest output of the consensus instances.
	outLatest map[gpa.NodeID]*cons.Output
	//
	// What has been provided to the consensus instances.
	handledNeedMempoolProposal       map[gpa.NodeID]bool
	handledNeedStateMgrStateProposal map[gpa.NodeID]bool
	handledNeedMempoolRequests       map[gpa.NodeID]bool
	handledNeedStateMgrDecidedState  map[gpa.NodeID]bool
	handledNeedVMResult              map[gpa.NodeID]bool
	handledNeedStateMgrSaveBlock     map[gpa.NodeID]bool
	//
	// Result of this instance provided to the next instance.
	done   map[gpa.NodeID]bool
	doneCB func(nextInput *testInstInput)
}

func newTestConsInst(
	t *testing.T,
	chainID *isc.ChainID,
	committeeAddress iotago.Address,
	stateIndex int,
	procCache *processors.Cache,
	nodeIDs []gpa.NodeID,
	nodeStates map[gpa.NodeID]state.Store,
	peerIdentities []*cryptolib.KeyPair,
	dkShareRegistryProviders []registry.DKShareRegistryProvider,
	requests []isc.Request,
	doneCB func(nextInput *testInstInput),
	log *logger.Logger,
) *testConsInst {
	consInstID := []byte(fmt.Sprintf("testConsInst-%v", stateIndex))
	nodes := map[gpa.NodeID]gpa.GPA{}
	for i, nid := range nodeIDs {
		nodeLog := log.Named(string(nid))
		nodeSK := peerIdentities[i].GetPrivateKey()
		nodeDKShare, err := dkShareRegistryProviders[i].LoadDKShare(committeeAddress)
		require.NoError(t, err)
		nodes[nid] = cons.New(*chainID, nodeStates[nid], nid, nodeSK, nodeDKShare, procCache, consInstID, nodeIDFromPubKey, nodeLog).AsGPA()
	}
	tci := &testConsInst{
		t:                                t,
		nodes:                            nodes,
		nodeStates:                       nodeStates,
		stateIndex:                       stateIndex,
		requests:                         requests,
		tcInputCh:                        make(chan map[gpa.NodeID]gpa.Input, len(nodeIDs)),
		tcTerminated:                     make(chan interface{}),
		lock:                             &sync.RWMutex{},
		compInputPipe:                    make(chan map[gpa.NodeID]gpa.Input, len(nodeIDs)*10),
		compInputClosed:                  &atomic.Bool{},
		inputCh:                          make(chan *testInstInput, len(nodeIDs)),
		inputs:                           map[gpa.NodeID]*testInstInput{},
		outLatest:                        map[gpa.NodeID]*cons.Output{},
		handledNeedMempoolProposal:       map[gpa.NodeID]bool{},
		handledNeedStateMgrStateProposal: map[gpa.NodeID]bool{},
		handledNeedMempoolRequests:       map[gpa.NodeID]bool{},
		handledNeedStateMgrDecidedState:  map[gpa.NodeID]bool{},
		handledNeedVMResult:              map[gpa.NodeID]bool{},
		handledNeedStateMgrSaveBlock:     map[gpa.NodeID]bool{},
		done:                             map[gpa.NodeID]bool{},
		doneCB:                           doneCB,
	}
	tci.tc = gpa.NewTestContext(nodes).
		WithOutputHandler(tci.outputHandler).
		WithInputChannel(tci.tcInputCh)
	return tci
}

func (tci *testConsInst) run() {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		tci.tc.RunAll()
		close(tci.tcTerminated)
		cancel()
	}()
	ticks := time.After(10 * time.Millisecond)
	tickSent := 0
	tickClose := false
	var timeForStatus <-chan time.Time
	for {
		select {
		case inp := <-tci.inputCh:
			tci.lock.Lock()
			if _, ok := tci.inputs[inp.nodeID]; ok {
				tci.lock.Unlock()
				panic("duplicate input")
			}
			tci.inputs[inp.nodeID] = inp
			tci.lock.Unlock()
			tci.tcInputCh <- map[gpa.NodeID]gpa.Input{inp.nodeID: cons.NewInputProposal(inp.baseAliasOutput)}
			timeForStatus = time.After(3 * time.Second)
			tci.tryHandleOutput(inp.nodeID)
		case compInp, ok := <-tci.compInputPipe:
			if !ok {
				tickClose = true
				if tickSent > 0 {
					close(tci.tcInputCh)
				}
				tci.compInputPipe = nil
				continue
			}
			tci.tcInputCh <- compInp
		case <-ctx.Done():
			return
		case <-ticks:
			if tickClose && tickSent > 0 {
				continue // tci.tcMessageCh already closed.
			}
			tickSent++
			for nodeID := range tci.nodes {
				tci.tcInputCh <- map[gpa.NodeID]gpa.Input{nodeID: cons.NewInputTimeData(time.Now())}
			}
			if tickClose {
				close(tci.tcInputCh)
				continue
			}
			ticks = time.After(20 * time.Millisecond)
		case <-timeForStatus:
			tci.tc.PrintAllStatusStrings(fmt.Sprintf("TCI[%v] timeForStatus", tci.stateIndex), tci.t.Logf)
			timeForStatus = time.After(3 * time.Second)
		}
	}
}

func (tci *testConsInst) input(input *testInstInput) {
	tci.inputCh <- input
}

func (tci *testConsInst) outputHandler(nodeID gpa.NodeID, out gpa.Output) {
	tci.lock.Lock()
	tci.outLatest[nodeID] = out.(*cons.Output)
	tci.lock.Unlock()
	tci.tryHandleOutput(nodeID)
}

// Here we respond to the node requests to other components (provided via the output).
// This can be executed in the TCI (on input) and TC (on output) threads.
func (tci *testConsInst) tryHandleOutput(nodeID gpa.NodeID) { //nolint: gocyclo
	tci.lock.Lock()
	defer tci.lock.Unlock()
	out, ok := tci.outLatest[nodeID]
	if !ok {
		return
	}
	inp, ok := tci.inputs[nodeID]
	if !ok {
		return
	}
	switch out.Status {
	case cons.Completed:
		if tci.done[nodeID] {
			return
		}
		resultBlock := tci.nodeStates[nodeID].Commit(out.ResultState)
		resultState, err := tci.nodeStates[nodeID].StateByTrieRoot(resultBlock.TrieRoot())
		require.NoError(tci.t, err)
		tci.doneCB(&testInstInput{
			nodeID:          nodeID,
			baseAliasOutput: out.ResultNextAliasOutput,
			baseState:       resultState,
		})
		tci.done[nodeID] = true
		return
	case cons.Skipped:
		if tci.done[nodeID] {
			return
		}
		tci.doneCB(inp)
		tci.done[nodeID] = true
		return
	}
	tci.tryHandledNeedMempoolProposal(nodeID, out, inp)
	tci.tryHandledNeedStateMgrStateProposal(nodeID, out, inp)
	tci.tryHandledNeedMempoolRequests(nodeID, out)
	tci.tryHandledNeedStateMgrDecidedState(nodeID, out, inp)
	tci.tryHandledNeedVMResult(nodeID, out)
	tci.tryHandledNeedStateMgrSaveBlock(nodeID, out)
	allClosed := true
	if len(tci.inputs) < len(tci.nodes) {
		allClosed = false
	}
	for nid := range tci.nodes {
		if tci.handledNeedMempoolProposal[nid] &&
			tci.handledNeedStateMgrStateProposal[nid] &&
			tci.handledNeedMempoolRequests[nid] &&
			tci.handledNeedStateMgrDecidedState[nid] &&
			tci.handledNeedVMResult[nid] &&
			tci.handledNeedStateMgrSaveBlock[nid] {
			continue
		}
		allClosed = false
	}
	if allClosed {
		tci.tryCloseCompInputPipe()
	}
}

func (tci *testConsInst) tryHandledNeedMempoolProposal(nodeID gpa.NodeID, out *cons.Output, inp *testInstInput) {
	if out.NeedMempoolProposal != nil && !tci.handledNeedMempoolProposal[nodeID] {
		require.Equal(tci.t, out.NeedMempoolProposal, inp.baseAliasOutput)
		reqRefs := []*isc.RequestRef{}
		for _, r := range tci.requests {
			reqRefs = append(reqRefs, isc.RequestRefFromRequest(r))
		}
		tci.compInputPipe <- map[gpa.NodeID]gpa.Input{nodeID: cons.NewInputMempoolProposal(reqRefs)}
		tci.handledNeedMempoolProposal[nodeID] = true
	}
}

func (tci *testConsInst) tryHandledNeedStateMgrStateProposal(nodeID gpa.NodeID, out *cons.Output, inp *testInstInput) {
	if out.NeedStateMgrStateProposal != nil && !tci.handledNeedStateMgrStateProposal[nodeID] {
		require.Equal(tci.t, out.NeedStateMgrStateProposal, inp.baseAliasOutput)
		tci.compInputPipe <- map[gpa.NodeID]gpa.Input{nodeID: cons.NewInputStateMgrProposalConfirmed()}
		tci.handledNeedStateMgrStateProposal[nodeID] = true
	}
}

func (tci *testConsInst) tryHandledNeedMempoolRequests(nodeID gpa.NodeID, out *cons.Output) {
	if out.NeedMempoolRequests != nil && !tci.handledNeedMempoolRequests[nodeID] {
		requests := []isc.Request{}
		for _, reqRef := range out.NeedMempoolRequests {
			for _, req := range tci.requests {
				if reqRef.IsFor(req) {
					requests = append(requests, req)
					break
				}
			}
		}
		if len(requests) == len(out.NeedMempoolRequests) {
			tci.compInputPipe <- map[gpa.NodeID]gpa.Input{nodeID: cons.NewInputMempoolRequests(requests)}
		} else {
			tci.t.Errorf("We have to sync between mempools, should not happen in this test.")
		}
		tci.handledNeedMempoolRequests[nodeID] = true
	}
}

func (tci *testConsInst) tryHandledNeedStateMgrDecidedState(nodeID gpa.NodeID, out *cons.Output, inp *testInstInput) {
	if out.NeedStateMgrDecidedState != nil && !tci.handledNeedStateMgrDecidedState[nodeID] {
		if out.NeedStateMgrDecidedState.OutputID() == inp.baseAliasOutput.OutputID() {
			tci.compInputPipe <- map[gpa.NodeID]gpa.Input{nodeID: cons.NewInputStateMgrDecidedVirtualState(inp.baseState)}
		} else {
			tci.t.Errorf("We have to sync between state managers, should not happen in this test.")
		}
		tci.handledNeedStateMgrDecidedState[nodeID] = true
	}
}

func (tci *testConsInst) tryHandledNeedVMResult(nodeID gpa.NodeID, out *cons.Output) {
	if out.NeedVMResult != nil && !tci.handledNeedVMResult[nodeID] {
		out.NeedVMResult.Log = out.NeedVMResult.Log.Desugar().WithOptions(zap.IncreaseLevel(logger.LevelError)).Sugar() // Decrease VM logging.
		require.NoError(tci.t, runvm.NewVMRunner().Run(out.NeedVMResult))
		tci.compInputPipe <- map[gpa.NodeID]gpa.Input{nodeID: cons.NewInputVMResult(out.NeedVMResult)}
		tci.handledNeedVMResult[nodeID] = true
	}
}

func (tci *testConsInst) tryHandledNeedStateMgrSaveBlock(nodeID gpa.NodeID, out *cons.Output) {
	if out.NeedStateMgrSaveBlock != nil && !tci.handledNeedStateMgrSaveBlock[nodeID] {
		tci.compInputPipe <- map[gpa.NodeID]gpa.Input{nodeID: cons.NewInputStateMgrBlockSaved()}
		tci.handledNeedStateMgrSaveBlock[nodeID] = true
	}
}

func (tci *testConsInst) tryCloseCompInputPipe() {
	if !tci.compInputClosed.Swap(true) {
		close(tci.compInputPipe)
	}
}

////////////////////////////////////////////////////////////////////////////////
// Helper functions.

func nodeIDsFromPubKeys(pubKeys []*cryptolib.PublicKey) []gpa.NodeID {
	ret := make([]gpa.NodeID, len(pubKeys))
	for i := range pubKeys {
		ret[i] = nodeIDFromPubKey(pubKeys[i])
	}
	return ret
}

func nodeIDFromPubKey(pubKey *cryptolib.PublicKey) gpa.NodeID {
	return gpa.NodeID("N#" + pubKey.String()[:6])
}
