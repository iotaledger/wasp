package statemgr

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/committee"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/plugins/nodeconn"
	"time"
)

func (sm *stateManager) takeAction() {
	if sm.checkStateTransition() {
		return
	}
	sm.requestStateTransactionIfNeeded()
	sm.requestStateUpdateFromPeerIfNeeded()
}

// checks the state of the state manager. If one of pending state update batches is confirmed
// by the nextStateTransaction changes the state to the next
func (sm *stateManager) checkStateTransition() bool {
	if sm.nextStateTransaction == nil {
		return false
	}
	// among pending state updates we locate the one, consistent with the next state transaction
	varStateHash := sm.nextStateTransaction.MustState().VariableStateHash()
	lst, ok := sm.pendingBatches[*varStateHash]
	if !ok {
		// corresponding batch wasn't found among pending state updates
		// state transition is not possible
		return false
	}
	// find pending batch who has the same state tx id
	var pending *pendingBatch
	for _, pb := range lst {
		if pb.batch.StateTransactionId() == sm.nextStateTransaction.ID() {
			pending = pb
			break
		}
	}
	if pending == nil {
		// for some reason nextStateTransaction doesn't approve anything
		return false
	}

	// the nextStateTransaction approves pending batch.
	// Commit the state to the ledger.
	if err := pending.nextVariableState.Commit(sm.committee.Address(), pending.batch); err != nil {
		sm.log.Errorf("failed to save next state #%d", pending.batch.StateIndex())
		return false
	}
	saveTx := sm.nextStateTransaction

	// update state manager variables to the new state
	sm.solidVariableState = pending.nextVariableState
	sm.nextStateTransaction = nil
	sm.pendingBatches = make(map[hashing.HashValue][]*pendingBatch) // clean pending batches
	sm.permutationOfPeers = util.GetPermutation(sm.committee.Size(), varStateHash.Bytes())
	sm.permutationIndex = 0
	sm.syncMessageDeadline = time.Now() // if not synced then immediately

	prevStateIndex := ""
	if sm.solidVariableState.StateIndex() > 0 {
		prevStateIndex = fmt.Sprintf("#%d", sm.solidVariableState.StateIndex()-1)
	}
	sm.log.Infof("STATE TRANSITION %s --> #%d addr %s",
		prevStateIndex, sm.solidVariableState.StateIndex(), sm.committee.Address().String())

	// if synchronized, notify consensus operator about state transition
	if sm.isSynchronized() {
		sm.committee.ReceiveMessage(&committee.StateTransitionMsg{
			VariableState:    sm.solidVariableState,
			StateTransaction: saveTx,
		})
	}
	return true
}

const periodBetweenSyncMessages = 1 * time.Second

func (sm *stateManager) requestStateUpdateFromPeerIfNeeded() {
	if sm.solidVariableState == nil {
		return
	}
	if sm.isSynchronized() {
		// state is synced, no need for more info
		return
	}
	// not synced
	if !sm.syncMessageDeadline.Before(time.Now()) {
		// not time yet for the next message
		return
	}
	// it is time to ask for the next state update to next peer in the permutation
	sm.permutationIndex = (sm.permutationIndex + 1) % sm.committee.Size()
	data := hashing.MustBytes(&committee.GetBatchMsg{
		PeerMsgHeader: committee.PeerMsgHeader{
			StateIndex: sm.solidVariableState.StateIndex() + 1,
		},
	})
	// send messages until first without error
	for i := uint16(0); i < sm.committee.Size(); i++ {
		targetPeerIndex := sm.permutationOfPeers[sm.permutationIndex]
		if err := sm.committee.SendMsg(targetPeerIndex, committee.MsgGetBatch, data); err == nil {
			break
		}
		sm.permutationIndex = (sm.permutationIndex + 1) % sm.committee.Size()
		sm.syncMessageDeadline = time.Now().Add(periodBetweenSyncMessages)
	}
}

const stateTransactionRequestTimeout = 10 * time.Second

func (sm *stateManager) requestStateTransactionIfNeeded() {
	if sm.nextStateTransaction != nil {
		return
	}
	for _, lst := range sm.pendingBatches {
		for _, pb := range lst {
			if pb.stateTransactionRequestDeadline.Before(time.Now()) {
				txid := pb.batch.StateTransactionId()
				sm.log.Debugf("query transaction from the node. txid = %s", txid.String())
				_ = nodeconn.RequestTransactionFromNode(&txid)
				pb.stateTransactionRequestDeadline = time.Now().Add(stateTransactionRequestTimeout)
			}
		}
	}
}

// index of evidenced state index is passed to record the largest one.
// This is needed to check synchronization status. If some state index is more than
// 1 behind the largest, node is not synced
// function returns if the message with idx must be passed to consensus operator, which works only with
// state indices of current or next state
func (sm *stateManager) CheckSynchronizationStatus(idx uint32) bool {
	// synced state is when current state index is behind
	// the largestEvidencedStateIndex no more than by 1 point
	wasSynchronized := sm.isSynchronized()
	if idx > sm.largestEvidencedStateIndex {
		sm.largestEvidencedStateIndex = idx
	}
	if !sm.isSynchronized() && wasSynchronized {
		sm.syncMessageDeadline = time.Now()
	}
	currentStateIndex := uint32(0)
	if sm.solidVariableState != nil {
		currentStateIndex = sm.solidVariableState.StateIndex()
	}
	return idx == currentStateIndex || idx == currentStateIndex+1
}

func (sm *stateManager) isSynchronized() bool {
	if sm.solidVariableState == nil {
		return false
	}
	return sm.largestEvidencedStateIndex-sm.solidVariableState.StateIndex() <= 1
}

// adding batch of state updates to the 'pending' map
func (sm *stateManager) addPendingBatch(batch state.Batch) bool {
	var varState state.VariableState
	var err error

	if sm.solidVariableState != nil {
		if batch.StateIndex() != sm.solidVariableState.StateIndex()+1 {
			return false
		}
	} else {
		// origin
		if batch.StateIndex() != 0 {
			return false
		}
	}
	// clone current variable state of make new empty
	varState = state.NewVariableState(sm.solidVariableState)
	// apply the batch of state updates
	err = varState.Apply(batch)
	if err != nil {
		sm.log.Warn(err)
		return false
	}
	// include the bach to pending batches map
	vh := varState.Hash()
	pendingBatches, ok := sm.pendingBatches[vh]
	if !ok {
		pendingBatches = make([]*pendingBatch, 0)
	}
	pb := &pendingBatch{
		batch:             batch,
		nextVariableState: varState,
	}
	pendingBatches = append(pendingBatches, pb)
	sm.pendingBatches[vh] = pendingBatches

	// request transaction from the node
	txid := batch.StateTransactionId()
	if err := nodeconn.RequestTransactionFromNode(&txid); err != nil {
		sm.log.Debug(err)
	}
	pb.stateTransactionRequestDeadline = time.Now().Add(stateTransactionRequestTimeout)
	return true
}
