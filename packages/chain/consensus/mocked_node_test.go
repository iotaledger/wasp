// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package consensus

import (
	"math/rand"
	"sync"
	"time"

	//"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/hive.go/kvstore/mapdb"
	"github.com/iotaledger/hive.go/logger"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/chain/committee"
	"github.com/iotaledger/wasp/packages/chain/mempool"
	"github.com/iotaledger/wasp/packages/chain/messages"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/kv/trie"
	// "github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/metrics"
	//"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/testutil/testchain"
	//	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/wal"
	"github.com/stretchr/testify/require"
)

type mockedNode struct {
	NodeID              string
	NodeIndex           uint16
	NodePubKey          *cryptolib.PublicKey
	Env                 *MockedEnv
	NodeConn            *testchain.MockedNodeConn  // GoShimmer mock
	ChainCore           *testchain.MockedChainCore // Chain mock
	Mempool             mempool.Mempool            // Consensus needs
	Consensus           chain.Consensus            // Consensus needs
	LastSolidStateIndex uint32
	SolidStates         map[uint32]state.VirtualStateAccess // State manager mock
	StateOutput         *iscp.AliasOutputWithID             // State manager mock
	Log                 *logger.Logger
	stateSync           coreutil.ChainStateSync
	mutex               sync.Mutex
}

func NewNode(env *MockedEnv, nodeIndex uint16, timers ConsensusTimers) *mockedNode { //nolint:revive
	nodeID := env.NodeIDs[nodeIndex]
	log := env.Log.Named(nodeID)
	ret := &mockedNode{
		NodeID:      nodeID,
		NodePubKey:  env.NodePubKeys[nodeIndex],
		Env:         env,
		NodeConn:    testchain.NewMockedNodeConnection("Node_"+nodeID, env.Ledger, log),
		ChainCore:   testchain.NewMockedChainCore(env.T, env.ChainID, log),
		SolidStates: make(map[uint32]state.VirtualStateAccess),
		Log:         log,
	}
	ret.stateSync = coreutil.NewChainStateSync()
	store := mapdb.NewMapDB()
	ret.ChainCore.OnGlobalStateSync(func() coreutil.ChainStateSync {
		return ret.stateSync
	})
	ret.ChainCore.OnGetStateReader(func() state.OptimisticStateReader {
		return state.NewOptimisticStateReader(store, ret.stateSync)
	})
	/*ret.NodeConn.OnPostTransaction(func(tx *ledgerstate.Transaction) {
		env.mutex.Lock()
		defer env.mutex.Unlock()

		if _, already := env.Ledger.GetTransaction(tx.ID()); !already {
			if err := env.Ledger.AddTransaction(tx); err != nil {
				ret.Log.Error(err)
				return
			}
			stateOutput := transaction.GetAliasOutput(tx, env.ChainID.AsAddress())
			require.NotNil(env.T, stateOutput)

			ret.Log.Infof("stored transaction to the ledger: %s", tx.ID().Base58())
			for _, node := range env.Nodes {
				go func(n *mockedNode) {
					n.mutex.Lock()
					defer n.mutex.Unlock()
					n.StateOutput = stateOutput
					n.checkStateApproval()
				}(node)
			}
		} else {
			ret.Log.Infof("transaction already in the ledger: %s", tx.ID().Base58())
		}
	})*/
	/*ret.NodeConn.OnPullTransactionInclusionState(func(txid ledgerstate.TransactionID) {
		if _, already := env.Ledger.GetTransaction(txid); already {
			go ret.Consensus.EnqueueInclusionsStateMsg(txid, ledgerstate.Confirmed)
		}
	})*/
	mempoolMetrics := metrics.DefaultChainMetrics()
	ret.Mempool = mempool.New(env.StateAddress, ret.ChainCore.GetStateReader(), log, mempoolMetrics)

	//
	// Pass the ACS mock, if it was set in env.MockedACS.
	acs := make([]chain.AsynchronousCommonSubsetRunner, 0, 1)
	if env.MockedACS != nil {
		acs = append(acs, env.MockedACS)
	}
	/*dkShare, err := env.DKShares[nodeIndex].LoadDKShare(env.StateAddress)
	if err != nil {
		panic(err)
	}*/
	cmt, cmtPeerGroup, err := committee.New(
		env.DKShares[nodeIndex],
		env.ChainID,
		env.NetworkProviders[nodeIndex],
		log,
		acs...,
	)
	require.NoError(env.T, err)
	/*cmtPeerGroup.Attach(peering.PeerMessageReceiverConsensus, func(peerMsg *peering.PeerMessageGroupIn) {
		log.Debugf("Consensus received peer message from %v of type %v", peerMsg.SenderPubKey.AsString(), peerMsg.MsgType)
		switch peerMsg.MsgType {
		case peerMsgTypeSignedResult:
			msg, err := messages.NewSignedResultMsg(peerMsg.MsgData)
			if err != nil {
				log.Error(err)
				return
			}
			ret.Consensus.EnqueueSignedResultMsg(&messages.SignedResultMsgIn{
				SignedResultMsg: *msg,
				SenderIndex:     peerMsg.SenderIndex,
			})
		case peerMsgTypeSignedResultAck:
			msg, err := messages.NewSignedResultAckMsg(peerMsg.MsgData)
			if err != nil {
				log.Error(err)
				return
			}
			ret.Consensus.EnqueueSignedResultAckMsg(&messages.SignedResultAckMsgIn{
				SignedResultAckMsg: *msg,
				SenderIndex:        peerMsg.SenderIndex,
			})
		}
	})*/

	originState, err := state.CreateOriginState(store, env.ChainID)
	require.NoError(env.T, err)
	require.Equal(env.T, uint32(0), originState.BlockIndex())
	require.True(env.T, ret.addNewState(originState))

	cons := New(ret.ChainCore, ret.Mempool, cmt, cmtPeerGroup, ret.NodeConn, true, metrics.DefaultChainMetrics(), wal.NewDefault(), timers)
	cons.(*consensus).vmRunner = testchain.NewMockedVMRunner(env.T, log)
	ret.Consensus = cons

	ret.doStateApproved(originState, env.InitStateOutput)

	ret.ChainCore.OnStateCandidate(func(newState state.VirtualStateAccess, approvingOutputID *iotago.UTXOInput) {
		nsCommitment := trie.RootCommitment(newState.TrieAccess())
		ret.Log.Debugf("State manager mock (OnStateCandidate): received state candidate: index %v, commitment %v, approving output ID %v",
			newState.BlockIndex(), nsCommitment, iscp.OID(approvingOutputID))

		if !ret.addNewState(newState) {
			return
		}

		go func() {
			var output *iotago.AliasOutput
			for output = env.Ledger.PullConfirmedOutput(approvingOutputID); output == nil; output = env.Ledger.PullConfirmedOutput(approvingOutputID) {
				ret.Log.Debugf("State manager mock (OnStateCandidate): transaction index %v has not been published yet", newState.BlockIndex())
				time.Sleep(50 * time.Millisecond)
			}

			ret.Log.Debugf("State manager mock (OnStateCandidate): approving output %v received", iscp.OID(approvingOutputID))
			aoCommitment, err := state.L1CommitmentFromAliasOutput(output)
			require.NoError(env.T, err)
			require.True(env.T, trie.EqualCommitments(nsCommitment, aoCommitment.Commitment))

			if output.StateIndex <= ret.StateOutput.GetStateIndex() {
				ret.Log.Debugf("State manager mock (OnStateCandidate): state output index %v received, but it is too old: current state output is %v",
					output.StateIndex, ret.StateOutput.GetStateIndex())
				return
			}

			ret.doStateApproved(newState, iscp.NewAliasOutputWithID(output, approvingOutputID))
		}()
	})
	go ret.pullStateLoop()
	return ret
}

func (n *mockedNode) addNewState(newState state.VirtualStateAccess) bool {
	newStateIndex := newState.BlockIndex()
	nsCommitment := trie.RootCommitment(newState.TrieAccess())
	oldState, ok := n.SolidStates[newStateIndex]
	if ok {
		osCommitment := trie.RootCommitment(oldState.TrieAccess())
		if trie.EqualCommitments(osCommitment, nsCommitment) {
			n.Log.Debugf("State manager mock: duplicating state candidate index %v commitment %s received; ignoring", newStateIndex, nsCommitment)
		} else {
			n.Log.Errorf("State manager mock: contradicting state candidate index %v received: current commitment %s, new commitment %s; ignoring",
				newStateIndex, osCommitment, nsCommitment)
		}
		return false
	}

	if (len(n.SolidStates) > 0) && (n.LastSolidStateIndex >= newStateIndex) {
		n.Log.Debugf("State manager mock: state candidate index %v commitment %s received, but it is not newer than current state %v; ignoring",
			newStateIndex, nsCommitment, n.LastSolidStateIndex)
		return false
	}

	n.SolidStates[newStateIndex] = newState
	n.LastSolidStateIndex = newStateIndex
	n.Log.Debugf("State manager mock: state candidate index %v commitment %s received and accepted", newStateIndex, nsCommitment)
	return true
}

func (n *mockedNode) getState(index uint32) state.VirtualStateAccess {
	result, ok := n.SolidStates[index]
	if !ok {
		n.Log.Debugf("State manager mock: node doesn't contain state index %v", index)
		return nil
	}
	n.Log.Debugf("State manager mock: state index %v found, commitment %s", index, trie.RootCommitment(result.TrieAccess()))
	return result
}

func (n *mockedNode) getStateFromNodes(index uint32) state.VirtualStateAccess {
	n.Log.Debugf("State manager mock: requesting state index %v", index)
	seed := make([]byte, 16)
	rand.Read(seed)
	permutation := util.NewPermutation16(uint16(len(n.Env.Nodes)-1), seed)
	for _, i := range permutation.GetArray() {
		var nodeIndex uint16
		if i < n.NodeIndex {
			nodeIndex = i
		} else {
			nodeIndex = i + 1
		}
		n.Log.Debugf("State manager mock: requesting state %v from node index %v", index, nodeIndex)
		result := n.Env.Nodes[nodeIndex].getState(index)
		if result != nil {
			n.Log.Debugf("State manager mock: requesting state %v from node index %v: found!", index, nodeIndex)
			return result
		}
	}
	n.Log.Errorf("State manager mock: no node has state index %v", index)
	return nil
}

func (n *mockedNode) doStateApproved(newState state.VirtualStateAccess, newStateOutput *iscp.AliasOutputWithID) {
	reqIDsForLastState := make([]iscp.RequestID, 0)
	prefix := kv.Key(util.Uint32To4Bytes(newState.BlockIndex()))
	err := newState.KVStoreReader().Iterate(prefix, func(key kv.Key, value []byte) bool {
		reqid, err := iscp.RequestIDFromBytes(value)
		require.NoError(n.Env.T, err)
		reqIDsForLastState = append(reqIDsForLastState, reqid)
		return true
	})
	require.NoError(n.Env.T, err)
	n.Mempool.RemoveRequests(reqIDsForLastState...)
	n.Log.Debugf("State manager mock: old requests removed: %v", reqIDsForLastState)

	n.StateOutput = newStateOutput
	n.stateSync.SetSolidIndex(n.StateOutput.GetStateIndex())
	n.Consensus.EnqueueStateTransitionMsg(newState, n.StateOutput, time.Now())
	n.Log.Debugf("State manager mock: new state %v approved, commitment %v, state output ID %v",
		n.StateOutput.GetStateIndex(), trie.RootCommitment(newState.TrieAccess()), iscp.OID(n.StateOutput.ID()))
}

func (n *mockedNode) pullStateLoop() {
	for {
		time.Sleep(200 * time.Millisecond)
		stateOutput := n.Env.Ledger.PullState()
		stateIndex := stateOutput.GetStateIndex()
		if stateOutput != nil && (stateIndex > n.StateOutput.GetStateIndex()) {
			n.Log.Debugf("State manager mock (pullStateLoop): new state output received: index %v, id %v",
				stateIndex, iscp.OID(stateOutput.ID()))
			vstate := n.getState(stateIndex)
			if vstate == nil {
				vstate = n.getStateFromNodes(stateIndex)
				require.NotNil(n.Env.T, vstate)
			}
			n.doStateApproved(vstate, stateOutput)
		}
	}
}

/*func (n *mockedNode) checkStateApproval() {
	if n.SolidState == nil || n.StateOutput == nil {
		return
	}
	if n.SolidState.BlockIndex() != n.StateOutput.GetStateIndex() {
		return
	}
	//stateHash, err := hashing.HashValueFromBytes(n.StateOutput.GetStateData())
	stateHash, err := state.L1CommitmentFromAliasOutput(n.StateOutput.GetAliasOutput())
	require.NoError(n.Env.T, err)
	require.True(n.Env.T, trie.EqualCommitments(stateHash.Commitment, trie.RootCommitment(n.SolidState.TrieAccess())))

	reqIDsForLastState := make([]iscp.RequestID, 0)
	prefix := kv.Key(util.Uint32To4Bytes(n.SolidState.BlockIndex()))
	err = n.SolidState.KVStoreReader().Iterate(prefix, func(key kv.Key, value []byte) bool {
		reqid, err := iscp.RequestIDFromBytes(value)
		require.NoError(n.Env.T, err)
		reqIDsForLastState = append(reqIDsForLastState, reqid)
		return true
	})
	require.NoError(n.Env.T, err)
	n.Mempool.RemoveRequests(reqIDsForLastState...)

	n.Log.Infof("STATE APPROVED (%d reqs). Index: %d, State output: %s",
		len(reqIDsForLastState), n.SolidState.BlockIndex(), iscp.OID(n.StateOutput.ID()))

	n.EventStateTransition()
}*/

/*func (n *mockedNode) EventStateTransition() {
	n.Log.Debugf("EventStateTransition")

	n.ChainCore.GlobalStateSync().SetSolidIndex(n.SolidState.BlockIndex())

	n.Consensus.EnqueueStateTransitionMsg(n.SolidState.Copy(), n.StateOutput, time.Now())
}*/

func (n *mockedNode) StartTimer() {
	n.Log.Debugf("started timer..")
	go func() {
		counter := 0
		for {
			n.Consensus.EnqueueTimerMsg(messages.TimerTick(counter))
			counter++
			time.Sleep(50 * time.Millisecond)
		}
	}()
}
