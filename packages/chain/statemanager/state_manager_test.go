package statemanager

import (
	"context"
	"math/rand"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/kvstore/mapdb"
	hivelog "github.com/iotaledger/hive.go/log"

	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/packages/chain/statemanager/gpa"
	"github.com/iotaledger/wasp/packages/chain/statemanager/gpa/sm_gpa_utils"
	"github.com/iotaledger/wasp/packages/chain/statemanager/snapshots"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/metrics"
	"github.com/iotaledger/wasp/packages/origin"
	"github.com/iotaledger/wasp/packages/parameters/parameterstest"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/packages/testutil/testpeers"
	"github.com/iotaledger/wasp/packages/vm/core/migrations/allmigrations"
)

func TestCruelWorld(t *testing.T) {
	t.Skip()
	log := testlogger.NewLogger(t)

	nodeCount := 15
	committeeSize := 5
	blockCount := 50
	minWaitToProduceBlock := 15 * time.Millisecond
	maxMinWaitsToProduceBlock := 10
	getBlockPeriod := 100 * time.Millisecond
	timerTickPeriod := 35 * time.Millisecond
	consensusStateProposalDelay := 50 * time.Millisecond
	consensusStateProposalCount := 50
	consensusDecidedStateDelay := 50 * time.Millisecond
	consensusDecidedStateCount := 50
	mempoolStateRequestDelay := 50 * time.Millisecond
	mempoolStateRequestCount := 50
	snapshotCreateNodeCount := 2
	snapshotCreatePeriod := uint32(7)
	snapshotDelayPeriod := uint32(4)
	snapshotCommitTime := 170 * time.Millisecond

	peeringURLs, peerIdentities := testpeers.SetupKeys(uint16(nodeCount))
	peerPubKeys := make([]*cryptolib.PublicKey, len(peerIdentities))
	for i := range peerPubKeys {
		peerPubKeys[i] = peerIdentities[i].GetPublicKey()
	}
	networkBehaviour := testutil.NewPeeringNetReliable(log)
	network := testutil.NewPeeringNetwork(
		peeringURLs, peerIdentities, 10000,
		networkBehaviour,
		log.NewChildLogger("net"),
	)
	netProviders := network.NetworkProviders()
	bf := sm_gpa_utils.NewBlockFactory(t)
	sms := make([]StateMgr, nodeCount)
	stores := make([]state.Store, nodeCount)
	snapMs := make([]snapshots.SnapshotManager, nodeCount)
	parameters := gpa.NewStateManagerParameters()
	parameters.StateManagerTimerTickPeriod = timerTickPeriod
	parameters.StateManagerGetBlockRetry = getBlockPeriod
	NewMockedSnapshotManagerFun := func(createSnapshots bool, store state.Store, log hivelog.Logger) snapshots.SnapshotManager {
		var createPeriod uint32
		var delayPeriod uint32
		if createSnapshots {
			createPeriod = snapshotCreatePeriod
			delayPeriod = snapshotDelayPeriod
		} else {
			createPeriod = 0
			delayPeriod = 0
		}
		return snapshots.NewMockedSnapshotManager(t, createPeriod, delayPeriod, bf.GetStore(), store, nil, snapshotCommitTime, parameters.TimeProvider, log)
	}
	for i := range sms {
		t.Logf("Creating %v-th state manager for node %s", i, peeringURLs[i])
		var err error
		logNode := log.NewChildLogger(peeringURLs[i])
		stores[i] = state.NewStoreWithUniqueWriteMutex(mapdb.NewMapDB())
		snapMs[i] = NewMockedSnapshotManagerFun(i < snapshotCreateNodeCount, stores[i], logNode)
		origin.InitChain(allmigrations.LatestSchemaVersion, stores[i], bf.GetChainInitParameters(), iotago.ObjectID{}, 0, parameterstest.L1Mock)
		chainMetrics := metrics.NewChainMetricsProvider().GetChainMetrics(isc.EmptyChainID())
		sms[i], err = New(
			context.Background(),
			bf.GetChainID(),
			peerPubKeys[i],
			peerPubKeys,
			netProviders[i],
			sm_gpa_utils.NewMockedTestBlockWAL(),
			snapMs[i],
			stores[i],
			nil,
			chainMetrics.StateManager,
			chainMetrics.Pipe,
			logNode,
			parameters,
		)
		require.NoError(t, err)
	}
	blocks := bf.GetBlocks(blockCount, 1)
	stateDrafts := make([]state.StateDraft, blockCount)
	blockProduced := make([]*atomic.Bool, blockCount)
	for i := range blocks {
		stateDrafts[i] = bf.GetStateDraft(blocks[i])
		blockProduced[i] = &atomic.Bool{}
	}

	// Send blocks to nodes (consensus mock)
	sendBlockResults := make([]<-chan bool, committeeSize)
	for i := range committeeSize {
		ii := i
		sendBlockResults[i] = makeNRequestsVarDelay(blockCount, func() time.Duration {
			return time.Duration(rand.Intn(maxMinWaitsToProduceBlock)+1) * minWaitToProduceBlock
		}, func(bi int) bool {
			t.Logf("Sending block %v to node %s", bi+1, peeringURLs[ii])
			block := <-sms[ii].ConsensusProducedBlock(context.Background(), stateDrafts[bi])
			if block == nil {
				t.Logf("Sending block %v to node %s FAILED", bi+1, peeringURLs[ii])
				return false
			}
			blockProduced[bi].Store(true)
			return true
		})
	}

	// Send ConsensusStateProposal requestss
	consensusStateProposalResult := makeNRequests(consensusStateProposalCount, consensusStateProposalDelay, func(_ int) bool {
		nodeIndex := rand.Intn(nodeCount)
		blockIndex := getRandomProducedBlockAIndex(blockProduced)
		t.Logf("Consensus state proposal request for block %v is sent to node %v", blockIndex+1, peeringURLs[nodeIndex])
		responseCh := sms[nodeIndex].ConsensusStateProposal(context.Background(), bf.GetAnchor(blocks[blockIndex].L1Commitment()))
		<-responseCh
		return true
	})

	// Send ConsensusDecidedState requests
	consensusDecidedStateResult := makeNRequests(consensusDecidedStateCount, consensusDecidedStateDelay, func(_ int) bool {
		nodeIndex := rand.Intn(nodeCount)
		blockIndex := getRandomProducedBlockAIndex(blockProduced)
		t.Logf("Consensus decided state proposal for block %v is sent to node %v", blockIndex+1, peeringURLs[nodeIndex])
		responseCh := sms[nodeIndex].ConsensusDecidedState(context.Background(), bf.GetAnchor(blocks[blockIndex].L1Commitment()))
		state := <-responseCh
		if !blocks[blockIndex].TrieRoot().Equals(state.TrieRoot()) {
			t.Logf("Consensus decided state proposal for block %v to node %v return wrong state: expected trie root %s, received %s",
				blockIndex+1, peeringURLs[nodeIndex], blocks[blockIndex].TrieRoot(), state.TrieRoot())
			return false
		}
		return true
	})

	// Send MempoolStateRequest requests
	mempoolStateRequestResult := makeNRequests(mempoolStateRequestCount, mempoolStateRequestDelay, func(_ int) bool {
		nodeIndex := rand.Intn(nodeCount)
		var newBlockIndex int
		for newBlockIndex == 0 {
			newBlockIndex = getRandomProducedBlockAIndex(blockProduced)
		}
		oldBlockIndex := rand.Intn(newBlockIndex)
		t.Logf("Mempool state request for new block %v and old block %v is sent to node %v", newBlockIndex+1, oldBlockIndex+1, peeringURLs[nodeIndex])
		oldStateOutput := bf.GetAnchor(blocks[oldBlockIndex].L1Commitment())
		newStateOutput := bf.GetAnchor(blocks[newBlockIndex].L1Commitment())
		responseCh := sms[nodeIndex].(*stateManager).ChainFetchStateDiff(context.Background(), oldStateOutput, newStateOutput)
		results := <-responseCh
		expectedNewState, err := bf.GetStore().StateByTrieRoot(blocks[newBlockIndex].TrieRoot())
		if err != nil {
			t.Logf("Mempool state request for new block %v and old block %v to node %v wasn't able to retrieve expected new state: %v",
				newBlockIndex+1, oldBlockIndex+1, peeringURLs[nodeIndex], err)
			return false
		}
		if !expectedNewState.Equals(results.GetNewState()) {
			t.Logf("Mempool state request for new block %v and old block %v to node %v return wrong new state: expected trie root %s, received %s",
				newBlockIndex+1, oldBlockIndex+1, peeringURLs[nodeIndex], blocks[newBlockIndex].TrieRoot(), results.GetNewState().TrieRoot())
			return false
		}
		expectedAddedLength := newBlockIndex - oldBlockIndex
		if len(results.GetAdded()) != expectedAddedLength {
			t.Logf("Mempool state request for new block %v and old block %v to node %v return wrong size added array: expected %v, received %v elements",
				newBlockIndex+1, oldBlockIndex+1, peeringURLs[nodeIndex], expectedAddedLength, len(results.GetAdded()))
			return false
		}
		for i := range results.GetAdded() {
			if !results.GetAdded()[i].Equals(blocks[oldBlockIndex+i+1]) {
				t.Logf("Mempool state request for new block %v and old block %v to node %v return wrong %v-th element of added array: expected commitment %v, received %v",
					newBlockIndex+1, oldBlockIndex+1, peeringURLs[nodeIndex], i, blocks[oldBlockIndex+i+1].L1Commitment(), results.GetAdded()[i].L1Commitment())
				return false
			}
		}
		if len(results.GetRemoved()) > 0 {
			t.Logf("Mempool state request for new block %v and old block %v to node %v return too large removed array: expected it to be empty, received %v elements",
				newBlockIndex+1, oldBlockIndex+1, peeringURLs[nodeIndex], len(results.GetRemoved()))
			return false
		}
		return true
	})

	// Check results
	for _, sendBlockResult := range sendBlockResults {
		requireTrueForSomeTime(t, sendBlockResult, 10*time.Second)
	}
	requireTrueForSomeTime(t, consensusStateProposalResult, 20*time.Second)
	requireTrueForSomeTime(t, consensusDecidedStateResult, 20*time.Second)
	requireTrueForSomeTime(t, mempoolStateRequestResult, 20*time.Second)
}

func getRandomProducedBlockAIndex(blockProduced []*atomic.Bool) int {
	for !blockProduced[0].Load() {
	}
	var maxIndex int
	for maxIndex < len(blockProduced) && blockProduced[maxIndex].Load() {
		maxIndex++
	}
	return rand.Intn(maxIndex)
}

func requireTrueForSomeTime(t *testing.T, ch <-chan bool, timeout time.Duration) {
	select {
	case result := <-ch:
		require.True(t, result)
	case <-time.After(timeout):
		t.Fatal("Timeout")
	}
}

func makeNRequests(count int, delay time.Duration, makeRequestFun func(int) bool) <-chan bool {
	return makeNRequestsVarDelay(count, func() time.Duration { return delay }, makeRequestFun)
}

func makeNRequestsVarDelay(count int, getDelayFun func() time.Duration, makeRequestFun func(int) bool) <-chan bool {
	responseCh := make(chan bool, 1)
	go func() {
		for i := range count {
			if !makeRequestFun(i) {
				responseCh <- false
				return
			}
			time.Sleep(getDelayFun())
		}
		responseCh <- true
	}()
	return responseCh
}
