package consensus

import (
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/core/kvstore/mapdb"
	"github.com/iotaledger/hive.go/core/logger"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/nodeclient"
	"github.com/iotaledger/trie.go/trie"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/chain/committee"
	"github.com/iotaledger/wasp/packages/chain/consensus/journal"
	dss_node "github.com/iotaledger/wasp/packages/chain/dss/node"
	"github.com/iotaledger/wasp/packages/chain/mempool"
	"github.com/iotaledger/wasp/packages/chain/messages"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/metrics"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/testutil/testchain"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/wal"
)

type mockedNode struct {
	NodeID              string
	NodeIndex           uint16
	NodeKeyPair         *cryptolib.KeyPair
	Env                 *MockedEnv
	NodeConn            *testchain.MockedNodeConn           // L1 mock
	ChainCore           *testchain.MockedChainCore          // Chain mock
	Mempool             mempool.Mempool                     // Chain mock
	Consensus           chain.Consensus                     // Consensus needs
	LastSolidStateIndex uint32                              // State manager mock
	SolidStates         map[uint32]state.VirtualStateAccess // State manager mock
	StateOutput         *isc.AliasOutputWithID              // State manager mock
	Log                 *logger.Logger
	stateSync           coreutil.ChainStateSync // Chain mock
}

func NewNode(env *MockedEnv, nodeIndex uint16, timers ConsensusTimers) *mockedNode { //nolint:revive
	nodeID := env.NodeIDs[nodeIndex]
	log := env.Log.Named(nodeID)
	ret := &mockedNode{
		NodeID:      nodeID,
		NodeIndex:   nodeIndex,
		NodeKeyPair: env.NodeKeyPairs[nodeIndex],
		Env:         env,
		NodeConn:    testchain.NewMockedNodeConnection("Node_"+nodeID, env.Ledgers, log),
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

	originState, err := state.CreateOriginState(store, env.ChainID)
	require.NoError(env.T, err)
	require.Equal(env.T, uint32(0), originState.BlockIndex())
	require.True(env.T, ret.addNewState(originState))

	var peeringID peering.PeeringID
	copy(peeringID[:], env.ChainID[:])
	dss := dss_node.New(&peeringID, env.NetworkProviders[nodeIndex], ret.NodeKeyPair, log)
	cmtN := int(cmt.Size())
	cmtF := cmtN - int(cmt.Quorum())
	registry, err := journal.LoadConsensusJournal(*env.ChainID, cmt.Address(), testchain.NewMockedConsensusJournalRegistry(), cmtN, cmtF, log)
	require.NoError(env.T, err)

	cons := New(ret.ChainCore, ret.Mempool, cmt, cmtPeerGroup, true, metrics.DefaultChainMetrics(), dss, registry, wal.NewDefault(), ret.NodeConn.PublishTransaction, timers)
	cons.(*consensus).vmRunner = testchain.NewMockedVMRunner(env.T, log)
	ret.Consensus = cons

	ret.NodeConn.RegisterChain(
		env.ChainID,
		func(oid iotago.OutputID, o iotago.Output) {
			ret.receiveStateOutput(isc.NewAliasOutputWithID(o.(*iotago.AliasOutput), oid.UTXOInput()))
		},
		func(iotago.OutputID, iotago.Output) {},
		func(milestonePointer *nodeclient.MilestoneInfo) {
			ret.Consensus.SetTimeData(time.Unix(int64(milestonePointer.Timestamp), 0))
		},
	)

	ret.doStateApproved(originState, env.InitStateOutput)

	ret.ChainCore.OnStateCandidate(func(newState state.VirtualStateAccess, approvingOutputID *iotago.UTXOInput) { // State manager mock: state candidate received and is approved by checking that L1 has approving output
		nsCommitment := trie.RootCommitment(newState.TrieNodeStore())
		ret.Log.Debugf("State manager mock (OnStateCandidate): received state candidate: index %v, commitment %v, approving output ID %v, timestamp %v",
			newState.BlockIndex(), nsCommitment, isc.OID(approvingOutputID), newState.Timestamp())

		if !ret.addNewState(newState) {
			return
		}

		go func() {
			var output *iotago.AliasOutput
			getOutputFun := func() *iotago.AliasOutput {
				return env.Ledgers.GetLedger(env.ChainID).GetOutputByID(approvingOutputID)
			}
			for output = getOutputFun(); output == nil; output = getOutputFun() {
				ret.Log.Debugf("State manager mock (OnStateCandidate): transaction index %v has not been published yet", newState.BlockIndex())
				time.Sleep(50 * time.Millisecond)
			}

			ret.Log.Debugf("State manager mock (OnStateCandidate): approving output %v received", isc.OID(approvingOutputID))
			aoCommitment, err := state.L1CommitmentFromAliasOutput(output)
			if err != nil {
				log.Panicf("State manager mock (OnStateCandidate): error retrieving L1 commitment from alias output: %v", err)
			}
			if !state.EqualCommitments(nsCommitment, aoCommitment.StateCommitment) {
				log.Panicf("State manager mock (OnStateCandidate): retrieved L1 commitment %s differs from new state commitment: %s",
					aoCommitment.StateCommitment, nsCommitment)
			}

			if output.StateIndex <= ret.StateOutput.GetStateIndex() {
				ret.Log.Debugf("State manager mock (OnStateCandidate): state output index %v received, but it is too old: current state output is %v",
					output.StateIndex, ret.StateOutput.GetStateIndex())
				return
			}

			ret.doStateApproved(newState, isc.NewAliasOutputWithID(output, approvingOutputID))
		}()
	})
	ret.Log.Debugf("Mocked node %v started: id %v public key %v", ret.NodeIndex, ret.NodeID, ret.NodeKeyPair.GetPublicKey().String())
	return ret
}

func (n *mockedNode) addNewState(newState state.VirtualStateAccess) bool {
	newStateIndex := newState.BlockIndex()
	nsCommitment := trie.RootCommitment(newState.TrieNodeStore())
	oldState, ok := n.SolidStates[newStateIndex]
	if ok {
		osCommitment := trie.RootCommitment(oldState.TrieNodeStore())
		if state.EqualCommitments(osCommitment, nsCommitment) {
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

	if newStateIndex > 0 {
		calcState := n.getState(newStateIndex - 1)
		if calcState != nil {
			block, err := newState.ExtractBlock()
			if err != nil {
				n.Log.Panicf("State manager mock: error extracting block: %v", err)
			}
			calcState = calcState.Copy()
			err = calcState.ApplyBlock(block)
			if err != nil {
				n.Log.Panicf("State manager mock: error applying to previous state: %v", err)
			}
			calcState.Commit()
			csCommitment := trie.RootCommitment(calcState.TrieNodeStore())
			if !state.EqualCommitments(nsCommitment, csCommitment) {
				n.Log.Panicf("State manager mock: calculated state commitment %s differs from new state commitment %s",
					csCommitment, nsCommitment)
			}
		}
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
	n.Log.Debugf("State manager mock: state index %v found, commitment %s", index, trie.RootCommitment(result.TrieNodeStore()))
	return result
}

func (n *mockedNode) getStateFromNodes(index uint32) state.VirtualStateAccess {
	n.Log.Debugf("State manager mock: requesting state index %v", index)
	permutation, err := util.NewPermutation16(uint16(len(n.Env.Nodes) - 1))
	if err != nil {
		n.Log.Panicf("State manager mock: obtaining permutation failed: %v", err)
	}
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

func (n *mockedNode) doStateApproved(newState state.VirtualStateAccess, newStateOutput *isc.AliasOutputWithID) {
	reqIDsForLastState := make([]isc.RequestID, 0)
	prefix := kv.Key(util.Uint32To4Bytes(newState.BlockIndex()))
	err := newState.KVStoreReader().Iterate(prefix, func(key kv.Key, value []byte) bool {
		reqid, err := isc.RequestIDFromBytes(value)
		if err != nil {
			n.Log.Panicf("State manager mock: failed to retrieve request ID from value %v: %v", value, err)
		}
		reqIDsForLastState = append(reqIDsForLastState, reqid)
		return true
	})
	if err != nil {
		n.Log.Panicf("State manager mock: failed to iterate: %v", err)
	}
	n.Mempool.RemoveRequests(reqIDsForLastState...)
	n.Log.Debugf("State manager mock: old requests removed: %v", reqIDsForLastState)

	n.StateOutput = newStateOutput
	n.stateSync.SetSolidIndex(n.StateOutput.GetStateIndex())
	n.Consensus.EnqueueStateTransitionMsg(false, newState, n.StateOutput, time.Now())
	n.Log.Debugf("State manager mock: new state %v approved, commitment %v, state output ID %v",
		n.StateOutput.GetStateIndex(), trie.RootCommitment(newState.TrieNodeStore()), isc.OID(n.StateOutput.ID()))
}

func (n *mockedNode) receiveStateOutput(stateOutput *isc.AliasOutputWithID) { // State manager mock: when node is behind and tries to catchup using state output from L1 and blocks (virtual states in mocke environment) from other nodes
	stateIndex := stateOutput.GetStateIndex()
	if stateOutput != nil && (stateIndex > n.StateOutput.GetStateIndex()) {
		n.Log.Debugf("State manager mock (pullStateLoop): new state output received: index %v, id %v",
			stateIndex, isc.OID(stateOutput.ID()))
		vstate := n.getState(stateIndex)
		if vstate == nil {
			vstate = n.getStateFromNodes(stateIndex)
			if vstate == nil {
				n.Log.Panicf("State manager mock (pullStateLoop): state obtained from nodes is nil")
			}
		}
		n.doStateApproved(vstate, stateOutput)
	}
}

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
