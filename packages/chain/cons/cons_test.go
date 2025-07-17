package cons_test

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/bcs-go"
	hivelog "github.com/iotaledger/hive.go/log"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago/iotatest"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/v2/clients/iscmove"
	"github.com/iotaledger/wasp/v2/packages/chain/cons"
	"github.com/iotaledger/wasp/v2/packages/coin"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/gpa"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/isc/isctest"
	"github.com/iotaledger/wasp/v2/packages/kvstore/mapdb"
	"github.com/iotaledger/wasp/v2/packages/origin"
	"github.com/iotaledger/wasp/v2/packages/parameters/parameterstest"
	"github.com/iotaledger/wasp/v2/packages/state"
	"github.com/iotaledger/wasp/v2/packages/state/indexedstore"
	"github.com/iotaledger/wasp/v2/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/v2/packages/testutil/testpeers"
	"github.com/iotaledger/wasp/v2/packages/transaction"
	"github.com/iotaledger/wasp/v2/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/v2/packages/vm/core/coreprocessors"
	"github.com/iotaledger/wasp/v2/packages/vm/core/migrations/allmigrations"
	"github.com/iotaledger/wasp/v2/packages/vm/vmimpl"
)

// Here we run a single consensus instance, step by step with
// regards to the requests to external components (mempool, stateMgr, VM).
func TestConsBasic(t *testing.T) {
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
			func(tt *testing.T) { testConsBasic(tt, test.n, test.f) },
		)
	}
}

func testConsBasic(t *testing.T, n, f int) {
	t.Parallel()
	log := testlogger.NewLogger(t)
	defer log.Shutdown()
	//
	// Node Identities and shared key.
	_, peerIdentities := testpeers.SetupKeys(uint16(n))
	committeeAddress, dkShareProviders := testpeers.SetupDkgTrivial(t, n, f, peerIdentities, nil)
	var chainID isc.ChainID

	initParams := origin.DefaultInitParams(isc.NewAddressAgentID(committeeAddress)).Encode()
	db := mapdb.NewMapDB()
	store := indexedstore.New(state.NewStoreWithUniqueWriteMutex(db))
	_, stateMetadata := origin.InitChain(allmigrations.LatestSchemaVersion, store, initParams, iotago.ObjectID{}, 0, parameterstest.L1Mock)

	stateAnchor0x := isctest.RandomStateAnchor(isctest.RandomAnchorOption{StateMetadata: stateMetadata})
	stateAnchor0 := &stateAnchor0x

	reqs := []isc.Request{
		RandomOnLedgerDepositRequest(stateAnchor0.Owner()),
		RandomOnLedgerDepositRequest(stateAnchor0.Owner()),
		RandomOnLedgerDepositRequest(stateAnchor0.Owner()),
	}
	reqRefs := isc.RequestRefsFromRequests(reqs)
	gasCoin := coin.CoinWithRef{
		Type:  coin.BaseTokenType,
		Value: coin.Value(100),
		Ref:   iotatest.RandomObjectRef(),
	}

	//
	// Construct the chain on L1.
	//utxodb.New(utxodb.DefaultInitParams())
	//
	// Construct the chain on L1: Create the accounts.
	// originator := cryptolib.NewKeyPair()
	// _, err := utxoDB.GetFundsFromFaucet(originator.Address())
	// require.NoError(t, err)
	//
	// Construct the chain on L1: Create the origin TX.
	// outputs, outIDs := utxoDB.GetUnspentOutputs(originator.Address())
	// panic("refactor me: origin.NewChainOriginTransaction")
	// var originTX *iotago.Transaction
	// err = errors.New("refactor me: testConsBasic")
	// require.NoError(t, err)

	// stateAnchor, aliasOutput, err := transaction.GetAnchorFromTransaction(originTX)
	// require.NoError(t, err)
	// require.NotNil(t, stateAnchor)
	// require.NotNil(t, aliasOutput)
	// stateAnchor0 := isc.NewAliasOutputWithID(aliasOutput, stateAnchor.OutputID)
	// err = utxoDB.AddToLedger(originTX)
	// require.NoError(t, err)

	//
	// Deposit some funds
	// outputs, outIDs = utxoDB.GetUnspentOutputs(originator.Address())
	// depositTx, err := transaction.NewRequestTransaction(
	// 	transaction.NewRequestTransactionParams{
	// 		SenderKeyPair:    originator,
	// 		SenderAddress:    originator.Address(),
	// 		UnspentOutputs:   outputs,
	// 		UnspentOutputIDs: outIDs,
	// 		Request: &isc.RequestParameters{
	// 			TargetAddress: chainID.AsAddress(),
	// 			Assets:        isc.NewAssets(100_000_000),
	// 			Metadata: &isc.SendMetadata{
	// 				Message:   accounts.FuncDeposit.Message(),
	// 				GasBudget: 10_000,
	// 			},
	// 		},
	// 	},
	// )
	// require.NoError(t, err)
	// err = utxoDB.AddToLedger(depositTx)
	// require.NoError(t, err)

	//
	// Construct the chain on L1: Find the requests (the first request).
	//
	// Construct the nodes.
	consInstID := []byte{1, 2, 3} // ID of the consensus.
	chainStates := map[gpa.NodeID]state.Store{}
	procConfig := coreprocessors.NewConfig()
	nodeIDs := gpa.NodeIDsFromPublicKeys(testpeers.PublicKeys(peerIdentities))
	nodes := map[gpa.NodeID]gpa.GPA{}
	for i, nid := range nodeIDs {
		nodeLog := log.NewChildLogger(nid.ShortString())
		nodeSK := peerIdentities[i].GetPrivateKey()
		nodeDKShare, err := dkShareProviders[i].LoadDKShare(committeeAddress)
		require.NoError(t, err)
		chainStates[nid] = state.NewStoreWithUniqueWriteMutex(mapdb.NewMapDB())
		_, err = origin.InitChainByStateMetadataBytes(chainStates[nid], stateAnchor0.GetStateMetadata(), 0, parameterstest.L1Mock)
		require.NoError(t, err)
		nodes[nid] = cons.New(
			chainID,
			chainStates[nid],
			nid,
			nodeSK,
			nodeDKShare,
			nil, // rotateTo
			procConfig,
			consInstID,
			gpa.NodeIDFromPublicKey,
			accounts.CommonAccount(),
			nodeLog,
		).AsGPA()
	}
	tc := gpa.NewTestContext(nodes)
	//
	// Provide inputs.
	t.Log("############ Provide Inputs.")
	now := time.Now()
	inputs := map[gpa.NodeID]gpa.Input{}
	for _, nid := range nodeIDs {
		inputs[nid] = cons.NewInputProposal(stateAnchor0)
	}
	tc.WithInputs(inputs).RunAll()
	tc.PrintAllStatusStrings("After Inputs", t.Logf)
	//
	// Provide SM and MP responses on proposals (and time data).
	t.Log("############ Provide TimeData and Proposals from SM/MP.")
	for nid, node := range nodes {
		out := node.Output().(*cons.Output)
		require.Equal(t, cons.Running, out.Status)
		require.NotNil(t, out.NeedMempoolProposal)
		require.NotNil(t, out.NeedStateMgrStateProposal)
		tc.WithInput(nid, cons.NewInputMempoolProposal(reqRefs))
		tc.WithInput(nid, cons.NewInputStateMgrProposalConfirmed())
		tc.WithInput(nid, cons.NewInputTimeData(now))
		tc.WithInput(nid, cons.NewInputL1Info([]*coin.CoinWithRef{&gasCoin}, parameterstest.L1Mock))
	}
	tc.RunAll()
	tc.PrintAllStatusStrings("After MP/SM proposals", t.Logf)
	//
	// Provide Decided data from SM and MP.
	t.Log("############ Provide Decided Data from SM/MP.")
	for nid, node := range nodes {
		out := node.Output().(*cons.Output)
		require.Equal(t, cons.Running, out.Status)
		require.Nil(t, out.NeedMempoolProposal)
		require.Nil(t, out.NeedStateMgrStateProposal)
		require.NotNil(t, out.NeedMempoolRequests)
		require.NotNil(t, out.NeedStateMgrDecidedState)
		l1Commitment, err := transaction.L1CommitmentFromAnchor(out.NeedStateMgrDecidedState)
		require.NoError(t, err)
		chainState, err := chainStates[nid].StateByTrieRoot(l1Commitment.TrieRoot())
		require.NoError(t, err)
		tc.WithInput(nid, cons.NewInputMempoolRequests(reqs))
		tc.WithInput(nid, cons.NewInputStateMgrDecidedVirtualState(chainState))
	}
	tc.RunAll()
	tc.PrintAllStatusStrings("After MP/SM data", t.Logf)
	//
	// Provide Decided data from SM and MP.
	t.Log("############ Run VM, validate the result.")
	for nid, node := range nodes {
		out := node.Output().(*cons.Output)
		require.Equal(t, cons.Running, out.Status)
		require.Nil(t, out.NeedMempoolProposal)
		require.Nil(t, out.NeedStateMgrStateProposal)
		require.Nil(t, out.NeedMempoolRequests)
		require.Nil(t, out.NeedStateMgrDecidedState)
		require.NotNil(t, out.NeedVMResult)
		out.NeedVMResult.Log = hivelog.NewLogger(hivelog.WithLevel(hivelog.LevelError)) // Decrease VM logging.
		vmResult, err := vmimpl.Run(out.NeedVMResult)
		require.NoError(t, err)
		tc.WithInput(nid, cons.NewInputVMResult(vmResult))
	}
	tc.RunAll()
	//
	// Provide Decided data from SM and MP.
	t.Log("############ After VM the VM Run.")
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
		block, _ := chainStates[nid].Commit(out.NeedStateMgrSaveBlock)
		require.NotNil(t, block)
		tc.WithInput(nid, cons.NewInputStateMgrBlockSaved(block))
	}
	tc.RunAll()
	t.Log("############ All should be done now.")
	tc.PrintAllStatusStrings("All done.", t.Logf)
	out0 := nodes[nodeIDs[0]].Output().(*cons.Output)
	for _, node := range nodes {
		out := node.Output().(*cons.Output)
		require.Equal(t, cons.Completed, out.Status)
		require.True(t, out.Terminated)
		require.Nil(t, out.NeedMempoolProposal)
		require.Nil(t, out.NeedStateMgrStateProposal)
		require.Nil(t, out.NeedMempoolRequests)
		require.Nil(t, out.NeedStateMgrDecidedState)
		require.Nil(t, out.NeedVMResult)
		require.NotNil(t, out.Result.Transaction)
		require.NotNil(t, out.Result.Block)
		require.Equal(t, out.Result.Block, out0.Result.Block)
		require.Equal(t, out.Result.Transaction, out0.Result.Transaction)
	}
}

// Run several consensus instances in a chain, receiving inputs from each other.
// This test case has much less of synchronization, because we don't wait for
// all messages to be delivered before responding to the instance requests to
// mempool, stateMgr and VM.
/*
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
	defer log.Shutdown()
	//
	// Node Identities, shared key and ledger.
	_, peerIdentities := testpeers.SetupKeys(uint16(n))
	committeeAddress, dkShareProviders := testpeers.SetupDkgTrivial(t, n, f, peerIdentities, nil)
	nodeIDs := gpa.NodeIDsFromPublicKeys(testpeers.PublicKeys(peerIdentities))
	//utxoDB := utxodb.New(utxodb.DefaultInitParams())
	//
	// Create the accounts.
	scClient := cryptolib.NewKeyPair()
	originator := cryptolib.NewKeyPair()
	_, err := utxoDB.GetFundsFromFaucet(originator.Address())
	require.NoError(t, err)
	//
	// Construct the chain on L1 and prepare requests.
	tcl := testchain.NewTestChainLedger(t, utxoDB, originator)
	anchor := tcl.MakeTxChainOrigin(committeeAddress)
	allRequests := map[int][]isc.Request{}
	if b > 0 {
		_, err = utxoDB.GetFundsFromFaucet(scClient.Address(), 150_000_000)
		require.NoError(t, err)
		allRequests[0] = append(tcl.MakeTxAccountsDeposit(scClient))
	}
	incTotal := 0
	for i := 0; i < b-1; i++ {
		reqs := []isc.Request{}
		reqPerBlock := 3
		for ii := 0; ii < reqPerBlock; ii++ {
			scRequest := isc.NewOffLedgerRequest(
				chainID,
				inccounter.FuncIncCounter.Message(nil),
				uint64(i*reqPerBlock+ii),
				gas.LimitsDefault.MinGasPerRequest,
			).Sign(scClient)
			reqs = append(reqs, scRequest)
			incTotal++
		}
		allRequests[i+1] = reqs
	}
	//
	// Construct the nodes for each instance.
	procConfig := coreprocessors.NewConfigWithTestContracts()
	procCache := processors.MustNew(procConfig)
	doneCHs := map[gpa.NodeID]chan *testInstInput{}
	for _, nid := range nodeIDs {
		doneCHs[nid] = make(chan *testInstInput, 1)
	}
	testNodeStates := map[gpa.NodeID]state.Store{}
	for _, nid := range nodeIDs {
		testNodeStates[nid] = state.NewStoreWithUniqueWriteMutex(mapdb.NewMapDB())
		origin.InitChainByAnchor(testNodeStates[nid], anchor)
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
		t.Log("Going to provide inputs.")
		originL1Commitment, err := transaction.L1CommitmentFromAliasOutput(anchor.GetAliasOutput())
		require.NoError(t, err)
		originState, err := testNodeStates[nid].StateByTrieRoot(originL1Commitment.TrieRoot())
		require.NoError(t, err)
		testChainInsts[0].input(&testInstInput{
			nodeID:          nid,
			baseAliasOutput: anchor,
			baseState:       originState,
		})
	}
	// Wait for all the instances to output.
	t.Log("Waiting for DONE for the last in the chain.")
	doneVals := map[gpa.NodeID]*testInstInput{}
	for nid, doneCH := range doneCHs {
		doneVals[nid] = <-doneCH
	}
	t.Log("Waiting for all instances to terminate.")
	for _, tci := range testChainInsts {
		<-tci.tcTerminated
	}
	t.Log("Done, last block was output and all instances terminated.")
	for _, doneVal := range doneVals {
		require.Equal(t, int64(incTotal), inccounter.NewStateAccess(doneVal.baseState).GetCounter())
	}
}

////////////////////////////////////////////////////////////////////////////////
// testConsInst

type testInstInput struct {
	nodeID          gpa.NodeID
	baseAliasOutput *isc.StateAnchor
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
	chainID isc.ChainID,
	committeeAddress *cryptolib.Address,
	stateIndex int,
	procCache *processors.Cache,
	nodeIDs []gpa.NodeID,
	nodeStates map[gpa.NodeID]state.Store,
	peerIdentities []*cryptolib.KeyPair,
	dkShareRegistryProviders []registry.DKShareRegistryProvider,
	requests []isc.Request,
	doneCB func(nextInput *testInstInput),
	log log.Logger,
) *testConsInst {
	consInstID := []byte(fmt.Sprintf("testConsInst-%v", stateIndex))
	nodes := map[gpa.NodeID]gpa.GPA{}
	for i, nid := range nodeIDs {
		nodeLog := log.NewChildLogger(nid.ShortString())
		nodeSK := peerIdentities[i].GetPrivateKey()
		nodeDKShare, err := dkShareRegistryProviders[i].LoadDKShare(committeeAddress)
		require.NoError(t, err)
		nodes[nid] = cons.New(chainID, nodeStates[nid], nid, nodeSK, nodeDKShare, procCache, consInstID, gpa.NodeIDFromPublicKey, accounts.CommonAccount(), nodeLog).AsGPA()
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
func (tci *testConsInst) tryHandleOutput(nodeID gpa.NodeID) { //nolint:gocyclo
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
		resultState, err := tci.nodeStates[nodeID].StateByTrieRoot(out.Result.Block.TrieRoot())
		require.NoError(tci.t, err)
		tci.doneCB(&testInstInput{
			nodeID:          nodeID,
			baseAliasOutput: out.Result.NextAliasOutput,
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
			tci.t.Error("we have to sync between mempools, should not happen in this test")
		}
		tci.handledNeedMempoolRequests[nodeID] = true
	}
}

func (tci *testConsInst) tryHandledNeedStateMgrDecidedState(nodeID gpa.NodeID, out *cons.Output, inp *testInstInput) {
	if out.NeedStateMgrDecidedState != nil && !tci.handledNeedStateMgrDecidedState[nodeID] {
		if out.NeedStateMgrDecidedState.OutputID() == inp.baseAliasOutput.OutputID() {
			tci.compInputPipe <- map[gpa.NodeID]gpa.Input{nodeID: cons.NewInputStateMgrDecidedVirtualState(inp.baseState)}
		} else {
			tci.t.Error("we have to sync between state managers, should not happen in this test")
		}
		tci.handledNeedStateMgrDecidedState[nodeID] = true
	}
}

func (tci *testConsInst) tryHandledNeedVMResult(nodeID gpa.NodeID, out *cons.Output) {
	if out.NeedVMResult != nil && !tci.handledNeedVMResult[nodeID] {
		out.NeedVMResult.Log = out.NeedVMResult.Log.Desugar().WithOptions(zap.IncreaseLevel(logger.LevelError)).Sugar() // Decrease VM logging.
		vmResult, err := vmimpl.Run(out.NeedVMResult)
		require.NoError(tci.t, err)
		tci.compInputPipe <- map[gpa.NodeID]gpa.Input{nodeID: cons.NewInputVMResult(vmResult)}
		tci.handledNeedVMResult[nodeID] = true
	}
}

func (tci *testConsInst) tryHandledNeedStateMgrSaveBlock(nodeID gpa.NodeID, out *cons.Output) {
	if out.NeedStateMgrSaveBlock != nil && !tci.handledNeedStateMgrSaveBlock[nodeID] {
		block := tci.nodeStates[nodeID].Commit(out.NeedStateMgrSaveBlock)
		tci.compInputPipe <- map[gpa.NodeID]gpa.Input{nodeID: cons.NewInputStateMgrBlockSaved(block)}
		tci.handledNeedStateMgrSaveBlock[nodeID] = true
	}
}

func (tci *testConsInst) tryCloseCompInputPipe() {
	if !tci.compInputClosed.Swap(true) {
		close(tci.compInputPipe)
	}
}

*/

func RandomOnLedgerDepositRequest(senders ...*cryptolib.Address) isc.OnLedgerRequest {
	sender := cryptolib.NewRandomAddress()
	if len(senders) != 0 {
		sender = senders[0]
	}
	ref := iotatest.RandomObjectRef()
	a := iscmove.AssetsBagWithBalances{
		AssetsBag: iscmove.AssetsBag{ID: *iotatest.RandomAddress(), Size: 1},
		Assets:    *iscmove.NewAssets(iotajsonrpc.CoinValue(rand.Int63())),
	}
	req := iscmove.RefWithObject[iscmove.Request]{
		ObjectRef: *ref,
		Object: &iscmove.Request{
			ID:        *ref.ObjectID,
			Sender:    sender,
			AssetsBag: a,
			Message: iscmove.Message{
				Contract: uint32(isc.Hn("accounts")),
				Function: uint32(isc.Hn("deposit")),
			},
			AllowanceBCS: bcs.MustMarshal(iscmove.NewAssets(10000)),
			GasBudget:    100000,
		},
		Owner: sender.AsIotaAddress(),
	}
	onReq, err := isc.OnLedgerFromMoveRequest(&req, sender)
	if err != nil {
		panic(err)
	}
	return onReq
}
