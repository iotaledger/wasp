package statemanager

import (
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/core/kvstore/mapdb"
	"github.com/iotaledger/wasp/packages/chain/statemanager/smGPA"
	"github.com/iotaledger/wasp/packages/chain/statemanager/smGPA/smGPAUtils"
	"github.com/iotaledger/wasp/packages/cryptolib"
	//"github.com/iotaledger/wasp/packages/gpa"
	//"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/packages/testutil/testpeers"
)

func TestCruelWorld(t *testing.T) {
	log := testlogger.NewLogger(t)
	defer log.Sync()

	nodeCount := 15
	committeeSize := 5
	blockCount := 50
	minWaitToProduceBlock := 5 * time.Millisecond
	maxMinWaitsToProduceBlock := 10
	approveOutputPeriod := 10 * time.Millisecond
	getBlockPeriod := 35 * time.Millisecond
	timerTickPeriod := 20 * time.Millisecond

	peerNetIDs, peerIdentities := testpeers.SetupKeys(uint16(nodeCount))
	peerPubKeys := make([]*cryptolib.PublicKey, len(peerIdentities))
	for i := range peerPubKeys {
		peerPubKeys[i] = peerIdentities[i].GetPublicKey()
	}
	networkBehaviour := testutil.NewPeeringNetReliable(log)
	network := testutil.NewPeeringNetwork(
		peerNetIDs, peerIdentities, 10000,
		networkBehaviour,
		log.Named("net"),
	)
	netProviders := network.NetworkProviders()
	bf := smGPAUtils.NewBlockFactory(t)
	sms := make([]StateMgr, nodeCount)
	stores := make([]state.Store, nodeCount)
	timers := smGPA.NewStateManagerTimers()
	timers.StateManagerTimerTickPeriod = timerTickPeriod
	timers.StateManagerGetBlockRetry = getBlockPeriod
	for i := range sms {
		t.Logf("Creating %v-th state manager for node %s", i, peerNetIDs[i])
		var err error
		stores[i] = state.InitChainStore(mapdb.NewMapDB())
		sms[i], err = New(
			context.Background(),
			bf.GetChainID(),
			peerPubKeys[i],
			peerPubKeys,
			netProviders[i],
			smGPAUtils.NewMockedBlockWAL(),
			stores[i],
			log.Named(peerNetIDs[i]),
			timers,
		)
		require.NoError(t, err)
	}
	blocks, stateOutputs := bf.GetBlocks(blockCount, 1)
	stateDrafts := make([]state.StateDraft, blockCount)
	blockProduced := make([]bool, blockCount)
	blockApproved := make([]bool, blockCount)
	for i := range blocks {
		stateDrafts[i] = bf.GetStateDraft(blocks[i])
		blockProduced[i] = false
		blockApproved[i] = false
	}

	// Send blocks to nodes (consensus mock)
	sendBlockResults := make([]<-chan bool, committeeSize)
	for i := 0; i < committeeSize; i++ {
		ii := i
		resultChan := make(chan bool, 1)
		sendBlockResults[i] = resultChan
		go func() {
			for bi := 0; bi < blockCount; bi++ {
				if !blockApproved[bi] { // If block is already approved, then consensus should not be working on it
					t.Logf("Sending block %v to node %s", bi+1, peerNetIDs[ii])
					err := <-sms[ii].ConsensusProducedBlock(context.Background(), stateDrafts[bi])
					if err != nil {
						t.Logf("Sending block %v to node %s FAILED: %v", bi+1, peerNetIDs[ii], err)
						resultChan <- false
						return
					}
					blockProduced[bi] = true
					time.Sleep(time.Duration(rand.Intn(maxMinWaitsToProduceBlock)+1) * minWaitToProduceBlock)
				}
			}
			resultChan <- true
		}()
	}

	// Approve blocks (consensus mock)
	approveOutputResult := make(chan bool, 1)
	go func() {
		for bi := 0; bi < blockCount; bi++ {
			for !blockProduced[bi] {
			} // Wait for block to be produced in some node at least
			time.Sleep(approveOutputPeriod)
			t.Logf("Approving alias output %v", bi+1)
			for i := 0; i < nodeCount; i++ {
				t.Logf("Approving alias output %v in node %v", bi+1, peerNetIDs[i])
				sms[i].ReceiveConfirmedAliasOutput(stateOutputs[bi])
			}
			blockApproved[bi] = true
		}
		approveOutputResult <- true
	}()

	// Check results
	for _, sendBlockResult := range sendBlockResults {
		requireTrueForSomeTime(t, sendBlockResult, 5*time.Second)
	}
	requireTrueForSomeTime(t, approveOutputResult, 5*time.Second)
	time.Sleep(15 * time.Second)
	expectedIndex := blockCount
	expectedCommitment := blocks[blockCount-1].L1Commitment()
	for i := 0; i < nodeCount; i++ {
		t.Logf("Checking state of node %v", i)
		actualIndex, err := stores[i].LatestBlockIndex()
		require.NoError(t, err)
		require.Equal(t, uint32(expectedIndex), actualIndex)
		actualCommitment, err := stores[i].LatestBlock()
		require.NoError(t, err)
		require.True(t, expectedCommitment.Equals(actualCommitment.L1Commitment()))
	}
}

func requireTrueForSomeTime(t *testing.T, ch <-chan bool, timeout time.Duration) {
	select {
	case result := <-ch:
		require.True(t, result)
	case <-time.After(timeout):
		t.Fatal("Timeout")
	}
}
